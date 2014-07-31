//  Copyright (c) 2012-2013 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//  http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type message struct {
	Cmd     string
	Message string
}

func getDefaultPath() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("tmp")
	} else {
		return "/tmp"
	}
}

const DEFAULT_PIPE_PATH = `\\.\pipe\`

func HandleLoggerCmds(w http.ResponseWriter, r *http.Request) {
	msg := message{}

	params := mux.Vars(r)
	module := params["module"]

	rl.LogInfo("", LOGGER, "Received logger request for module %s", module)
	// Connect to the module
	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// Send commands to all modules
	if strings.ToLower(module) == "all" {
		var pattern string
		switch msg.Cmd {
		case "log":
			pattern = getDefaultPath() + "/*.log*"
			scanLogs(w, pattern)
		case "traceLog":
			pattern = getDefaultPath() + "/trace_" + msg.Message + ".log"
			scanLogs(w, pattern)
		case "level":
			requestStr := "level:" + msg.Message
			pattern = getDefaultPath() + "/*.sock"
			sendCmdAll(w, requestStr, pattern)
		case "rotate":
			requestStr := "rotate:"
			pattern = getDefaultPath() + "/*.sock"
			sendCmdAll(w, requestStr, pattern)
		case "traceEnable":
			requestStr := "trace:"
			pattern = getDefaultPath() + "/*.sock"
			sendCmdAll(w, requestStr, pattern)
		case "traceDisable":
			requestStr := "traceoff:"
			pattern = getDefaultPath() + "/*.sock"
			sendCmdAll(w, requestStr, pattern)
		case "alarmSet":
			requestStr := "alarm:" + msg.Message
			pattern = getDefaultPath() + "/*.sock"
			sendCmdAll(w, requestStr, pattern)
		case "alarmClear":
			requestStr := "alarmoff:"
			pattern = getDefaultPath() + "/*.sock"
			sendCmdAll(w, requestStr, pattern)
		default:
			http.Error(w, "Invalid Command", http.StatusInternalServerError)
		}

		return
	}

	var requestStr string
	stream := false

	switch msg.Cmd {
	case "level":
		requestStr = "level:" + msg.Message
	case "trace":
		requestStr = "trace:" + msg.Message
	case "traceLog":
		if msg.Message == "" {
			http.Error(w, "Missing trace Id", http.StatusInternalServerError)
			return
		}
		requestStr = getDefaultPath() + "/" + "trace_" + msg.Message + ".log"
		stream = true
	case "log":
		requestStr = getDefaultPath() + "/" + module + ".log"
		stream = true
	case "file":
		requestStr = msg.Message
		stream = true
	case "loglist":
		requestStr = "loglist:"
	case "rotate":
		requestStr = "rotate:"
	case "traceEnable":
		requestStr = "trace:"
	case "traceDisable":
		requestStr = "traceoff:"
	case "alarmSet":
		requestStr = "alarm:" + msg.Message
	case "alarmClear":
		requestStr = "alarmoff:"
	case "path":
		requestStr = "setpath:" + msg.Message
	default:
		http.Error(w, "Invalid Command", http.StatusInternalServerError)
		return

	}

	if stream == false {
		// connect to the module to check if the target process is running
		var module_path string
		if runtime.GOOS == "windows" {
			module_path = DEFAULT_PIPE_PATH + "log_" + module + ".pipe"
		} else {
			module_path = getDefaultPath() + "/" + "log_" + module + ".sock"
		}

		c, err := connect(module_path)

		if err != nil {
			err_msg := "Module " + module + " not found.  Err  " + err.Error()
			http.Error(w, err_msg, http.StatusInternalServerError)
			return
		}
		defer c.Close()

		response := sendCmd(c, w, requestStr)
		io.WriteString(w, response)
	} else {
		//stream data to the user
		streamLog(w, requestStr)
	}

}

func streamLog(w http.ResponseWriter, filePath string) {

	file, err := os.OpenFile(filePath, os.O_RDWR, 0666)

	rl.LogInfo("", LOGGER, "Opening file %s", filePath)
	if err != nil {
		errMsg := "Cannot open file." + "Error: " + err.Error()
		rl.LogWarn("", LOGGER, errMsg)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}
	defer file.Close()

	io.Copy(w, file)
}

func sendCmd(c net.Conn, w http.ResponseWriter, message string) string {

	//set the log level and wait for status
	_, err := c.Write([]byte(message))

	if err != nil {
		errMsg := "Error communicating with module. Reason : " + err.Error()
		rl.LogWarn("", LOGGER, errMsg)
		return errMsg
	}

	buf := make([]byte, 1024)
	n, err := c.Read(buf[:])
	if err != nil {
		errMsg := "Error communicating with module. Reason : " + err.Error()
		rl.LogWarn("", getDefaultPath(), errMsg)
		return errMsg

	}

	// all okay return response to the client
	return string(buf[0:n])
}

// send the command to all the units operating on this server
func sendCmdAll(w http.ResponseWriter, message string, pattern string) {

	fileList, err := filepath.Glob(pattern)
	if err != nil {
		rl.LogWarn("", LOGGER, "No files found for pattern %s", pattern)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(fileList) == 0 {
		fmt.Fprintf(w, "No logs found for pattern %s", pattern)
		return
	}

	fail := 0
	for _, fileName := range fileList {

		if runtime.GOOS == "windows" {
			// extract the module name from the filelist
			basePath := filepath.Base(fileName)
			splits := strings.Split(basePath, ".")
			moduleName := splits[0]
			fileName = DEFAULT_PIPE_PATH + moduleName + ".pipe"
		}

		c, err := connect(fileName)

		if err != nil {
			fmt.Fprintf(w, "%s \n", err.Error())
			fail++
			rl.LogWarn("", LOGGER, err.Error())
			continue
		}
		defer c.Close()

		response := sendCmd(c, w, message)
		fmt.Fprintf(w, "%s %s\n", fileName, response)
	}
	if fail > 0 {
		fmt.Fprintf(w, "Failures %d", fail)
	} else {
		io.WriteString(w, "All OK")
	}

}

func scanLogs(w http.ResponseWriter, pattern string) {

	fileList, err := filepath.Glob(pattern)
	if err != nil {
		rl.LogWarn("", LOGGER, "No files found for pattern %s", pattern)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(fileList) == 0 {
		fmt.Fprintf(w, "No logs found for pattern %s", pattern)
		return
	}

	for _, fileName := range fileList {
		fp, err := os.OpenFile(fileName, os.O_RDWR, 0666)
		if err != nil {
			rl.LogWarn("", LOGGER, err.Error())
			continue
		}
		fmt.Fprintf(w, "\n---- file %s ----- \n", fileName)
		io.Copy(w, fp)
	}
}
