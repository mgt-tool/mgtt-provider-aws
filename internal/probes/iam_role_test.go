package probes

import (
	"context"
	"errors"
	"testing"

	"github.com/mgt-tool/mgtt/sdk/provider"
)

func TestIAMRole_Exists_True(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return []byte("arn:aws:iam::111:role/r\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerIAMRole(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "iam_role", Name: "r", Fact: "exists"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Value != true {
		t.Errorf("want true; got %v", res.Value)
	}
}

func TestIAMRole_Exists_FalseWhenNotFound(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return nil,
			[]byte("An error occurred (NoSuchEntity) when calling the GetRole operation: The role with name ghost cannot be found."),
			errors.New("exit status 254")
	})
	r := provider.NewRegistry()
	registerIAMRole(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "iam_role", Name: "ghost", Fact: "exists"})
	if err != nil {
		t.Fatalf("not-found should map to false; got err=%v", err)
	}
	if res.Value != false {
		t.Errorf("want false; got %v", res.Value)
	}
}

func TestIAMRole_MissingName(t *testing.T) {
	r := provider.NewRegistry()
	registerIAMRole(r, fakeClient(func(args []string) ([]byte, []byte, error) { return nil, nil, nil }))
	_, err := r.Probe(context.Background(), provider.Request{Type: "iam_role", Fact: "exists"})
	if !errors.Is(err, provider.ErrUsage) {
		t.Errorf("want ErrUsage; got %v", err)
	}
}

func TestIAMRole_Assumable_TrueWhenTrustPolicy(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return []byte("{\"Version\":\"2012-10-17\"}\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerIAMRole(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "iam_role", Name: "r", Fact: "assumable"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Value != true {
		t.Errorf("want true; got %v", res.Value)
	}
}

func TestIAMRole_AttachedPolicyCount_ReadsLength(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return []byte("4\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerIAMRole(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "iam_role", Name: "r", Fact: "attached_policy_count"})
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := res.Value.(int); !ok || got != 4 {
		t.Errorf("want 4; got %v", res.Value)
	}
}
