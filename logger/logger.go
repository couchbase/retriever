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
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/couchbaselabs/retriever/lockfile"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
)

type LogLevel int8

func getDefaultPath() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("tmp")
	} else {
		return "/tmp"
	}
}

func pathSeparator() string {
	if runtime.GOOS == "windows" {
		return "/"
	} else {
		return "/"
	}
}

const MAX_CLEANUP_COUNTER = 300
const MAX_LOCK_RETRY = 10

const (
	LevelError = LogLevel(iota)
	LevelWarn
	LevelInfo
	LevelDebug
)

type logDevice int

const (
	Default logDevice = iota << 1 // log to stdout
	File                          // user specified file name
	Remote                        // remote host
)

type TraceLogger struct {
	file     *os.File
	logger   *log.Logger
	counter  uint64
	fileLock lockfile.Lockfile // file lock used for trace logging
}

type AlarmMessage struct {
	Module  string
	TraceId string
	Key     string
	Message string
}

type LogWriter struct {
	module         string                 // name of logging module
	level          LogLevel               // current log leve
	keyList        map[string]bool        // list of enabled keys
	mu             sync.Mutex             // mutex for this structure
	logger         *log.Logger            // instance of logger module
	filePath       string                 // path of log file for this module
	traceFileMap   map[string]interface{} // table of trace logs - trace id
	traceMu        sync.RWMutex           // R/W mutex to sync access to the above structure
	traceMode      bool                   // trace mode enabled
	cleanerRunning bool                   // trace log cleaner process
	logCounter     uint64                 // count of log messages
	file           *os.File               // file handle of log file
	alarmEnabled   bool                   // endpoint alarms enabled
	alarmLogger    AlarmLogger            // instance of alarm logger
	defaultPath    string                 // default logging path
	color          bool                   // enable/disable colour logging
}

type AlarmLogger struct {
	endpoint string            // address of alarm endpoint
	cMsg     chan AlarmMessage // channel used to communicate messages to remote server
	cStop    chan bool         // stop channel
}

// Create a new instance of a logWriter
func NewLogger(module string, level LogLevel) (*LogWriter, error) {

	if module == "" {
		return nil, fmt.Errorf("Required module name")
	}

	// set loglevel to default
	if level > LevelDebug || level < LevelError {
		level = LevelWarn
	}

	var lw *LogWriter

	if runtime.GOOS == "windows" {
		// disable color logging on windows
		lw = &LogWriter{module: module,
			level:        level,
			keyList:      make(map[string]bool),
			logger:       log.New(os.Stderr, "", log.Lmicroseconds),
			traceFileMap: make(map[string]interface{}),
			color:        false,
		}
	} else {
		lw = &LogWriter{module: module,
			level:        level,
			keyList:      make(map[string]bool),
			logger:       log.New(os.Stderr, "", log.Lmicroseconds),
			traceFileMap: make(map[string]interface{}),
			color:        true,
		}
	}

	lw.keyList["Default"] = true

	go handleConnections(lw, module)
	return lw, nil
}

type Message struct {
	Cmd     string
	Message string
}

// Set the log level
func (lw *LogWriter) SetLogLevel(level LogLevel) error {

	if lw.level > LevelDebug || lw.level < LevelError {
		return fmt.Errorf("Log level unchanged")
	}
	lw.level = level
	return nil
}

// Set the output device. Use module Id for name
func (lw *LogWriter) SetFile() error {

	if lw.defaultPath == "" {
		lw.filePath = getDefaultPath() + pathSeparator() + lw.module + ".log"
	} else {
		lw.filePath = lw.defaultPath + pathSeparator() + lw.module + ".log"
	}

	fp, err := os.OpenFile(lw.filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("Unable to open file %s", err.Error())
	}

	lw.file = fp
	lw.logger = log.New(fp, "", log.Lmicroseconds)
	return nil
}

// set the default logging path. If trace logging is not enabled then only new trace
// files will use the new default path.

func (lw *LogWriter) SetDefaultPath(defaultPath string) error {

	if len(defaultPath) == 0 {
		return fmt.Errorf("No path specified")
	}

	if lw.defaultPath == defaultPath {
		return nil
	}

	newPath := defaultPath + "/" + lw.module + ".log"
	fp, err := os.OpenFile(newPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("Unable access path %s", err.Error())
	}
	lw.defaultPath = defaultPath

	if lw.file != nil {
		// switch the log file
		lw.file.Close()
		lw.file = fp
		lw.filePath = newPath
		lw.logger = log.New(fp, "", log.Lmicroseconds)
	}

	return nil
}

