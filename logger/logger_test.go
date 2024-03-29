//  Copyright 2012-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included
//  in the file licenses/BSL-Couchbase.txt.  As of the Change Date specified
//  in that file, in accordance with the Business Source License, use of this
//  software will be governed by the Apache License, Version 2.0, included in
//  the file licenses/APL2.txt.

package logger

import (
	"runtime"
	"testing"
	"time"
)

func TestLogger(t *testing.T) {

	mylog, err := NewLogger("test", LevelDebug)
	if err != nil {
		t.Errorf("Failed ! %s", err.Error())
	}
	mylog.EnableKeys([]string{"test1", "test2", "test3", "test4"})
	traceId := "0x007"
	mylog.LogDebug(traceId, "test1", "hello dolly")
	mylog.LogWarn(traceId, "test2", " well hello ")
	mylog.LogInfo(traceId, "test5", "no logging this one")
	mylog.LogInfo(traceId, "test1", "this is an error ")

	mylog.SetLogLevel(LevelWarn)

	mylog.LogDebug("", "", "not logging this one too ")
	mylog.LogWarn("", "", "this one works ?")

	err = mylog.SetFile()
	if err != nil {
		t.Errorf("Failed ! Error %s", err.Error())
	}

	mylog.LogError(traceId, "test2", "where has this one gone ")
	mylog.DisableKeys([]string{"test3", "test4"})
	mylog.LogError(traceId, "test3", " no logging for me ")
	mylog.LogWarn("", "", "file test works !! ")

	mylog.EnableTraceLogging()

	mylog.LogError(traceId, "test2", "this should go to the traceaction log")
	traceId2 := "0666"
	mylog.LogError(traceId, "test1", "this should too ")
	mylog.LogError(traceId2, "test2", "goes to new one ")
	mylog.LogError(traceId, "test2", "goes to the first id")
	mylog.LogError(traceId2, "test1", "new file ")

	mylog.DisableTraceLogging()
	mylog.LogError(traceId, "test1", "goes back to the file")

	mylog.SetLogLevel(LevelDebug)

	for i := 0; i < 5; i++ {
		if i == 3 {
			if runtime.GOOS != "windows" {
				if err = mylog.SetDefaultPath("/tmp"); err != nil {
					t.Errorf("Failed ! Error %s", err.Error())
				}

			}
		}
		mylog.RegisterAlarm("http://localhost:9111/alarm/")
		mylog.LogError(traceId, "test1", "----Big time error---")
		mylog.LogInfo(traceId, "test1", "Info log")
		mylog.LogWarn(traceId, "test1", "Warning log")
		mylog.LogDebug(traceId, "test1", "Debug log")
		mylog.SetColor(false)
		mylog.ClearAlarm()

		time.Sleep(1 * time.Second)
	}

}
