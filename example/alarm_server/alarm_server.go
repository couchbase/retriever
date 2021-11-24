/*
Copyright 2014-Present Couchbase, Inc.

Use of this software is governed by the Business Source License included in
the file licenses/BSL-Couchbase.txt.  As of the Change Date specified in that
file, in accordance with the Business Source License, use of this software will
be governed by the Apache License, Version 2.0, included in the file
licenses/APL2.txt.
*/

package main

import (
	"encoding/json"
	"fmt"
	"github.com/couchbase/retriever/logger"
	"net/http"
)

func cmdHandler(w http.ResponseWriter, r *http.Request) {
	var alertMessage = logger.AlarmMessage{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&alertMessage); err != nil {
		http.Error(w, "Unable to decode message", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "OK")

	fmt.Printf("------- Alert Message Received ------- \n")
	fmt.Printf("Module : %s\n", alertMessage.Module)
	fmt.Printf("Transaction Id: %s\n", alertMessage.TraceId)
	fmt.Printf("Key %s \n", alertMessage.Key)
	fmt.Printf("Error Message: %s\n", alertMessage.Message)
	fmt.Printf("------ End Alert Message ------------ \n")

}

func main() {
	http.HandleFunc("/alarm/", cmdHandler)
	http.HandleFunc("/alarm", cmdHandler)
	fmt.Printf("\n----Alarm Event Receiver started ----\n")
	http.ListenAndServe(":9111", nil)
}
