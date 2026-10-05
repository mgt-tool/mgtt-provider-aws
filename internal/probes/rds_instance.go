// Copyright (C) 2026 Alex Kunich
// SPDX-License-Identifier: Apache-2.0

package probes

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mgt-tool/mgtt/sdk/provider"
	"github.com/mgt-tool/mgtt/sdk/provider/shell"
)

// rds_instance facts query the AWS APIs via aws-cli. `available` maps the
// DBInstanceStatus string to a bool (anything other than "available" is
// considered unavailable); `connection_count` reads the most recent
// DatabaseConnections CloudWatch datapoint.
//
// We shell out to aws-cli rather than pulling in the Go SDK to keep this
// provider image-installable from any environment that has aws-cli on PATH
// (or inside an image that bundles it — see Dockerfile + image.needs).

const metricsWindow = 5 * time.Minute

func registerRDSInstance(r *provider.Registry, cli *shell.Client) {
	r.Register("rds_instance", map[string]provider.ProbeFn{
		"available": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			status, err := describeDBInstanceStatus(ctx, cli, req.Name)
			if err != nil {
				return provider.Result{}, err
			}
			ok := strings.EqualFold(status, "available")
			return provider.Result{
				Value:  ok,
				Raw:    status,
				Status: provider.StatusOk,
			}, nil
		},

		// The peak over the fact's window: a spike to the limit between two
		// one-minute snapshots still counts.
		"connection_count_max_5m": provider.Windowed(func(ctx context.Context, req provider.Request, since time.Time) ([]provider.Sample, error) {
			if err := requireName("rds_instance", req.Name); err != nil {
				return nil, err
			}
			return readCloudWatchSeries(ctx, cli, "AWS/RDS", "DatabaseConnections", "Maximum", since, []string{dim("DBInstanceIdentifier", req.Name)})
		}),
		"connection_count": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			count, err := describeDBConnectionCount(ctx, cli, req.Name)
			if err != nil {
				return provider.Result{}, err
			}
			return provider.IntResult(count), nil
		},
	})
}

func describeDBInstanceStatus(ctx context.Context, cli *shell.Client, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("%w: rds_instance probe requires a component name", provider.ErrUsage)
	}
	out, err := cli.Run(ctx,
		"rds", "describe-db-instances",
		"--db-instance-identifier", name,
		"--query", "DBInstances[0].DBInstanceStatus",
		"--output", "text")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func describeDBConnectionCount(ctx context.Context, cli *shell.Client, name string) (int, error) {
	if name == "" {
		return 0, fmt.Errorf("%w: rds_instance probe requires a component name", provider.ErrUsage)
	}
	f, err := readCloudWatchStatistic(ctx, cli, "AWS/RDS", "DatabaseConnections", "Maximum",
		[]string{dim("DBInstanceIdentifier", name)})
	return int(f), err
}
