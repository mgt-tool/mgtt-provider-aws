package probes

import (
	"context"
	"errors"
	"time"

	"github.com/mgt-tool/mgtt/sdk/provider"
	"github.com/mgt-tool/mgtt/sdk/provider/shell"
)

func registerSSMParameter(r *provider.Registry, cli *shell.Client) {
	const typ = "ssm_parameter"
	r.Register(typ, map[string]provider.ProbeFn{
		"exists": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			_, err := cli.Run(ctx,
				"ssm", "get-parameter",
				"--name", req.Name,
				"--output", "text")
			if err != nil {
				if errors.Is(err, provider.ErrNotFound) {
					return provider.BoolResult(false), nil
				}
				return provider.Result{}, err
			}
			return provider.BoolResult(true), nil
		},
		"version": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			text, err := describeText(ctx, cli,
				"ssm", "get-parameter",
				"--name", req.Name,
				"--query", "Parameter.Version",
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
		"last_modified_age_seconds": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			text, err := describeText(ctx, cli,
				"ssm", "describe-parameters",
				"--filters", "Key=Name,Values="+req.Name,
				"--query", "Parameters[0].LastModifiedDate",
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
