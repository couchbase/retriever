package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/couchbaselabs/retriever/logger"
	"github.com/couchbaselabs/retriever/stats"
	"io/ioutil"
	"net/http"
	"time"
)

var lw *logger.LogWriter
var sc *stats.StatsCollector

const EC = "ExampleClient"

const (
	RESPONSE_OK = iota
	RESPONSE_INVALID_CMD
	RESPONSE_INVALID_MSG
	RESPONSE_UNKNOWN_ERROR
)

type Command struct {
	TransactionId int
	Cmd           int
	Message       string
}

type Response struct {
	TransactionId int
	ResponseCode  int
	Message       string
}

const (
	STAT_REQUESTS      = "requests"
	STAT_FAILURES      = "failures"
	STAT_BYTESTRANS    = "bytesSent"
	STAT_BYTESRECEIVED = "bytesReceived"
)

func do_requests(clientId int, uri string) {

	command := Command{}
	response := Response{}
	i := 0
	client := &http.Client{}
	for {

		i++
		sc.IncrementStat(STAT_REQUESTS)
		command.TransactionId = (clientId * 1000000) + i
		command.Cmd = i % 4
		if i%50 == 0 {
			// 1 out of 50 traces generates an error
			command.Cmd = 5
		}
		command.Message = "Client Id " + string(clientId)
		reqBody, err := json.Marshal(command)
		traceId := fmt.Sprintf("%d", command.TransactionId)
		if err != nil {
			lw.LogError(traceId, EC, "Error marshalling %s", err.Error())
		}
		lw.LogDebug(traceId, EC, "Sending command %d", command.Cmd)

		r, err := http.NewRequest("POST", uri, bytes.NewBufferString(string(reqBody)))
		resp, err := client.Do(r)
		if err != nil {
			lw.LogError(traceId, EC, "Error sending HTTP request %s", err.Error())
			sc.IncrementStat(STAT_FAILURES)
			continue
		}

		sc.UpdateStat(STAT_BYTESTRANS, sc.GetStat(STAT_BYTESTRANS).(int)+len(reqBody))

		respBody, _ := ioutil.ReadAll(resp.Body)
		if err := json.Unmarshal(respBody, &response); err != nil {
			lw.LogError(traceId, EC, "Cannot read response %s", err.Error())
			sc.IncrementStat(STAT_FAILURES)
			continue
		}
		resp.Body.Close()
		if response.ResponseCode != RESPONSE_OK {
			lw.LogError(traceId, EC, "Server returned an error. Code %d", response.ResponseCode)
			sc.IncrementStat(STAT_FAILURES)
		}
		lw.LogDebug(traceId, EC, "Received response from server %s", response.Message)
		sc.UpdateStat(STAT_BYTESRECEIVED, sc.GetStat(STAT_BYTESTRANS).(int)+len(respBody))
		time.Sleep(300 * time.Millisecond)
	}
}

func main() {
	var err error
	lw, err = logger.NewLogger("ExampleClient", logger.LevelInfo)
	if err != nil {
		fmt.Printf("Cannot create logger instance %s", err.Error())
		panic("die")
	}
	lw.EnableKeys([]string{EC})
	lw.SetFile()
	sc, err = stats.NewStatsCollector(EC)
	if err != nil {
		lw.LogError("", EC, "Unable to initialize stats module %s", err.Error())
	}

	sc.AddStatKey(STAT_REQUESTS, 0)
	sc.AddStatKey(STAT_FAILURES, 0)
	sc.AddStatKey(STAT_BYTESTRANS, 0)
	sc.AddStatKey(STAT_BYTESRECEIVED, 0)
	do_requests(1, "http://localhost:9191/command/")
}
