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
	"fmt"
	"github.com/couchbaselabs/retriever/logger"
	"github.com/gorilla/mux"
	"net/http"
)

var rl logger.LogWriter

const DEFAULT = "Retriever"
const LOGGER = "Logger"

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/logger/{module}", HandleLoggerCmds).Methods("GET", "PUT", "POST")
	r.HandleFunc("/stats/{module}", HandleStatsCmds).Methods("GET", "PUT", "POST")
	http.Handle("/", r)

	rl, err := logger.NewLogger(DEFAULT, logger.LevelDebug)
	if err != nil {
		panic_msg := fmt.Sprintf("Cannot intialize logger %s", err.Error())
		panic(panic_msg)
	}
	rl.SetFile()
	rl.EnableKeys([]string{DEFAULT, LOGGER, "Stats"})

	rl.LogInfo("", DEFAULT, "Retriever Server started")

	http.ListenAndServe(":8080", nil)
}
