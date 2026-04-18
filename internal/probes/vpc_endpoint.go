package probes

import (
	"context"
	"strings"

	"github.com/mgt-tool/mgtt/sdk/provider"
	"github.com/mgt-tool/mgtt/sdk/provider/shell"
)

func registerVPCEndpoint(r *provider.Registry, cli *shell.Client) {
	const typ = "vpc_endpoint"
	r.Register(typ, map[string]provider.ProbeFn{
		"available": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			text, err := describeText(ctx, cli,
				"ec2", "describe-vpc-endpoints",
				"--vpc-endpoint-ids", req.Name,
				"--query", "VpcEndpoints[0].State",
				"--output", "text")
			if err != nil {
				return provider.Result{}, err
			}
			return provider.Result{
				Value:  strings.EqualFold(text, "available"),
				Raw:    text,
				Status: provider.StatusOk,
			}, nil
		},
		"dns_enabled": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			text, err := describeText(ctx, cli,
				"ec2", "describe-vpc-endpoints",
				"--vpc-endpoint-ids", req.Name,
				"--query", "VpcEndpoints[0].PrivateDnsEnabled",
				"--output", "text")
			if err != nil {
				return provider.Result{}, err
			}
			return provider.Result{
				Value:  strings.EqualFold(text, "True"),
				Raw:    text,
				Status: provider.StatusOk,
			}, nil
		},
	})
}
