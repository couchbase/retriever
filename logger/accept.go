//  Copyright (c) 2012-2013 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//  http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

// +build !windows

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
	sock := getDefaultPath() + pathSeparator() + "log_" + module + ".sock"
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
		handleCommand(lw, c, cmds, data)
		c.Close()
	}
}
