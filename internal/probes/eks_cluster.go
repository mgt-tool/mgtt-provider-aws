package probes

import (
	"context"
	"strings"

	"github.com/mgt-tool/mgtt/sdk/provider"
	"github.com/mgt-tool/mgtt/sdk/provider/shell"
)

func registerEKSCluster(r *provider.Registry, cli *shell.Client) {
	const typ = "eks_cluster"
	r.Register(typ, map[string]provider.ProbeFn{
		"active": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			text, err := describeText(ctx, cli,
				"eks", "describe-cluster",
				"--name", req.Name,
				"--query", "cluster.status",
				"--output", "text")
			if err != nil {
				return provider.Result{}, err
			}
			return provider.Result{
				Value:  strings.EqualFold(text, "ACTIVE"),
				Raw:    text,
				Status: provider.StatusOk,
			}, nil
		},
		"api_reachable": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			text, err := describeText(ctx, cli,
				"eks", "describe-cluster",
				"--name", req.Name,
				"--query", "cluster.endpoint",
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
		// node_count is approximated as the number of managed nodegroups
		// to avoid kubectl round-trips; fleets using self-managed nodes
		// or Fargate profiles will under-report.
		"node_count": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			text, err := describeText(ctx, cli,
				"eks", "list-nodegroups",
				"--cluster-name", req.Name,
				"--query", "length(nodegroups)",
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
