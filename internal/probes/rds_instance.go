package probes

import (
	"context"
	"fmt"
	"strconv"
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
	now := time.Now().UTC()
	start := now.Add(-metricsWindow)
	out, err := cli.Run(ctx,
		"cloudwatch", "get-metric-statistics",
		"--namespace", "AWS/RDS",
		"--metric-name", "DatabaseConnections",
		"--dimensions", "Name=DBInstanceIdentifier,Value="+name,
		"--start-time", start.Format(time.RFC3339),
		"--end-time", now.Format(time.RFC3339),
		"--period", "60",
		"--statistics", "Maximum",
		"--query", "Datapoints[0].Maximum",
		"--output", "text")
	if err != nil {
		return 0, err
	}
	text := strings.TrimSpace(string(out))
	// aws-cli returns the literal "None" when the Datapoints array is empty
	// (e.g. the metric hasn't been emitted in the last 5 minutes). Treat
	// that as zero connections rather than a parse failure — it's the
	// correct interpretation for the fact name.
	if text == "" || text == "None" {
		return 0, nil
	}
	// The output is a float (e.g. "42.0"); convert via strconv + int().
	f, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: unexpected aws output %q", provider.ErrProtocol, text)
	}
	return int(f), nil
}
