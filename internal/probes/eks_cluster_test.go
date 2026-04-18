package probes

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mgt-tool/mgtt/sdk/provider"
)

func TestEKSCluster_Active_True(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return []byte("ACTIVE\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerEKSCluster(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "eks_cluster", Name: "prod", Fact: "active"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Value != true {
		t.Errorf("want true; got %v", res.Value)
	}
}

func TestEKSCluster_ApiReachable_TrueOnEndpoint(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return []byte("https://AAAA.gr7.us-east-1.eks.amazonaws.com\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerEKSCluster(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "eks_cluster", Name: "prod", Fact: "api_reachable"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Value != true {
		t.Errorf("want true; got %v", res.Value)
	}
}

func TestEKSCluster_NodeCount_ListsNodegroups(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		if !strings.Contains(strings.Join(args, " "), "list-nodegroups") {
			t.Errorf("want list-nodegroups; got %v", args)
		}
		return []byte("3\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerEKSCluster(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "eks_cluster", Name: "prod", Fact: "node_count"})
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := res.Value.(int); !ok || got != 3 {
		t.Errorf("want 3; got %v", res.Value)
	}
}

func TestEKSCluster_MissingName(t *testing.T) {
	r := provider.NewRegistry()
	registerEKSCluster(r, fakeClient(func(args []string) ([]byte, []byte, error) { return nil, nil, nil }))
	_, err := r.Probe(context.Background(), provider.Request{Type: "eks_cluster", Fact: "active"})
	if !errors.Is(err, provider.ErrUsage) {
		t.Errorf("want ErrUsage; got %v", err)
	}
}
