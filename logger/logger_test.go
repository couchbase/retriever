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
	"testing"
	"time"
)

func TestLogger(t *testing.T) {

	mylog, err := NewLogger("test", LevelDebug)
	if err != nil {
		t.Errorf("Failed ! %s", err.Error())
	}
	mylog.EnableKeys([]string{"test1", "test2", "test3", "test4"})
	transId := "0x007"
	mylog.LogDebug(transId, "test1", "hello dolly")
	mylog.LogWarn(transId, "test2", " well hello ")
	mylog.LogInfo(transId, "test5", "no logging this one")
	mylog.LogInfo(transId, "test1", "this is an error ")

	mylog.SetLogLevel(LevelWarn)

	mylog.LogDebug("", "", "not logging this one too ")
	mylog.LogWarn("", "", "this one works ?")

	err = mylog.SetFile("test.log")
	if err != nil {
		t.Errorf("Failed ! Error %s", err.Error())
	}

	mylog.LogError(transId, "test2", "where has this one gone ")
	mylog.DisableKeys([]string{"test3", "test4"})
	mylog.LogError(transId, "test3", " no logging for me ")
	mylog.LogWarn("", "", "file test works !! ")

	mylog.EnableTransactionLogging()

	mylog.LogError(transId, "test2", "this should go to the transaction log")
	transId2 := "0666"
	mylog.LogError(transId, "test1", "this should too ")
	mylog.LogError(transId2, "test2", "goes to new one ")
	mylog.LogError(transId, "test2", "goes to the first id")
	mylog.LogError(transId2, "test1", "new file ")

	mylog.DisableTransactionLogging()
	mylog.LogError(transId, "test1", "goes back to the file")

	for i := 0; i < 5; i++ {
		mylog.RegisterAlarm("http://localhost:9111/alarm/")
		mylog.LogError(transId, "test1", "----Big time error---")
		mylog.LogInfo(transId, "test1", "Info log")
		mylog.LogWarn(transId, "test1", "Warning log")
		mylog.LogDebug(transId, "test1", "Debug log")
		mylog.ClearAlarm()

		time.Sleep(1 * time.Second)
	}

}
