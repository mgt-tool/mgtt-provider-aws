package probes

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mgt-tool/mgtt-provider-aws/internal/awsclassify"
	"github.com/mgt-tool/mgtt/sdk/provider"
	"github.com/mgt-tool/mgtt/sdk/provider/shell"
)

// fakeClient returns a shell.Client whose Exec is stubbed with `fn`.
// The classifier is preserved so stderr-to-sentinel translation still
// runs on error paths — that's a real part of the probe contract.
func fakeClient(fn func(args []string) (stdout, stderr []byte, err error)) *shell.Client {
	c := shell.New("aws")
	c.Classify = awsclassify.Classify
	c.Exec = func(ctx context.Context, args ...string) ([]byte, []byte, error) {
		return fn(args)
	}
	return c
}

func TestRDSInstance_Available_True(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		if !contains(strings.Join(args, " "), "describe-db-instances") {
			t.Errorf("expected describe-db-instances; got %v", args)
		}
		return []byte("available\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerRDSInstance(r, cli)

	res, err := r.Probe(context.Background(), provider.Request{
		Type: "rds_instance",
		Name: "prod-db",
		Fact: "available",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Value != true {
		t.Errorf("status 'available' must map to Value=true; got %v", res.Value)
	}
	if res.Raw != "available" {
		t.Errorf("Raw must be the backend string; got %q", res.Raw)
	}
}

func TestRDSInstance_Available_OtherStatuses(t *testing.T) {
	cases := []string{"stopped", "rebooting", "maintenance", "failed"}
	for _, s := range cases {
		s := s
		t.Run(s, func(t *testing.T) {
			cli := fakeClient(func(args []string) ([]byte, []byte, error) {
				return []byte(s + "\n"), nil, nil
			})
			r := provider.NewRegistry()
			registerRDSInstance(r, cli)

			res, err := r.Probe(context.Background(), provider.Request{
				Type: "rds_instance", Name: "db", Fact: "available",
			})
			if err != nil {
				t.Fatal(err)
			}
			if res.Value != false {
				t.Errorf("status %q must map to Value=false; got %v", s, res.Value)
			}
			if res.Raw != s {
				t.Errorf("Raw must preserve backend status %q; got %q", s, res.Raw)
			}
		})
	}
}

func TestRDSInstance_Available_NotFound(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return nil,
			[]byte("An error occurred (DBInstanceNotFound) when calling the DescribeDBInstances operation: DBInstance ghost not found.\n"),
			errors.New("exit status 254")
	})
	r := provider.NewRegistry()
	registerRDSInstance(r, cli)

	res, err := r.Probe(context.Background(), provider.Request{
		Type: "rds_instance", Name: "ghost", Fact: "available",
	})
	if err != nil {
		t.Fatalf("ErrNotFound must be translated to Status=not_found, not propagated as err; got err=%v", err)
	}
	if res.Status != provider.StatusNotFound {
		t.Errorf("want Status=not_found; got %q", res.Status)
	}
}

func TestRDSInstance_ConnectionCount_ReadsDatapoint(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		joined := strings.Join(args, " ")
		if !contains(joined, "cloudwatch get-metric-statistics") {
			t.Errorf("want cloudwatch get-metric-statistics; got %v", args)
		}
		if !contains(joined, "Name=DBInstanceIdentifier,Value=prod-db") {
			t.Errorf("component name must flow into --dimensions; got %v", args)
		}
		return []byte("42.0\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerRDSInstance(r, cli)

	res, err := r.Probe(context.Background(), provider.Request{
		Type: "rds_instance", Name: "prod-db", Fact: "connection_count",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := res.Value.(int); !ok || got != 42 {
		t.Errorf("want Value=int(42); got %v (%T)", res.Value, res.Value)
	}
}

func TestRDSInstance_ConnectionCount_NoneIsZero(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return []byte("None\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerRDSInstance(r, cli)

	res, err := r.Probe(context.Background(), provider.Request{
		Type: "rds_instance", Name: "db", Fact: "connection_count",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := res.Value.(int); !ok || got != 0 {
		t.Errorf("aws-cli 'None' must map to 0; got %v", res.Value)
	}
}

func TestRDSInstance_ConnectionCount_ProtocolErrorOnGarbage(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return []byte("not-a-number\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerRDSInstance(r, cli)

	_, err := r.Probe(context.Background(), provider.Request{
		Type: "rds_instance", Name: "db", Fact: "connection_count",
	})
	if !errors.Is(err, provider.ErrProtocol) {
		t.Errorf("unparseable aws output must be ErrProtocol; got %v", err)
	}
}

func TestRDSInstance_Available_MissingName(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		t.Fatal("aws must not be invoked when name is empty")
		return nil, nil, nil
	})
	r := provider.NewRegistry()
	registerRDSInstance(r, cli)

	_, err := r.Probe(context.Background(), provider.Request{
		Type: "rds_instance", Fact: "available",
	})
	if !errors.Is(err, provider.ErrUsage) {
		t.Errorf("empty component name must be ErrUsage; got %v", err)
	}
}

func contains(s, sub string) bool {
	return strings.Contains(s, sub)
}
