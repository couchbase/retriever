//  Copyright 2012-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included
//  in the file licenses/BSL-Couchbase.txt.  As of the Change Date specified
//  in that file, in accordance with the Business Source License, use of this
//  software will be governed by the Apache License, Version 2.0, included in
//  the file licenses/APL2.txt.

package logger

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func handleCommand(lw *LogWriter, c net.Conn, cmds []string, data string) {

	var err error

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
