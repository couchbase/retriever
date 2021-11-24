//  Copyright 2012-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included
//  in the file licenses/BSL-Couchbase.txt.  As of the Change Date specified
//  in that file, in accordance with the Business Source License, use of this
//  software will be governed by the Apache License, Version 2.0, included in
//  the file licenses/APL2.txt.

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
