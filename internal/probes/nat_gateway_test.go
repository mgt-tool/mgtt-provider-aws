package probes

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mgt-tool/mgtt/sdk/provider"
)

func TestNATGateway_Available_True(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return []byte("available\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerNATGateway(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "nat_gateway", Name: "nat-1", Fact: "available"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Value != true {
		t.Errorf("want true; got %v", res.Value)
	}
}

func TestNATGateway_ErrorPortAllocation_Reads(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		if !strings.Contains(strings.Join(args, " "), "ErrorPortAllocation") {
			t.Errorf("want ErrorPortAllocation metric; got %v", args)
		}
		return []byte("5.0\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerNATGateway(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "nat_gateway", Name: "nat-1", Fact: "error_port_allocation_count"})
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := res.Value.(int); !ok || got != 5 {
		t.Errorf("want 5; got %v", res.Value)
	}
}

func TestNATGateway_BytesOut_DividesByPeriod(t *testing.T) {
	// 600 bytes over 60s should be 10 B/s
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return []byte("600.0\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerNATGateway(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "nat_gateway", Name: "nat-1", Fact: "bytes_out_per_second"})
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := res.Value.(int); !ok || got != 10 {
		t.Errorf("want 10; got %v", res.Value)
	}
}

func TestNATGateway_MissingName(t *testing.T) {
	r := provider.NewRegistry()
	registerNATGateway(r, fakeClient(func(args []string) ([]byte, []byte, error) { return nil, nil, nil }))
	_, err := r.Probe(context.Background(), provider.Request{Type: "nat_gateway", Fact: "available"})
	if !errors.Is(err, provider.ErrUsage) {
		t.Errorf("want ErrUsage; got %v", err)
	}
}
