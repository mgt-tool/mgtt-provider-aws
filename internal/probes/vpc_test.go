package probes

import (
	"context"
	"errors"
	"testing"

	"github.com/mgt-tool/mgtt/sdk/provider"
)

func TestVPC_Available_True(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return []byte("available\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerVPC(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "vpc", Name: "vpc-1", Fact: "available"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Value != true {
		t.Errorf("want true; got %v", res.Value)
	}
}

func TestVPC_IPUtilization_FromSubnets(t *testing.T) {
	// /29 has 8 raw addrs − 5 AWS-reserved = 3 usable; 1 free → 2 used; 66.67%
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return []byte(`[{"cidr":"10.0.0.0/29","free":1}]`), nil, nil
	})
	r := provider.NewRegistry()
	registerVPC(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "vpc", Name: "vpc-1", Fact: "ip_utilization"})
	if err != nil {
		t.Fatal(err)
	}
	got, ok := res.Value.(float64)
	if !ok {
		t.Fatalf("want float64; got %T", res.Value)
	}
	if got < 66 || got > 67 {
		t.Errorf("want ~66.67; got %v", got)
	}
}

func TestVPC_IPUtilization_EmptyVPC(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return []byte(`[]`), nil, nil
	})
	r := provider.NewRegistry()
	registerVPC(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "vpc", Name: "vpc-1", Fact: "ip_utilization"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Value != 0.0 {
		t.Errorf("no subnets must yield 0; got %v", res.Value)
	}
}

func TestVPC_MissingName(t *testing.T) {
	r := provider.NewRegistry()
	registerVPC(r, fakeClient(func(args []string) ([]byte, []byte, error) { return nil, nil, nil }))
	_, err := r.Probe(context.Background(), provider.Request{Type: "vpc", Fact: "available"})
	if !errors.Is(err, provider.ErrUsage) {
		t.Errorf("want ErrUsage; got %v", err)
	}
}
