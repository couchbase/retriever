//  Copyright 2012-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included
//  in the file licenses/BSL-Couchbase.txt.  As of the Change Date specified
//  in that file, in accordance with the Business Source License, use of this
//  software will be governed by the Apache License, Version 2.0, included in
//  the file licenses/APL2.txt.

// +build !windows

package stats

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func handleConnections(sc *StatsCollector) {

	// create an I/O channel based on the module name
	// for the server to connect to
	sock := "/tmp/" + "stats_" + sc.Module + ".sock"
	os.Remove(sock)
	listener, err := net.Listen("unix", sock)

	if err != nil {
		fmt.Printf("Failed to listen ", err.Error())
	}

	defer os.Remove(sock)
	defer listener.Close()

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