// Rotate the current log file
func (lw *LogWriter) Rotate() error {
	renamePath := fmt.Sprintf("%s.%s", lw.filePath, time.Now().String())
	err := os.Rename(lw.filePath, renamePath)
	if err != nil {
		return err
	}

	fp, err := os.OpenFile(lw.filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("Unable to open file %s", err.Error())
	}

	lw.logger = log.New(fp, "", log.Lmicroseconds)
	lw.file.Close()
	lw.file = fp

	return nil
}

// Set the logging to the log to a trace file
func (lw *LogWriter) EnableTraceLogging() {
	lw.traceMode = true
	if lw.cleanerRunning == false {
		go cleanupMap(lw)
		lw.cleanerRunning = true
	}
}

// Disable logging to a trace file
func (lw *LogWriter) DisableTraceLogging() {
	lw.traceMode = false
}

// Set the remote host
func (lw *LogWriter) SetLogHost(string) error {
	return nil
}

//enable component keys
func (lw *LogWriter) EnableKeys(keys []string) error {
	lw.mu.Lock()
	defer lw.mu.Unlock()
	for _, key := range keys {
		lw.keyList[key] = true
	}
	return nil
}

// Enable/Disable color logging
func (lw *LogWriter) SetColor(value bool) {
	lw.color = value
}

// disable component keys
func (lw *LogWriter) DisableKeys(keys []string) error {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	for _, key := range keys {
		delete(lw.keyList, key)
	}
	return nil
}

// Check to see if logging is enabled for a key
func (lw *LogWriter) keyEnabled(key string) bool {
	_, found := lw.keyList[key]
	return found
}

func (lw *LogWriter) logTrace(traceId string, logString string) bool {
	var logger *log.Logger
	lw.traceMu.RLock()
	tl := lw.traceFileMap[traceId]
	lw.traceMu.RUnlock()
	if tl == nil {
		filePath := getDefaultPath() + pathSeparator() + "trace_" + traceId + ".log"
		file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			lw.logger.Print("Logger: Unable to create trace file %s, Error %s", filePath, err.Error())
			return false
		}
		logger = log.New(file, "", log.Lmicroseconds)
		fileLockPath := getDefaultPath() + pathSeparator() + "trace_" + traceId + ".lock"
		fl, _ := lockfile.New(fileLockPath)
		tl = &TraceLogger{file: file, logger: logger, counter: lw.logCounter, fileLock: fl}
		lw.traceMu.Lock()
		lw.traceFileMap[traceId] = tl
		lw.traceMu.Unlock()
		if lw.cleanerRunning == false {
			// restart the cleaner
			go cleanupMap(lw)
		}
	} else {
		logger = tl.(*TraceLogger).logger
		tl.(*TraceLogger).counter = lw.logCounter
	}

	var locked bool
	err := tl.(*TraceLogger).fileLock.TryLock()
	if err != nil {
		// busy wait a few times before giving up
		for i := 0; i < MAX_LOCK_RETRY; i++ {
			time.Sleep(10 * time.Millisecond)
			err = tl.(*TraceLogger).fileLock.TryLock()
			if err == nil {
				locked = true
				break
			}
		}
		// unable to acquire lock
		if locked == false {
			return false
		}
	}
	defer tl.(*TraceLogger).fileLock.Unlock()

	logger.Print(logString)
	return true
}

