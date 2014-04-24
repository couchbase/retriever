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
	"github.com/couchbaselabs/retriever/logger"
	"github.com/gorilla/mux"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var rl logger.LogWriter

const DEFAULT_PATH = "/tmp"
const DEFAULT_MODULE = "Retriever"

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/logger/{module}", HandleLoggerCmds).Methods("PUT", "POST")
	http.Handle("/", r)

	rl, err := logger.NewLogger(DEFAULT_MODULE, logger.LevelInfo)
	if err != nil {
		panic_msg := fmt.Sprintf("Cannot intialize logger %s", err.Error())
		panic(panic_msg)
	}
	rl.SetFile("Retriever.log")
	rl.EnableKeys([]string{DEFAULT_MODULE, "Logger", "Stats"})

	rl.LogInfo("", DEFAULT_MODULE, "Retriever Server started")

	http.ListenAndServe(":8080", nil)

}

type Message struct {
	Cmd     string
	Message string
}

func HandleLoggerCmds(w http.ResponseWriter, r *http.Request) {
	msg := Message{}
	//err := r.DecodeJsonPayload(&msg)

	params := mux.Vars(r)
	module := params["module"]

	rl.LogInfo("", DEFAULT_MODULE, "Received logger request for module %s", module)
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
		case "Log":
			pattern = DEFAULT_PATH + "/*.log.*"
			scanLogs(w, pattern)
		case "TransactionLog":
			pattern = DEFAULT_PATH + "/trans_*.log"
			scanLogs(w, pattern)
		case "Level":
			requestStr := "Level:" + msg.Message
			pattern = DEFAULT_PATH + "/*.sock"
			sendCmdAll(w, requestStr, pattern)
		case "Rotate":
			requestStr := "Rotate:"
			pattern = DEFAULT_PATH + "/*.sock"
			sendCmdAll(w, requestStr, pattern)

		}

		return
	}

	var requestStr string
	stream := false

	switch msg.Cmd {
	case "Level":
		requestStr = "Level:" + msg.Message
	case "Transaction":
		requestStr = "Transaction:" + msg.Message
	case "TransactionLog":
		if msg.Message == "" {
			http.Error(w, "Missing transaction Id", http.StatusInternalServerError)
			return
		}
		requestStr = "trans_" + msg.Message + ".log"
		stream = true
	case "Log":
		requestStr = module + ".log"
		stream = true
	case "Loglist":
		requestStr = "Loglist:"
	case "Rotate":
		requestStr = "Rotate:"
	}

	if stream == false {
		// connect to the module to check if the target process is running
		module_path := "/tmp/log_" + module + ".sock"
		c, err := net.Dial("unix", module_path)

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

	filePath = DEFAULT_PATH + "/" + filePath
	file, err := os.OpenFile(filePath, os.O_RDWR, 0666)

	rl.LogInfo("", DEFAULT_MODULE, "Opening file %s", filePath)
	if err != nil {
		errMsg := "Cannot open file." + "Error: " + err.Error()
		rl.LogWarn("", DEFAULT_MODULE, errMsg)
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
		rl.LogWarn("", DEFAULT_MODULE, errMsg)
		return errMsg
	}

	buf := make([]byte, 1024)
	n, err := c.Read(buf[:])
	if err != nil {
		errMsg := "Error communicating with module. Reason : " + err.Error()
		rl.LogWarn("", DEFAULT_PATH, errMsg)
		return errMsg

	}

	// all okay return response to the client
	return string(buf[0:n])

}

// send the command to all the units operating on this server
func sendCmdAll(w http.ResponseWriter, message string, pattern string) {

	fileList, err := filepath.Glob(pattern)
	if err != nil {
		rl.LogWarn("", DEFAULT_MODULE, "No files found for pattern %s", pattern)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(fileList) == 0 {
		fmt.Fprintf(w, "No logs found for pattern %s", pattern)
		return
	}

	fail := 0
	for _, fileName := range fileList {
		c, err := net.Dial("unix", fileName)

		if err != nil {
			fmt.Fprintf(w, "%s \n", err.Error())
			fail++
			rl.LogWarn("", DEFAULT_MODULE, err.Error())
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
		rl.LogWarn("", DEFAULT_MODULE, "No files found for pattern %s", pattern)
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
			rl.LogWarn("", DEFAULT_MODULE, err.Error())
			continue
		}
		fmt.Fprintf(w, "\n---- file %s ----- \n", fileName)
		io.Copy(w, fp)
	}
}
