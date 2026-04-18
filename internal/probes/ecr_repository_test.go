package probes

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/mgt-tool/mgtt/sdk/provider"
)

func TestECRRepository_Exists_True(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return []byte("repo\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerECRRepository(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "ecr_repository", Name: "repo", Fact: "exists"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Value != true {
		t.Errorf("want true; got %v", res.Value)
	}
}

func TestECRRepository_Exists_FalseWhenNotFound(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return nil,
			[]byte("An error occurred (RepositoryNotFoundException) when calling the DescribeRepositories operation: The repository with name 'ghost' does not exist in the registry with id '1234'"),
			errors.New("exit status 254")
	})
	r := provider.NewRegistry()
	registerECRRepository(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "ecr_repository", Name: "ghost", Fact: "exists"})
	if err != nil {
		t.Fatalf("not-found should map to false; got err=%v", err)
	}
	if res.Value != false {
		t.Errorf("want false; got %v", res.Value)
	}
}

func TestECRRepository_LatestPushAge_ParsesISO8601(t *testing.T) {
	// aws-cli v2 emits ISO8601 timestamps by default.
	pushed := time.Now().Add(-90 * time.Second).UTC().Format(time.RFC3339)
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return []byte(pushed + "\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerECRRepository(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "ecr_repository", Name: "repo", Fact: "latest_push_age_seconds"})
	if err != nil {
		t.Fatal(err)
	}
	got, ok := res.Value.(int)
	if !ok {
		t.Fatalf("want int; got %T", res.Value)
	}
	if got < 80 || got > 120 {
		t.Errorf("want ~90; got %d", got)
	}
}

func TestECRRepository_MissingName(t *testing.T) {
	r := provider.NewRegistry()
	registerECRRepository(r, fakeClient(func(args []string) ([]byte, []byte, error) { return nil, nil, nil }))
	_, err := r.Probe(context.Background(), provider.Request{Type: "ecr_repository", Fact: "exists"})
	if !errors.Is(err, provider.ErrUsage) {
		t.Errorf("want ErrUsage; got %v", err)
	}
}

func TestECRRepository_ImageCount_ReadsLength(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		if !strings.Contains(strings.Join(args, " "), "describe-images") {
			t.Errorf("want describe-images; got %v", args)
		}
		return []byte("12\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerECRRepository(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "ecr_repository", Name: "repo", Fact: "image_count"})
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := res.Value.(int); !ok || got != 12 {
		t.Errorf("want 12; got %v", res.Value)
	}
}

func TestECRRepository_LatestPushAge_NoneIsZero(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return []byte("None\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerECRRepository(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "ecr_repository", Name: "repo", Fact: "latest_push_age_seconds"})
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := res.Value.(int); !ok || got != 0 {
		t.Errorf("empty repo must be 0; got %v", res.Value)
	}
}
