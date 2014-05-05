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
	"fmt"
	"testing"
)

func TestStats(t *testing.T) {
	sc, err := NewStatsCollector("test")
	if err != nil {
		t.Errorf("Failed %s", err.Error())
	}

	sc.AddStatKey("Connections", 0)
	sc.AddStatKey("Failures", 0)
	sc.AddStatKey("Transport", "tcp")
	sc.AddStatKey("RTT", 0)

	var conn uint16
	var failures uint64
	var rtt float64

	conn = 16
	err = sc.UpdateStat("Connections", conn)
	if err != nil {
		t.Errorf("Failed %s", err.Error())
	}
	failures = 400
	err = sc.UpdateStat("Failures", failures)
	if err != nil {
		t.Errorf("Failed %s", err.Error())
	}
	err = sc.UpdateStat("Transport", "udp")
	if err != nil {
		t.Errorf("Failed %s", err.Error())
	}
	rtt = 3.142
	err = sc.UpdateStat("RTT", rtt)
	if err != nil {
		t.Errorf("Failed %s", err.Error())
	}

	// increment -decrement operations
	err = sc.IncrementStat("Connections")
	if err != nil {
		t.Errorf("Failed %s", err.Error())
	}
	err = sc.DecrementStat("Failures")
	if err != nil {
		t.Errorf("Failed %s", err.Error())
	}
	err = sc.IncrementStat("RTT")
	if err != nil {
		t.Errorf("failed %s", err.Error())
	}

	// should fail
	err = sc.IncrementStat("Transport")
	if err == nil {
		t.Errorf("failed %s")
	}

	// get stats

	connections := sc.GetStat("Connections").(uint16) + 1

	fmt.Println("Connections ", connections)
	fmt.Println("Failures ", sc.GetStat("Failures"))
	fmt.Println("RTT ", sc.GetStat("RTT"))
	fmt.Println("Transport ", sc.GetStat("Transport"))

	//get all stats
	stats := sc.GetAllStat()
	fmt.Printf(" Stats : %s", stats)

}
