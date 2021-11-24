//  Copyright 2012-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included
//  in the file licenses/BSL-Couchbase.txt.  As of the Change Date specified
//  in that file, in accordance with the Business Source License, use of this
//  software will be governed by the Apache License, Version 2.0, included in
//  the file licenses/APL2.txt.

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
