package probes

import (
	"context"
	"errors"
	"testing"

	"github.com/mgt-tool/mgtt/sdk/provider"
)

func TestSecurityGroup_Exists_True(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return []byte("sg-123\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerSecurityGroup(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "security_group", Name: "sg-123", Fact: "exists"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Value != true {
		t.Errorf("want true; got %v", res.Value)
	}
}

func TestSecurityGroup_Exists_FalseWhenNotFound(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return nil,
			[]byte("An error occurred (InvalidGroup.NotFound) when calling the DescribeSecurityGroups operation: The security group 'sg-ghost' does not exist"),
			errors.New("exit status 254")
	})
	r := provider.NewRegistry()
	registerSecurityGroup(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "security_group", Name: "sg-ghost", Fact: "exists"})
	if err != nil {
		t.Fatalf("not-found should map to false; got err=%v", err)
	}
	if res.Value != false {
		t.Errorf("want false; got %v", res.Value)
	}
}

func TestSecurityGroup_MissingName(t *testing.T) {
	r := provider.NewRegistry()
	registerSecurityGroup(r, fakeClient(func(args []string) ([]byte, []byte, error) { return nil, nil, nil }))
	_, err := r.Probe(context.Background(), provider.Request{Type: "security_group", Fact: "exists"})
	if !errors.Is(err, provider.ErrUsage) {
		t.Errorf("want ErrUsage; got %v", err)
	}
}

func TestSecurityGroup_IngressRuleCount_Reads(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return []byte("3\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerSecurityGroup(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "security_group", Name: "sg-123", Fact: "ingress_rule_count"})
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := res.Value.(int); !ok || got != 3 {
		t.Errorf("want 3; got %v", res.Value)
	}
}

func TestSecurityGroup_EgressRuleCount_Reads(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return []byte("1\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerSecurityGroup(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "security_group", Name: "sg-123", Fact: "egress_rule_count"})
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := res.Value.(int); !ok || got != 1 {
		t.Errorf("want 1; got %v", res.Value)
	}
}
