// Copyright (C) 2026 Alex Kunich
// SPDX-License-Identifier: Apache-2.0

package probes

import (
	"context"
	"strings"
	"time"

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
		// How much the queue grew over the fact's window: a deep queue that
		// drains is not a backlog, a shallow one climbing fast is.
		"queue_depth_delta_5m": provider.Windowed(func(ctx context.Context, req provider.Request, since time.Time) ([]provider.Sample, error) {
			if err := requireName(typ, req.Name); err != nil {
				return nil, err
			}
			return readCloudWatchSeries(ctx, cli, "AWS/AmazonMQ", "MessageCount", "Sum", since, []string{dim("Broker", req.Name)})
		}),
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