//cleanup trace filemap
func cleanupMap(lw *LogWriter) {
	defer func() {
		if r := recover(); r != nil {
			lw.logger.Print("Logger: Crash in cleanupMap")
			lw.cleanerRunning = false
		}
	}()

	for {
		if len(lw.traceFileMap) > 0 {
			for key, _ := range lw.traceFileMap {
				tl := lw.traceFileMap[key].(*TraceLogger)
				if lw.logCounter-tl.counter >= MAX_CLEANUP_COUNTER {
					fmt.Println("Closing File ", tl.file.Name())
					tl.file.Close()
					lw.traceMu.Lock()
					delete(lw.traceFileMap, key)
					lw.traceMu.Unlock()
				}
			}
		} else {
			if lw.traceMode == false && len(lw.traceFileMap) == 0 {
				lw.cleanerRunning = false
				return
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func (lw *LogWriter) logMessage(color string, traceId string, key string, format string, args ...interface{}) {
	var logString string
	lw.logCounter++

	if lw.color == false {
		color = reset
	}

	// color formatting doesn't work on windows.
	if runtime.GOOS == "windows" {
		if traceId != "" {
			logString = fmt.Sprintf("%s %s %s", key, traceId, fmt.Sprintf(format, args...))
		} else {
			logString = fmt.Sprintf("%s None %s", key, fmt.Sprintf(format, args...))
		}
	} else {
		if traceId != "" {
			logString = fmt.Sprintf("%s %s %s", color+key, reset+traceId, fmt.Sprintf(format, args...))
		} else {
			logString = fmt.Sprintf("%s None %s", color+key, reset+fmt.Sprintf(format, args...))
		}
	}

	if lw.traceMode == true && len(traceId) > 0 {
		if lw.logTrace(traceId, logString) {
			return
		}
	}

	if runtime.GOOS == "windows" {
		lw.logger.Print(logString)
	} else {
		lw.logger.Print(color, logString)
	}
}

// log debug. trace id, component id, log message
func (lw *LogWriter) LogDebug(traceId string, key string, format string, args ...interface{}) {
	if lw.level >= LevelDebug {
		if key == "" {
			key = "Default"
		}
		if lw.keyEnabled(key) {
			lw.logMessage(fgWhite, traceId, key, format, args...)
		}
	}
}

//log info. trace id, component id, log message
func (lw *LogWriter) LogInfo(traceId string, key string, format string, args ...interface{}) {
	if lw.level >= LevelInfo {
		if key == "" {
			key = "Default"
		}
		if lw.keyEnabled(key) {
			lw.logMessage(fgBlue, traceId, key, format, args...)
		}
	}
}

//log warning trace id, component id, log message
func (lw *LogWriter) LogWarn(traceId string, key string, format string, args ...interface{}) {
	if lw.level >= LevelWarn {
		if key == "" {
			key = "Default"
		}
		if lw.keyEnabled(key) {
			lw.logMessage(fgYellow, traceId, key, format, args...)
		}
	}
}

//log error trace id, component id, log message
func (lw *LogWriter) LogError(traceId string, key string, format string, args ...interface{}) {
	if lw.level >= LevelError {
		if key == "" {
			key = "Default"
		}
		if lw.keyEnabled(key) {
			lw.logMessage(fgRed, traceId, key, format, args...)
		}
		if lw.alarmEnabled == true {
			// send alarm to remote host
			message := fmt.Sprintf(format, args...)
			lw.alarmLogger.cMsg <- AlarmMessage{Module: lw.module, Key: key, TraceId: traceId, Message: message}
		}
	}
}

// register alarm endpoint. Any error log will be sent to this remote endpoint
func (lw *LogWriter) RegisterAlarm(endpoint string) error {

	if lw.alarmEnabled == false {
		lw.alarmLogger = AlarmLogger{endpoint: endpoint, cMsg: make(chan AlarmMessage), cStop: make(chan bool)}
		lw.alarmEnabled = true
		go sendAlarm(endpoint, lw.alarmLogger.cMsg, lw.alarmLogger.cStop)
	}
	return nil
}

func (lw *LogWriter) ClearAlarm() {
	lw.alarmEnabled = false
	lw.alarmLogger.cStop <- true
}

func sendAlarm(endpoint string, cMsg chan AlarmMessage, cStop chan bool) {

	client := &http.Client{}
	ok := true
	for ok {
		select {
		case msg := <-cMsg:
			reqBody, _ := json.Marshal(msg)
			r, _ := http.NewRequest("POST", endpoint, bytes.NewBufferString(string(reqBody)))
			resp, err := client.Do(r)
			if err != nil {
				log.Print(fgWhite, "Logger Error sending request to endpoint %s", err.Error())
				continue
			}
			ioutil.ReadAll(resp.Body)
			resp.Body.Close()
		case <-cStop:
			ok = false
		}
	}
}

// ANSI color control escape sequences.
// Shamelessly copied from https://github.com/sqp/godock/blob/master/libs/log/colors.go
var (
	reset      = "\x1b[0m"
	bright     = "\x1b[1m"
	dim        = "\x1b[2m"
	underscore = "\x1b[4m"
	blink      = "\x1b[5m"
	reverse    = "\x1b[7m"
	hidden     = "\x1b[8m"
	fgBlack    = "\x1b[30m"
	fgRed      = "\x1b[31m"
	fgGreen    = "\x1b[32m"
	fgYellow   = "\x1b[33m"
	fgBlue     = "\x1b[34m"
	fgMagenta  = "\x1b[35m"
	fgCyan     = "\x1b[36m"
	fgWhite    = "\x1b[37m"
	bgBlack    = "\x1b[40m"
	bgRed      = "\x1b[41m"
	bgGreen    = "\x1b[42m"
	bgYellow   = "\x1b[43m"
	bgBlue     = "\x1b[44m"
	bgMagenta  = "\x1b[45m"
	bgCyan     = "\x1b[46m"
	bgWhite    = "\x1b[47m"
)
