//  Copyright (c) 2012-2013 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//  http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package logger

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func handleConnections(lw *LogWriter, module string) {
	for {
		doHandleConnections(lw, module)
	}
}

func doHandleConnections(lw *LogWriter, module string) {

	// create an I/O channel based on the module name
	// for the server to connect to
	sock := DEFAULT_PATH + "/log_" + module + ".sock"
	os.Remove(sock)
	listener, err := net.Listen("unix", sock)

	if err != nil {
		log.Fatal("Failed to listen ", err.Error())
	}

	defer os.Remove(sock)
	defer listener.Close()

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()

	for {
		c, err := listener.Accept()
		if err != nil {
			fmt.Printf("Unable to accept " + err.Error()) // FIXME
			continue
		}
		buf := make([]byte, 512)
		nr, err := c.Read(buf)
		if err != nil {
			fmt.Printf(" Could not read from buffer %s", err.Error())
			c.Close()
			continue
		}
		data := string(buf[0:nr])
		cmds := strings.SplitN(data, ":", 2)
		switch {
		case strings.Contains(strings.ToLower(cmds[0]), "level"):
			setLevel(lw, cmds[1])
			c.Write([]byte("OK"))
		case strings.Contains(strings.ToLower(cmds[0]), "filelog"):
			sendfile(lw.filePath, c)
		case strings.Contains(strings.ToLower(cmds[0]), "rotate"):
			// rotate the current log file
			err := lw.Rotate()
			if err != nil {
				c.Write([]byte(err.Error()))
			} else {
				c.Write([]byte("OK"))
			}
		case strings.Contains(strings.ToLower(cmds[0]), "traceoff"):
			lw.DisableTraceLogging()
			c.Write([]byte("OK"))

		case strings.Contains(strings.ToLower(cmds[0]), "trace"):
			lw.EnableTraceLogging()
			c.Write([]byte("OK"))
		case strings.Contains(strings.ToLower(cmds[0]), "alarmoff"):
			lw.ClearAlarm()
			c.Write([]byte("OK"))
		case strings.Contains(strings.ToLower(cmds[0]), "alarm"):
			lw.RegisterAlarm(cmds[1])
			c.Write([]byte("OK"))
		case strings.Contains(strings.ToLower(cmds[0]), "setpath"):
			if err = lw.SetDefaultPath(cmds[1]); err != nil {
				c.Write([]byte(err.Error()))
			}
			c.Write([]byte("OK"))

		}
		c.Close()
	}
}

func setLevel(lw *LogWriter, level string) {

	level = strings.ToLower(level)
	switch {
	case strings.Contains(level, "info"):
		fmt.Printf("Setting to level info ")
		lw.SetLogLevel(LevelInfo)
	case strings.Contains(level, "warn"):
		lw.SetLogLevel(LevelWarn)
	case strings.Contains(level, "error"):
		lw.SetLogLevel(LevelError)
	case strings.Contains(level, "debug"):
		lw.SetLogLevel(LevelDebug)
	}

}

func sendfile(filePath string, c net.Conn) {

	file, err := os.OpenFile(filePath, os.O_RDWR, 0666)

	if err != nil {
		c.Write([]byte("Cannot open file " + filePath + "Error:" + err.Error()))
		return
	}
	defer file.Close()
	buffer := make([]byte, 8192)
	var bytesWr uint64

	for {
		n, err := file.Read(buffer)
		if n > 0 {
			bytes, sockerr := c.Write(buffer[0:n])
			if sockerr != nil {
				fmt.Printf("Error writing %s", sockerr.Error())
			}
			bytesWr += uint64(bytes)

		}
		if err != nil {
			if bytesWr == 0 {
				fmt.Printf("I/O error")
			}
			break
		}
	}
}
