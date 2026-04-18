package probes

import (
	"context"
	"strings"

	"github.com/mgt-tool/mgtt/sdk/provider"
	"github.com/mgt-tool/mgtt/sdk/provider/shell"
)

func registerNATGateway(r *provider.Registry, cli *shell.Client) {
	const typ = "nat_gateway"
	r.Register(typ, map[string]provider.ProbeFn{
		"available": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			text, err := describeText(ctx, cli,
				"ec2", "describe-nat-gateways",
				"--nat-gateway-ids", req.Name,
				"--query", "NatGateways[0].State",
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
		"error_port_allocation_count": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			f, err := readCloudWatchStatistic(ctx, cli,
				"AWS/NATGateway", "ErrorPortAllocation", "Sum",
				[]string{dim("NatGatewayId", req.Name)})
			if err != nil {
				return provider.Result{}, err
			}
			return provider.IntResult(int(f)), nil
		},
		"bytes_out_per_second": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			// BytesOutToDestination is a per-minute sum; divide by 60 for per-second rate.
			f, err := readCloudWatchStatistic(ctx, cli,
				"AWS/NATGateway", "BytesOutToDestination", "Sum",
				[]string{dim("NatGatewayId", req.Name)})
			if err != nil {
				return provider.Result{}, err
			}
			return provider.IntResult(int(f / 60)), nil
		},
	})
}
