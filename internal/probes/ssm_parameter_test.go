package probes

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/mgt-tool/mgtt/sdk/provider"
)

func TestSSMParameter_Exists_True(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return []byte("value\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerSSMParameter(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "ssm_parameter", Name: "/cfg", Fact: "exists"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Value != true {
		t.Errorf("want true; got %v", res.Value)
	}
}

func TestSSMParameter_Exists_FalseWhenNotFound(t *testing.T) {
	// Real aws-cli v2 stderr for a missing SSM parameter — note the message
	// uses "not found" (lowercase) with no explicit "does not exist".
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return nil,
			[]byte("An error occurred (ParameterNotFound) when calling the GetParameter operation: Parameter /x not found."),
			errors.New("exit status 254")
	})
	r := provider.NewRegistry()
	registerSSMParameter(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "ssm_parameter", Name: "/x", Fact: "exists"})
	if err != nil {
		t.Fatalf("not-found should map to false; got err=%v", err)
	}
	if res.Value != false {
		t.Errorf("want false; got %v", res.Value)
	}
}

func TestSSMParameter_LastModifiedAge_ParsesISO8601(t *testing.T) {
	modified := time.Now().Add(-5 * time.Minute).UTC().Format(time.RFC3339)
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return []byte(modified + "\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerSSMParameter(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "ssm_parameter", Name: "/cfg", Fact: "last_modified_age_seconds"})
	if err != nil {
		t.Fatal(err)
	}
	got, ok := res.Value.(int)
	if !ok {
		t.Fatalf("want int; got %T", res.Value)
	}
	if got < 290 || got > 330 {
		t.Errorf("want ~300s; got %d", got)
	}
}

func TestSSMParameter_MissingName(t *testing.T) {
	r := provider.NewRegistry()
	registerSSMParameter(r, fakeClient(func(args []string) ([]byte, []byte, error) { return nil, nil, nil }))
	_, err := r.Probe(context.Background(), provider.Request{Type: "ssm_parameter", Fact: "exists"})
	if !errors.Is(err, provider.ErrUsage) {
		t.Errorf("want ErrUsage; got %v", err)
	}
}

func TestSSMParameter_Version_ReadsInt(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		if !strings.Contains(strings.Join(args, " "), "Parameter.Version") {
			t.Errorf("want Parameter.Version query; got %v", args)
		}
		return []byte("7\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerSSMParameter(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "ssm_parameter", Name: "/cfg", Fact: "version"})
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := res.Value.(int); !ok || got != 7 {
		t.Errorf("want 7; got %v", res.Value)
	}
}
