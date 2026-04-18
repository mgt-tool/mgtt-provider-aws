package probes

import (
	"context"
	"strings"

	"github.com/mgt-tool/mgtt/sdk/provider"
	"github.com/mgt-tool/mgtt/sdk/provider/shell"
)

func registerMQBroker(r *provider.Registry, cli *shell.Client) {
	const typ = "mq_broker"
	r.Register(typ, map[string]provider.ProbeFn{
		"available": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			text, err := describeText(ctx, cli,
				"mq", "describe-broker",
				"--broker-id", req.Name,
				"--query", "BrokerState",
				"--output", "text")
			if err != nil {
				return provider.Result{}, err
			}
			return provider.Result{
				Value:  strings.EqualFold(text, "RUNNING"),
				Raw:    text,
				Status: provider.StatusOk,
			}, nil
		},
		"queue_depth": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			f, err := readCloudWatchStatistic(ctx, cli,
				"AWS/AmazonMQ", "MessageCount", "Sum",
				[]string{dim("Broker", req.Name)})
			if err != nil {
				return provider.Result{}, err
			}
			return provider.IntResult(int(f)), nil
		},
		"consumer_count": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			f, err := readCloudWatchStatistic(ctx, cli,
				"AWS/AmazonMQ", "ConsumerCount", "Sum",
				[]string{dim("Broker", req.Name)})
			if err != nil {
				return provider.Result{}, err
			}
			return provider.IntResult(int(f)), nil
		},
	})
}
