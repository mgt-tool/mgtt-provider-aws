package probes

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mgt-tool/mgtt/sdk/provider"
)

func TestS3Bucket_Accessible_True(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		if !strings.Contains(strings.Join(args, " "), "head-bucket") {
			t.Errorf("want head-bucket; got %v", args)
		}
		return nil, nil, nil
	})
	r := provider.NewRegistry()
	registerS3Bucket(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "s3_bucket", Name: "b", Fact: "accessible"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Value != true {
		t.Errorf("want true; got %v", res.Value)
	}
}

func TestS3Bucket_Accessible_NotFoundReturnsFalse(t *testing.T) {
	// Real aws-cli v2 head-bucket stderr for a missing bucket — no
	// NoSuchBucket code, just a bare 404.
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return nil,
			[]byte("An error occurred (404) when calling the HeadBucket operation: Not Found"),
			errors.New("exit status 254")
	})
	r := provider.NewRegistry()
	registerS3Bucket(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "s3_bucket", Name: "ghost", Fact: "accessible"})
	if err != nil {
		t.Fatalf("must suppress not-found into bool=false; got err=%v", err)
	}
	if res.Value != false {
		t.Errorf("missing bucket must map to accessible=false; got %v", res.Value)
	}
}

func TestS3Bucket_Accessible_ForbiddenReturnsFalse(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		return nil,
			[]byte("An error occurred (403) when calling the HeadBucket operation: Forbidden"),
			errors.New("exit status 254")
	})
	r := provider.NewRegistry()
	registerS3Bucket(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "s3_bucket", Name: "locked", Fact: "accessible"})
	if err != nil {
		t.Fatalf("must suppress forbidden into bool=false; got err=%v", err)
	}
	if res.Value != false {
		t.Errorf("forbidden bucket must map to accessible=false; got %v", res.Value)
	}
}

func TestS3Bucket_MissingName(t *testing.T) {
	r := provider.NewRegistry()
	registerS3Bucket(r, fakeClient(func(args []string) ([]byte, []byte, error) { return nil, nil, nil }))
	_, err := r.Probe(context.Background(), provider.Request{Type: "s3_bucket", Fact: "accessible"})
	if !errors.Is(err, provider.ErrUsage) {
		t.Errorf("want ErrUsage; got %v", err)
	}
}

func TestS3Bucket_VersioningEnabled_True(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		if !strings.Contains(strings.Join(args, " "), "get-bucket-versioning") {
			t.Errorf("want get-bucket-versioning; got %v", args)
		}
		return []byte("Enabled\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerS3Bucket(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "s3_bucket", Name: "b", Fact: "versioning_enabled"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Value != true {
		t.Errorf("want true; got %v", res.Value)
	}
}

func TestS3Bucket_ObjectCount_TwoDimensions(t *testing.T) {
	cli := fakeClient(func(args []string) ([]byte, []byte, error) {
		joined := strings.Join(args, " ")
		if !strings.Contains(joined, "Name=BucketName,Value=b") || !strings.Contains(joined, "Name=StorageType,Value=AllStorageTypes") {
			t.Errorf("both dims required; got %v", args)
		}
		return []byte("42.0\n"), nil, nil
	})
	r := provider.NewRegistry()
	registerS3Bucket(r, cli)
	res, err := r.Probe(context.Background(), provider.Request{Type: "s3_bucket", Name: "b", Fact: "object_count"})
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := res.Value.(int); !ok || got != 42 {
		t.Errorf("want 42; got %v", res.Value)
	}
}
