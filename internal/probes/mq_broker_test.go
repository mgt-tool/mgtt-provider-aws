package probes

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mgt-tool/mgtt/sdk/provider"
)

func TestMQBroker_Available_RunningIsTrue(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		if !strings.Contains(strings.Join(args, " "), "mq describe-broker") {
			t.Errorf("want mq describe-broker; got %v", args)
		}
		return []byte("RUNNING\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerMQBroker(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "mq_broker", Name: "b-123", Fact: "available"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Value != true || res.Raw != "RUNNING" {
		t.Errorf("want true + Raw=RUNNING; got %v raw=%q", res.Value, res.Raw)
	}
}

func TestMQBroker_QueueDepth_ReadsMessageCount(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		joined := strings.Join(args, " ")
		if !strings.Contains(joined, "MessageCount") {
			t.Errorf("want MessageCount; got %v", args)
		}
		return []byte("250.0\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerMQBroker(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "mq_broker", Name: "b-123", Fact: "queue_depth"})
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := res.Value.(int); !ok || got != 250 {
		t.Errorf("want 250; got %v", res.Value)
	}
}

func TestMQBroker_MissingName(t *testing.T) {
	r := provider.NewRegistry()
	registerMQBroker(r, fakeClient(func(args []string) ([]byte, []byte, error) { return nil, nil, nil }))
	_, err := r.Probe(context.Background(), provider.Request{Type: "mq_broker", Fact: "available"})
	if !errors.Is(err, provider.ErrUsage) {
		t.Errorf("want ErrUsage; got %v", err)
	}
}
