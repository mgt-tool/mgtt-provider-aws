package probes

import (
	"context"
	"strings"

	"github.com/mgt-tool/mgtt/sdk/provider"
	"github.com/mgt-tool/mgtt/sdk/provider/shell"
)

func registerElasticacheCluster(r *provider.Registry, cli *shell.Client) {
	const typ = "elasticache_cluster"
	r.Register(typ, map[string]provider.ProbeFn{
		"available": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			text, err := describeText(ctx, cli,
				"elasticache", "describe-replication-groups",
				"--replication-group-id", req.Name,
				"--query", "ReplicationGroups[0].Status",
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
		"connection_count": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			f, err := readCloudWatchStatistic(ctx, cli,
				"AWS/ElastiCache", "CurrConnections", "Sum",
				[]string{dim("CacheClusterId", req.Name)})
			if err != nil {
				return provider.Result{}, err
			}
			return provider.IntResult(int(f)), nil
		},
		"cache_hit_ratio": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			dims := []string{dim("CacheClusterId", req.Name)}
			hits, err := readCloudWatchStatistic(ctx, cli, "AWS/ElastiCache", "CacheHits", "Sum", dims)
			if err != nil {
				return provider.Result{}, err
			}
			misses, err := readCloudWatchStatistic(ctx, cli, "AWS/ElastiCache", "CacheMisses", "Sum", dims)
			if err != nil {
				return provider.Result{}, err
			}
			total := hits + misses
			if total == 0 {
				return provider.FloatResult(0), nil
			}
			return provider.FloatResult(hits / total * 100), nil
		},
	})
}
