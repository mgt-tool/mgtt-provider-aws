package probes

import (
	"context"
	"errors"

	"github.com/mgt-tool/mgtt/sdk/provider"
	"github.com/mgt-tool/mgtt/sdk/provider/shell"
)

func registerSecurityGroup(r *provider.Registry, cli *shell.Client) {
	const typ = "security_group"
	r.Register(typ, map[string]provider.ProbeFn{
		"exists": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			_, err := cli.Run(ctx,
				"ec2", "describe-security-groups",
				"--group-ids", req.Name,
				"--output", "text")
			if err != nil {
				if errors.Is(err, provider.ErrNotFound) {
					return provider.BoolResult(false), nil
				}
				return provider.Result{}, err
			}
			return provider.BoolResult(true), nil
		},
		"ingress_rule_count": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			text, err := describeText(ctx, cli,
				"ec2", "describe-security-groups",
				"--group-ids", req.Name,
				"--query", "length(SecurityGroups[0].IpPermissions)",
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
		"egress_rule_count": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			text, err := describeText(ctx, cli,
				"ec2", "describe-security-groups",
				"--group-ids", req.Name,
				"--query", "length(SecurityGroups[0].IpPermissionsEgress)",
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
