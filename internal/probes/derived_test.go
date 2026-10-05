// Copyright (C) 2026 Alex Kunich
// SPDX-License-Identifier: Apache-2.0

package probes

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/mgt-tool/mgtt/sdk/provider"
)

// CloudWatch datapoints come back unordered, one per line; the series is
// reduced as the request's window and derivation say.
const messageCounts = "2026-10-05T12:03:00+00:00\t3900.0\n2026-10-05T12:00:00+00:00\t40.0\n2026-10-05T12:04:00+00:00\t4200.0\n"

func TestMQBroker_QueueDepthDelta(t *testing.T) {
	var args string
	cli := fakeClient(func(a []string) ([]byte, []byte, error) {
		args = strings.Join(a, " ")
		return []byte(messageCounts), nil, nil
	})
	r := provider.NewRegistry()
	registerMQBroker(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "mq_broker", Name: "b-123", Fact: "queue_depth_delta_5m", Window: 5 * time.Minute, Derive: "delta"})
	if err != nil || res.Value != 4160.0 {
		t.Fatalf("got %+v, %v; want a delta of 4160", res, err)
	}
	if !strings.Contains(args, "MessageCount") || !strings.Contains(args, "Datapoints[*].[Timestamp,Sum]") {
		t.Errorf("want every MessageCount datapoint; ran %s", args)
	}
}

func TestRDSInstance_ConnectionCountMax(t *testing.T) {
	cli := fakeClient(func([]string) ([]byte, []byte, error) {
		return []byte("2026-10-05T12:00:00+00:00\t310.0\n2026-10-05T12:02:00+00:00\t512.0\n2026-10-05T12:04:00+00:00\t480.0\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerRDSInstance(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "rds_instance", Name: "db", Fact: "connection_count_max_5m", Window: 5 * time.Minute, Derive: "max"})
	if err != nil || res.Value != 512.0 {
		t.Fatalf("got %+v, %v; want the 512 spike", res, err)
	}
}

// No datapoints in the window is unknown, not zero.
func TestMQBroker_QueueDepthDelta_NoDatapoints(t *testing.T) {
	r := provider.NewRegistry()
	registerMQBroker(r, fakeClient(func([]string) ([]byte, []byte, error) { return []byte("None\n"), nil, nil }))
	_, err := r.Probe(context.Background(), provider.Request{Type: "mq_broker", Name: "b-123", Fact: "queue_depth_delta_5m", Window: 5 * time.Minute, Derive: "delta"})
	if !errors.Is(err, provider.ErrTransient) {
		t.Errorf("want transient; got %v", err)
	}
}
