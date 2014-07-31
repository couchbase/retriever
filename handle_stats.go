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
