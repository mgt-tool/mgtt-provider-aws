package probes

import (
	"context"
	"errors"
	"strings"

	"github.com/mgt-tool/mgtt/sdk/provider"
	"github.com/mgt-tool/mgtt/sdk/provider/shell"
)

func registerS3Bucket(r *provider.Registry, cli *shell.Client) {
	const typ = "s3_bucket"
	r.Register(typ, map[string]provider.ProbeFn{
		"accessible": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			_, err := cli.Run(ctx, "s3api", "head-bucket", "--bucket", req.Name)
			if err != nil {
				if errors.Is(err, provider.ErrNotFound) || errors.Is(err, provider.ErrForbidden) {
					return provider.BoolResult(false), nil
				}
				return provider.Result{}, err
			}
			return provider.BoolResult(true), nil
		},
		"versioning_enabled": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			text, err := describeText(ctx, cli,
				"s3api", "get-bucket-versioning",
				"--bucket", req.Name,
				"--query", "Status",
				"--output", "text")
			if err != nil {
				return provider.Result{}, err
			}
			return provider.Result{
				Value:  strings.EqualFold(text, "Enabled"),
				Raw:    text,
				Status: provider.StatusOk,
			}, nil
		},
		"object_count": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			f, err := readCloudWatchStatistic(ctx, cli,
				"AWS/S3", "NumberOfObjects", "Average",
				[]string{dim("BucketName", req.Name), dim("StorageType", "AllStorageTypes")})
			if err != nil {
				return provider.Result{}, err
			}
			return provider.IntResult(int(f)), nil
		},
	})
}
