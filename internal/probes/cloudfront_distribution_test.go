package probes

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mgt-tool/mgtt/sdk/provider"
)

func TestCloudFrontDistribution_Deployed_True(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return []byte("Deployed\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerCloudFrontDistribution(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "cloudfront_distribution", Name: "E123", Fact: "deployed"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Value != true {
		t.Errorf("want true; got %v", res.Value)
	}
}

func TestCloudFrontDistribution_Enabled_True(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		if !strings.Contains(strings.Join(args, " "), "DistributionConfig.Enabled") {
			t.Errorf("want DistributionConfig.Enabled query; got %v", args)
		}
		return []byte("True\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerCloudFrontDistribution(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "cloudfront_distribution", Name: "E123", Fact: "enabled"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Value != true {
		t.Errorf("want true; got %v", res.Value)
	}
}

func TestCloudFrontDistribution_ErrorRate5xx_Reads(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		joined := strings.Join(args, " ")
		if !strings.Contains(joined, "5xxErrorRate") {
			t.Errorf("want 5xxErrorRate metric; got %v", args)
		}
		if !strings.Contains(joined, "Name=DistributionId,Value=E123") || !strings.Contains(joined, "Name=Region,Value=Global") {
			t.Errorf("both dims required; got %v", args)
		}
		return []byte("0.42\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerCloudFrontDistribution(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "cloudfront_distribution", Name: "E123", Fact: "error_rate_5xx"})
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := res.Value.(float64); !ok || got != 0.42 {
		t.Errorf("want 0.42; got %v", res.Value)
	}
}

func TestCloudFrontDistribution_MissingName(t *testing.T) {
	r := provider.NewRegistry()
	registerCloudFrontDistribution(r, fakeClient(func(args []string) ([]byte, []byte, error) { return nil, nil, nil }))
	_, err := r.Probe(context.Background(), provider.Request{Type: "cloudfront_distribution", Fact: "deployed"})
	if !errors.Is(err, provider.ErrUsage) {
		t.Errorf("want ErrUsage; got %v", err)
	}
}
