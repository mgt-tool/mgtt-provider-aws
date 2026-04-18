package probes

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mgt-tool/mgtt/sdk/provider"
)

func TestElasticacheCluster_Available_True(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		if !strings.Contains(strings.Join(args, " "), "describe-replication-groups") {
			t.Errorf("expected describe-replication-groups; got %v", args)
		}
		return []byte("available\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerElasticacheCluster(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "elasticache_cluster", Name: "cache", Fact: "available"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Value != true {
		t.Errorf("want true; got %v", res.Value)
	}
}

func TestElasticacheCluster_ConnectionCount_ReadsCloudWatch(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		joined := strings.Join(args, " ")
		if !strings.Contains(joined, "CurrConnections") {
			t.Errorf("want CurrConnections metric; got %v", args)
		}
		if !strings.Contains(joined, "Name=CacheClusterId,Value=cache") {
			t.Errorf("component name must flow into dimensions; got %v", args)
		}
		return []byte("17.0\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerElasticacheCluster(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "elasticache_cluster", Name: "cache", Fact: "connection_count"})
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := res.Value.(int); !ok || got != 17 {
		t.Errorf("want 17; got %v (%T)", res.Value, res.Value)
	}
}

func TestElasticacheCluster_CacheHitRatio_ZeroWhenNoTraffic(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return []byte("None\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerElasticacheCluster(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "elasticache_cluster", Name: "cache", Fact: "cache_hit_ratio"})
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := res.Value.(float64); !ok || got != 0 {
		t.Errorf("want 0; got %v (%T)", res.Value, res.Value)
	}
}

func TestElasticacheCluster_MissingName(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		t.Fatal("must not invoke aws on empty name")
		return nil, nil, nil
	})
	r := provider.NewRegistry()
	registerElasticacheCluster(r, cli)
	_, err := r.Probe(context.Background(), provider.Request{Type: "elasticache_cluster", Fact: "available"})
	if !errors.Is(err, provider.ErrUsage) {
		t.Errorf("want ErrUsage; got %v", err)
	}
}
