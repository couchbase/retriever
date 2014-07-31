//  Copyright (c) 2012-2013 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//  http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package stats

import (
	"fmt"
	"github.com/natefinch/npipe"
	"log"
	"os"
	"strings"
)

const DEFAULT_PIPE_PATH = `\\.\pipe\`

func handleConnections(sc *StatsCollector) {

	// create an I/O channel based on the module name
	// for the server to connect to
	pipe := DEFAULT_PIPE_PATH + "stats_" + sc.Module + ".pipe"
	os.Remove(pipe)
	listener, err := npipe.Listen(pipe)

	if err != nil {
		fmt.Printf("Failed to listen ", err.Error())
	}

	defer os.Remove(pipe)
	defer listener.Close()

	// create a file entry for the pipe in the default pathname so that clients
	// can discover the pipe entry
	pipeEntry := getDefaultPath() + "/stats_" + sc.Module + ".sock"
	os.Remove(pipeEntry)

	fp, err := os.OpenFile(pipeEntry, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Unable to open file %s", err.Error())
	}

	defer fp.Close()
	defer os.Remove(pipeEntry)

	for {
		c, err := listener.Accept()
		if err != nil {
			fmt.Printf("Unable to accept " + err.Error())
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
		cmds := strings.Split(data, ":")
		switch {
		case strings.Contains(strings.ToLower(cmds[0]), "stats"):
			// rotate the current log file
			statsOutput := sc.GetAllStat()
			c.Write([]byte(statsOutput))
		}
		c.Close()
	}
}
