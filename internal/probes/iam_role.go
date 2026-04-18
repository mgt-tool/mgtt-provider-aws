package probes

import (
	"context"
	"errors"

	"github.com/mgt-tool/mgtt/sdk/provider"
	"github.com/mgt-tool/mgtt/sdk/provider/shell"
)

func registerIAMRole(r *provider.Registry, cli *shell.Client) {
	const typ = "iam_role"
	r.Register(typ, map[string]provider.ProbeFn{
		"exists": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			_, err := cli.Run(ctx,
				"iam", "get-role",
				"--role-name", req.Name,
				"--output", "text")
			if err != nil {
				if errors.Is(err, provider.ErrNotFound) {
					return provider.BoolResult(false), nil
				}
				return provider.Result{}, err
			}
			return provider.BoolResult(true), nil
		},
		// assumable is a static check: does the role have a trust policy at
		// all. A full assume-role-as-caller check would require calling
		// sts:AssumeRole, which is a write-ish side effect we intentionally
		// avoid in probes.
		"assumable": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			text, err := describeText(ctx, cli,
				"iam", "get-role",
				"--role-name", req.Name,
				"--query", "Role.AssumeRolePolicyDocument",
				"--output", "text")
			if err != nil {
				return provider.Result{}, err
			}
			return provider.Result{
				Value:  text != "" && text != "None",
				Raw:    text,
				Status: provider.StatusOk,
			}, nil
		},
		"attached_policy_count": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			text, err := describeText(ctx, cli,
				"iam", "list-attached-role-policies",
				"--role-name", req.Name,
				"--query", "length(AttachedPolicies)",
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
	})
}
