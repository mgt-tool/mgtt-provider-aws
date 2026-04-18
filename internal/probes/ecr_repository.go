package probes

import (
	"context"
	"errors"
	"time"

	"github.com/mgt-tool/mgtt/sdk/provider"
	"github.com/mgt-tool/mgtt/sdk/provider/shell"
)

func registerECRRepository(r *provider.Registry, cli *shell.Client) {
	const typ = "ecr_repository"
	r.Register(typ, map[string]provider.ProbeFn{
		"exists": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			_, err := cli.Run(ctx,
				"ecr", "describe-repositories",
				"--repository-names", req.Name,
				"--output", "text")
			if err != nil {
				if errors.Is(err, provider.ErrNotFound) {
					return provider.BoolResult(false), nil
				}
				return provider.Result{}, err
			}
			return provider.BoolResult(true), nil
		},
		"image_count": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			text, err := describeText(ctx, cli,
				"ecr", "describe-images",
				"--repository-name", req.Name,
				"--query", "length(imageDetails)",
				"--output", "text")
			if err != nil {
				return provider.Result{}, err
			}
			n, err := parseIntOrZero(text)
			if err != nil {
				return provider.Result{}, err
			}
			return provider.IntResult(n), nil
		},
		"latest_push_age_seconds": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			text, err := describeText(ctx, cli,
				"ecr", "describe-images",
				"--repository-name", req.Name,
				"--query", "max(imageDetails[].imagePushedAt)",
				"--output", "text")
			if err != nil {
				return provider.Result{}, err
			}
			t, err := parseAWSTimestamp(text)
			if err != nil {
				return provider.Result{}, err
			}
			if t.IsZero() {
				return provider.IntResult(0), nil
			}
			age := time.Since(t).Seconds()
			if age < 0 {
				age = 0
			}
			return provider.IntResult(int(age)), nil
		},
	})
}
