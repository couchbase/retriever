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
	"fmt"
	"github.com/couchbaselabs/retriever/lockfile"
	"log"
	"os"
	"sync"
	"time"
)

type logLevel int

const DEFAULT_PATH = "/tmp"
const MAX_CLEANUP_COUNTER = 1000
const MAX_LOCK_RETRY = 10

const (
	LevelError = logLevel(iota)
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

type TransactionLogger struct {
	file     *os.File
	logger   *log.Logger
	counter  uint64
	fileLock lockfile.Lockfile // file lock used for transaction logging
}

type LogWriter struct {
	module         string                 // name of logging module
	level          logLevel               // current log leve
	keyList        map[string]bool        // list of enabled keys
	mu             sync.Mutex             // mutex for this structure
	logger         *log.Logger            // instance of logger module
	filePath       string                 // path of log file for this module
	transFileMap   map[string]interface{} // table of transaction logs - transaction id
	transMu        sync.RWMutex           // R/W mutex to sync access to the above structure
	transMode      bool                   // transaction mode enabled
	cleanerRunning bool                   // transaction log cleaner process
	logCounter     uint64                 // count of log messages
	file           *os.File               // file handle of log file
}

// Create a new instance of a logWriter
func NewLogger(module string, level logLevel) (*LogWriter, error) {

	if module == "" {
		return nil, fmt.Errorf("Required module name")
	}

	// set loglevel to default
	if level > LevelDebug || level < LevelError {
		level = LevelWarn
	}

	lw := &LogWriter{module: module,
		level:        level,
		keyList:      make(map[string]bool),
		logger:       log.New(os.Stderr, "", log.Lmicroseconds),
		transFileMap: make(map[string]interface{}),
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
func (lw *LogWriter) SetLogLevel(level logLevel) error {

	if lw.level > LevelDebug || lw.level < LevelError {
		return fmt.Errorf("Log level unchanged")
	}
	lw.level = level
	return nil
}

// Set the output device
func (lw *LogWriter) SetFile(path string) error {
	lw.filePath = DEFAULT_PATH + "/" + path
	fp, err := os.OpenFile(lw.filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("Unable to open file %s", err.Error())
	}
	//lw.w = bufio.NewWriter(fp)
	lw.file = fp
	lw.logger = log.New(fp, "", log.Lmicroseconds)
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

// Set the logging to the log to a transaction file
func (lw *LogWriter) EnableTransactionLogging() {
	lw.transMode = true
	if lw.cleanerRunning == false {
		go cleanupMap(lw)
		lw.cleanerRunning = true
	}
}

// Disable logging to a transaction file
func (lw *LogWriter) DisableTransactionLogging() {
	lw.transMode = false
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

func (lw *LogWriter) logTransaction(transactionId string, logString string) bool {
	var logger *log.Logger
	lw.transMu.RLock()
	tl := lw.transFileMap[transactionId]
	lw.transMu.RUnlock()
	if tl == nil {
		filePath := DEFAULT_PATH + "/" + "trans_" + transactionId + ".log"
		file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			lw.logger.Print("Logger: Unable to create transaction file %s, Error %s", filePath, err.Error())
			return false
		}
		logger = log.New(file, "", log.Lmicroseconds)
		fileLockPath := DEFAULT_PATH + "/" + "trans_" + transactionId + ".lock"
		fl, _ := lockfile.New(fileLockPath)
		tl = &TransactionLogger{file: file, logger: logger, counter: lw.logCounter, fileLock: fl}
		lw.transMu.Lock()
		lw.transFileMap[transactionId] = tl
		lw.transMu.Unlock()
		if lw.cleanerRunning == false {
			// restart the cleaner
			go cleanupMap(lw)
		}
	} else {
		logger = tl.(*TransactionLogger).logger
		tl.(*TransactionLogger).counter = lw.logCounter
	}

	var locked bool
	err := tl.(*TransactionLogger).fileLock.TryLock()
	if err != nil {
		// busy wait a few times before giving up
		for i := 0; i < MAX_LOCK_RETRY; i++ {
			time.Sleep(10 * time.Millisecond)
			err = tl.(*TransactionLogger).fileLock.TryLock()
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
	defer tl.(*TransactionLogger).fileLock.Unlock()

	logger.Print(logString)
	return true
}

//cleanup transaction filemap
func cleanupMap(lw *LogWriter) {
	defer func() {
		if r := recover(); r != nil {
			lw.logger.Print("Logger: Crash in cleanupMap")
			lw.cleanerRunning = false
		}
	}()

	for {
		if len(lw.transFileMap) > 0 {
			for key, _ := range lw.transFileMap {
				tl := lw.transFileMap[key].(*TransactionLogger)
				if lw.logCounter-tl.counter >= MAX_CLEANUP_COUNTER {
					fmt.Println("Closing File ", tl.file.Name())
					tl.file.Close()
					lw.transMu.Lock()
					delete(lw.transFileMap, key)
					lw.transMu.Unlock()
				}
			}
		} else {
			if lw.transMode == false && len(lw.transFileMap) == 0 {
				lw.cleanerRunning = false
				return
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func (lw *LogWriter) logMessage(transactionId string, key string, format string, args ...interface{}) {
	var logString string
	lw.logCounter++
	if transactionId != "" {
		logString = fmt.Sprintf("%s %s %s", key, transactionId, fmt.Sprintf(format, args...))
	} else {
		logString = fmt.Sprintf("%s None %s", key, fmt.Sprintf(format, args...))
	}
	if lw.transMode == true && len(transactionId) > 0 {
		if lw.logTransaction(transactionId, logString) {
			return
		}
	}
	lw.logger.Print(logString)
}

// log debug. transaction id, component id, log message
func (lw *LogWriter) LogDebug(transactionId string, key string, format string, args ...interface{}) {
	if lw.level >= LevelDebug {
		if key == "" {
			key = "Default"
		}
		if lw.keyEnabled(key) {
			lw.logMessage(transactionId, key, format, args...)
		}
	}
}

//log info. transaction id, component id, log message
func (lw *LogWriter) LogInfo(transactionId string, key string, format string, args ...interface{}) {
	if lw.level >= LevelInfo {
		if key == "" {
			key = "Default"
		}
		if lw.keyEnabled(key) {
			lw.logMessage(transactionId, key, format, args...)
		}
	}
}

//log warning transaction id, component id, log message
func (lw *LogWriter) LogWarn(transactionId string, key string, format string, args ...interface{}) {
	if lw.level >= LevelWarn {
		if key == "" {
			key = "Default"
		}
		if lw.keyEnabled(key) {
			lw.logMessage(transactionId, key, format, args...)
		}
	}
}

//log error transaction id, component id, log message
func (lw *LogWriter) LogError(transactionId string, key string, format string, args ...interface{}) {
	if lw.level >= LevelError {
		if key == "" {
			key = "Default"
		}
		if lw.keyEnabled(key) {
			lw.logMessage(transactionId, key, format, args...)
		}
	}
}
