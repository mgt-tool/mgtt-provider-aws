package probes

import (
	"context"
	"strings"

	"github.com/mgt-tool/mgtt/sdk/provider"
	"github.com/mgt-tool/mgtt/sdk/provider/shell"
)

func registerCloudFrontDistribution(r *provider.Registry, cli *shell.Client) {
	const typ = "cloudfront_distribution"
	r.Register(typ, map[string]provider.ProbeFn{
		"deployed": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			text, err := describeText(ctx, cli,
				"cloudfront", "get-distribution",
				"--id", req.Name,
				"--query", "Distribution.Status",
				"--output", "text")
			if err != nil {
				return provider.Result{}, err
			}
			return provider.Result{
				Value:  strings.EqualFold(text, "Deployed"),
				Raw:    text,
				Status: provider.StatusOk,
			}, nil
		},
		"enabled": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			text, err := describeText(ctx, cli,
				"cloudfront", "get-distribution",
				"--id", req.Name,
				"--query", "Distribution.DistributionConfig.Enabled",
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
		"error_rate_5xx": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			// CloudFront's Region dimension is always "Global" (the service
			// aggregates across all edge locations into a single series).
			f, err := readCloudWatchStatistic(ctx, cli,
				"AWS/CloudFront", "5xxErrorRate", "Average",
				[]string{dim("DistributionId", req.Name), dim("Region", "Global")})
			if err != nil {
				return provider.Result{}, err
			}
			return provider.FloatResult(f), nil
		},
	})
}
