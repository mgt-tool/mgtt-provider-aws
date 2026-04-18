package probes

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/mgt-tool/mgtt/sdk/provider"
)

func TestACMCertificate_Issued_True(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return []byte("ISSUED\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerACMCertificate(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "acm_certificate", Name: "arn:aws:acm:...", Fact: "issued"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Value != true {
		t.Errorf("want true; got %v", res.Value)
	}
}

func TestACMCertificate_InUse_TrueWhenNonZero(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return []byte("2\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerACMCertificate(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "acm_certificate", Name: "arn", Fact: "in_use"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Value != true {
		t.Errorf("want true; got %v", res.Value)
	}
}

func TestACMCertificate_DaysUntilExpiry_ParsesEpoch(t *testing.T) {
	future := time.Now().Add(60 * 24 * time.Hour).Unix()
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return []byte(fmt.Sprintf("%d.0\n", future)), nil, nil
	})
	r := provider.NewRegistry()
	registerACMCertificate(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "acm_certificate", Name: "arn", Fact: "days_until_expiry"})
	if err != nil {
		t.Fatal(err)
	}
	got, ok := res.Value.(int)
	if !ok {
		t.Fatalf("want int; got %T", res.Value)
	}
	if got < 58 || got > 60 {
		t.Errorf("want ~60; got %d", got)
	}
}

func TestACMCertificate_DaysUntilExpiry_ParsesISO8601(t *testing.T) {
	future := time.Now().Add(30 * 24 * time.Hour).UTC().Format(time.RFC3339)
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return []byte(future + "\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerACMCertificate(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "acm_certificate", Name: "arn", Fact: "days_until_expiry"})
	if err != nil {
		t.Fatal(err)
	}
	got, ok := res.Value.(int)
	if !ok {
		t.Fatalf("want int; got %T", res.Value)
	}
	if got < 28 || got > 30 {
		t.Errorf("want ~30; got %d", got)
	}
}

func TestACMCertificate_MissingName(t *testing.T) {
	r := provider.NewRegistry()
	registerACMCertificate(r, fakeClient(func(args []string) ([]byte, []byte, error) { return nil, nil, nil }))
	_, err := r.Probe(context.Background(), provider.Request{Type: "acm_certificate", Fact: "issued"})
	if !errors.Is(err, provider.ErrUsage) {
		t.Errorf("want ErrUsage; got %v", err)
	}
}
