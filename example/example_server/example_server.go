package main

import (
	"encoding/json"
	"fmt"
	"github.com/couchbaselabs/retriever/logger"
	"github.com/couchbaselabs/retriever/stats"
	"math/rand"
	"net/http"
)

var lw *logger.LogWriter
var sc *stats.StatsCollector

const ES = "ExampleServer"

const (
	RESPONSE_OK = iota
	RESPONSE_INVALID_CMD
	RESPONSE_INVALID_MSG
	RESPONSE_UNKNOWN_ERROR
)

const (
	CMD_HELLO = iota
	CMD_STATS
	CMD_DATA
	CMD_RESTART
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

var answers = []string{
	"It is certain",
	"It is decidedly so",
	"Without a doubt",
	"Yes definitely",
	"You may rely on it",
	"As I see it yes",
	"Most likely",
	"Outlook good",
	"Yes",
	"Signs point to yes",
	"Reply hazy try again",
	"Ask again later",
	"Better not tell you now",
	"Cannot predict now",
	"Concentrate and ask again",
	"Don't count on it",
	"My reply is no",
	"My sources say no",
	"Outlook not so good",
	"Very doubtful",
}

func cmdHandler(w http.ResponseWriter, r *http.Request) {
	command := Command{}
	response := Response{}

	sc.IncrementStat("Requests")
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&command)
	if err != nil {
		lw.LogError("", ES, "Unable to decode message from client")
		sc.IncrementStat("Failures")
		http.Error(w, "Unable to decode message", http.StatusInternalServerError)
		return
	}

	traceId := fmt.Sprintf("%d", command.TransactionId)
	bytesRecvd := sc.GetStat("bytesReceived").(int)
	sc.UpdateStat("bytesReceived", bytesRecvd+int(r.ContentLength))
	lw.LogDebug(traceId, ES, "Received command %d message %s", command.Cmd, command.Message)

	response.TransactionId = command.TransactionId
	response.ResponseCode = RESPONSE_OK

	switch command.Cmd {
	case CMD_HELLO:
		response.Message = "AOK"
	case CMD_STATS:
		response.Message = sc.GetAllStat()
		lw.LogDebug(traceId, ES, "Stats request received")
	case CMD_DATA:
		response.Message = answers[rand.Intn(len(answers))]
	case CMD_RESTART:
		response.Message = "Sorry, No can do "
		lw.LogWarn(traceId, ES, "Unable to restart at this point")
	default:
		sc.IncrementStat("Failures")
		response.ResponseCode = RESPONSE_INVALID_CMD
		lw.LogError(traceId, ES, "Invalid command code %d", command.Cmd)
	}

	lw.LogDebug(traceId, ES, "Response message %d message %s", response.ResponseCode, response.Message)

	respBody, err := json.Marshal(response)
	sc.UpdateStat("bytesSent", sc.GetStat("bytesSent").(int)+len(respBody))
	lw.LogDebug("", ES, "Bytes sent %d", sc.GetStat("bytesSent"))

	fmt.Fprintf(w, string(respBody))

}

func main() {
	http.HandleFunc("/command/", cmdHandler)
	var err error
	lw, err = logger.NewLogger("ExampleServer", logger.LevelInfo)
	if err != nil {
		fmt.Printf("Cannot create logger instance %s", err.Error())
		panic("die")
	}
	lw.EnableKeys([]string{ES})
	lw.SetFile()
	sc, err = stats.NewStatsCollector(ES)
	if err != nil {
		lw.LogError("", ES, "Unable to initialize stats module %s", err.Error())
	}

	sc.AddStatKey("Requests", 0)
	sc.AddStatKey("Success", 0)
	sc.AddStatKey("Failures", 0)
	sc.AddStatKey("Server Port", 9191)
	sc.AddStatKey("bytesReceived", 0)
	sc.AddStatKey("bytesSent", 0)

	lw.LogInfo("", ES, "Example Server starting on port 9191")
	http.ListenAndServe(":9191", nil)
}
