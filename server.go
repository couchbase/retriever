//  Copyright 2012-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included
//  in the file licenses/BSL-Couchbase.txt.  As of the Change Date specified
//  in that file, in accordance with the Business Source License, use of this
//  software will be governed by the Apache License, Version 2.0, included in
//  the file licenses/APL2.txt.

package main

import (
	"fmt"
	"github.com/couchbase/retriever/logger"
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
