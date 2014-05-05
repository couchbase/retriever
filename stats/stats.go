//  Copyright (c) 2012-2013 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//  http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package stats

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"
)

type processStats struct {
	CpuNum       int   `json:"cpu_num"`
	GoroutineNum int   `json:"goroutine_num"`
	Gomaxprocs   int   `json:"gomaxprocs"`
	CgoCallNum   int64 `json:"cgo_call_num"`
	// memory
	MemoryAlloc      uint64 `json:"memory_alloc"`
	MemoryTotalAlloc uint64 `json:"memory_total_alloc"`
	MemorySys        uint64 `json:"memory_sys"`
	MemoryLookups    uint64 `json:"memory_lookups"`
	MemoryMallocs    uint64 `json:"memory_mallocs"`
	MemoryFrees      uint64 `json:"memory_frees"`
	// heap
	HeapAlloc    uint64 `json:"heap_alloc"`
	HeapSys      uint64 `json:"heap_sys"`
	HeapIdle     uint64 `json:"heap_idle"`
	HeapInuse    uint64 `json:"heap_inuse"`
	HeapReleased uint64 `json:"heap_released"`
	HeapObjects  uint64 `json:"heap_objects"`
	// gabarage collection
	GcNext uint64 `json:"gc_next"`
	GcLast uint64 `json:"gc_last"`
	GcNum  uint32 `json:"gc_num"`
}

type StatsCollector struct {
	Module   string
	SysStats *processStats
	Stats    map[string]interface{}
	mu       sync.RWMutex
}

func NewStatsCollector(module string) (*StatsCollector, error) {

	if module == "" {
		return nil, fmt.Errorf("Required module name")
	}

	sc := &StatsCollector{Module: module,
		SysStats: &processStats{},
		Stats:    make(map[string]interface{}),
	}
	go handleConnections(sc)
	return sc, nil
}

func (sc *StatsCollector) AddStatKey(key string, initial interface{}) error {

	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}
	sc.mu.RLock()
	_, ok := sc.Stats[key]
	sc.mu.RUnlock()
	if ok { // key already exists
		return fmt.Errorf("key exists")
	}
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.Stats[key] = initial
	return nil
}

func (sc *StatsCollector) UpdateStat(key string, value interface{}) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty ")
	}
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	_, ok := sc.Stats[key]
	if !ok {
		return fmt.Errorf("key exists")
	}
	sc.Stats[key] = value
	return nil
}

func (sc *StatsCollector) IncrementStat(key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty ")
	}
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	value, ok := sc.Stats[key]
	if !ok {
		return fmt.Errorf("key exists")
	}

	switch value := value.(type) {
	case int8:
		sc.Stats[key] = value + 1
	case int16:
		sc.Stats[key] = value + 1
	case int32:
		sc.Stats[key] = value + 1
	case int64:
		sc.Stats[key] = value + 1
	case int:
		sc.Stats[key] = value + 1
	case uint8:
		sc.Stats[key] = value + 1
	case uint16:
		sc.Stats[key] = value + 1
	case uint32:
		sc.Stats[key] = value + 1
	case uint64:
		sc.Stats[key] = value + 1
	case uint:
		sc.Stats[key] = value + 1
	case float32:
		sc.Stats[key] = value + 1
	case float64:
		sc.Stats[key] = value + 1
	default:
		return fmt.Errorf("Unsupported type")
	}

	return nil
}

func (sc *StatsCollector) DecrementStat(key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty ")
	}
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	value, ok := sc.Stats[key]
	if !ok {
		return fmt.Errorf("key exists")
	}
	switch value := value.(type) {
	case int8:
		sc.Stats[key] = value - 1
	case int16:
		sc.Stats[key] = value - 1
	case int32:
		sc.Stats[key] = value - 1
	case int64:
		sc.Stats[key] = value - 1
	case int:
		sc.Stats[key] = value - 1
	case uint8:
		sc.Stats[key] = value - 1
	case uint16:
		sc.Stats[key] = value - 1
	case uint32:
		sc.Stats[key] = value - 1
	case uint64:
		sc.Stats[key] = value - 1
	case uint:
		sc.Stats[key] = value - 1
	case float32:
		sc.Stats[key] = value - 1
	case float64:
		sc.Stats[key] = value - 1
	default:
		return fmt.Errorf("Unsupported type")
	}

	return nil
}

func (sc *StatsCollector) GetStat(key string) interface{} {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	value, ok := sc.Stats[key]
	if !ok {
		return nil
	}
	return value
}

func (sc *StatsCollector) GetAllStat() string {

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	sc.SysStats = &processStats{
		GoroutineNum: runtime.NumGoroutine(),
		Gomaxprocs:   runtime.GOMAXPROCS(0),
		CgoCallNum:   runtime.NumCgoCall(),
		// memory
		MemoryAlloc:      mem.Alloc,
		MemoryTotalAlloc: mem.TotalAlloc,
		MemorySys:        mem.Sys,
		MemoryLookups:    mem.Lookups,
		MemoryMallocs:    mem.Mallocs,
		MemoryFrees:      mem.Frees,
		// heap
		HeapAlloc:    mem.HeapAlloc,
		HeapSys:      mem.HeapSys,
		HeapIdle:     mem.HeapIdle,
		HeapInuse:    mem.HeapInuse,
		HeapReleased: mem.HeapReleased,
		HeapObjects:  mem.HeapObjects,
		// gabarage collection
		GcNext: mem.NextGC,
		GcLast: mem.LastGC,
		GcNum:  mem.NumGC,
	}

	jsonBytes, jsonErr := json.MarshalIndent(sc, "", "    ")
	var body string
	if jsonErr != nil {
		body = jsonErr.Error()
	} else {
		body = string(jsonBytes)
	}

	return body
}

func handleConnections(sc *StatsCollector) {

	// create an I/O channel based on the module name
	// for the server to connect to
	sock := "/tmp/" + "stats_" + sc.Module + ".sock"
	os.Remove(sock)
	listener, err := net.Listen("unix", sock)

	if err != nil {
		fmt.Printf("Failed to listen ", err.Error())
	}

	defer os.Remove(sock)
	defer listener.Close()

	for {
		c, err := listener.Accept()
		if err != nil {
			fmt.Printf("Unable to accept " + err.Error())
			continue
		}
		buf := make([]byte, 512)
		nr, err := c.Read(buf)
		if err != nil {
			fmt.Printf(" Could not read from buffer %s", err.Error())
			c.Close()
			continue
		}
		data := string(buf[0:nr])
		cmds := strings.Split(data, ":")
		switch {
		case strings.Contains(strings.ToLower(cmds[0]), "stats"):
			// rotate the current log file
			statsOutput := sc.GetAllStat()
			c.Write([]byte(statsOutput))
		}
		c.Close()
	}
}
