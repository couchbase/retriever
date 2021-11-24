//  Copyright 2012-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included
//  in the file licenses/BSL-Couchbase.txt.  As of the Change Date specified
//  in that file, in accordance with the Business Source License, use of this
//  software will be governed by the Apache License, Version 2.0, included in
//  the file licenses/APL2.txt.

package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"runtime"
	"strings"
)

func HandleStatsCmds(w http.ResponseWriter, r *http.Request) {
	msg := message{}
	//err := r.DecodeJsonPayload(&msg)

	params := mux.Vars(r)
	module := params["module"]

	rl.LogInfo("", LOGGER, "Received stats request for module %s", module)
	// Connect to the module
	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// Send commands to all modules
	requestStr := "stats:"
	if strings.ToLower(module) == "all" {
		pattern := getDefaultPath() + "/stats_*.sock"
		sendCmdAll(w, requestStr, pattern)
		return
	}

	// connect to the module to check if the target process is running
	var module_path string
	if runtime.GOOS == "windows" {
		module_path = DEFAULT_PIPE_PATH + "stats_" + module + ".pipe"
	} else {
		module_path = getDefaultPath() + "/stats_" + module + ".sock"
	}

	c, err := connect(module_path)

	if err != nil {
		err_msg := "Module " + module + " not found.  Err  " + err.Error()
		http.Error(w, err_msg, http.StatusInternalServerError)
		return
	}
	defer c.Close()

	response := sendCmd(c, w, requestStr)
	io.WriteString(w, response)

}
