package probes

import (
	"context"
	"errors"
	"testing"

	"github.com/mgt-tool/mgtt/sdk/provider"
)

func TestVPCEndpoint_Available_True(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return []byte("Available\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerVPCEndpoint(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "vpc_endpoint", Name: "vpce-1", Fact: "available"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Value != true {
		t.Errorf("want true; got %v", res.Value)
	}
}

func TestVPCEndpoint_DNSEnabled_True(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return []byte("True\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerVPCEndpoint(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "vpc_endpoint", Name: "vpce-1", Fact: "dns_enabled"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Value != true {
		t.Errorf("want true; got %v", res.Value)
	}
}

func TestVPCEndpoint_MissingName(t *testing.T) {
	r := provider.NewRegistry()
	registerVPCEndpoint(r, fakeClient(func(args []string) ([]byte, []byte, error) { return nil, nil, nil }))
	_, err := r.Probe(context.Background(), provider.Request{Type: "vpc_endpoint", Fact: "available"})
	if !errors.Is(err, provider.ErrUsage) {
		t.Errorf("want ErrUsage; got %v", err)
	}
}
