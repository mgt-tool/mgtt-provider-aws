// Copyright (C) 2026 Alex Kunich
// SPDX-License-Identifier: Apache-2.0

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

func requireName(typ, name string) error {
	if name == "" {
		return fmt.Errorf("%w: %s probe requires a component name", provider.ErrUsage, typ)
	}
	return nil
}

func describeText(ctx context.Context, cli *shell.Client, args ...string) (string, error) {
	out, err := cli.Run(ctx, args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func dim(name, value string) string {
	return fmt.Sprintf("Name=%s,Value=%s", name, value)
}

// readCloudWatchStatistic returns the most recent datapoint of
// AWS/<namespace>/<metric> for <statistic> in the last metricsWindow.
// CloudWatch returns datapoints in no particular order, so every one is
// read and the newest kept; reading the first, as this did, could report
// any minute of the window. No datapoint is 0, as aws-cli's "None" was.
func readCloudWatchStatistic(
	ctx context.Context,
	cli *shell.Client,
	namespace, metric, statistic string,
	dimensions []string,
) (float64, error) {
	samples, err := readCloudWatchSeries(ctx, cli, namespace, metric, statistic, time.Now().Add(-metricsWindow), dimensions)
	if err != nil || len(samples) == 0 {
		return 0, err
	}
	latest := samples[0]
	for _, s := range samples[1:] {
		if s.At.After(latest.At) {
			latest = s
		}
	}
	return latest.Value, nil
}

func parseFloatOrZero(text string) (float64, error) {
	if text == "" || text == "None" {
		return 0, nil
	}
	f, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: unexpected aws output %q", provider.ErrProtocol, text)
	}
	return f, nil
}

func parseIntOrZero(text string) (int, error) {
	f, err := parseFloatOrZero(text)
	if err != nil {
		return 0, err
	}
	return int(f), nil
}

// parseAWSTimestamp handles both of aws-cli's timestamp rendering modes:
// the default ISO8601 string (`cli_timestamp_format=iso8601`, which is the
// v2 default) and the legacy numeric epoch-seconds rendering some older
// versions or configs produce. Returns (zero time, nil) for the empty or
// "None" inputs so callers can treat "no datapoint" as a concrete zero.
func parseAWSTimestamp(text string) (time.Time, error) {
	if text == "" || text == "None" {
		return time.Time{}, nil
	}
	if f, err := strconv.ParseFloat(text, 64); err == nil {
		return time.Unix(int64(f), 0), nil
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		if t, err := time.Parse(layout, text); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("%w: unrecognised aws timestamp %q", provider.ErrProtocol, text)
}

// readCloudWatchSeries reads every one-minute datapoint of
// AWS/<namespace>/<metric> for <statistic> since a time, for a derived
// fact: provider.Windowed reduces them as the fact spec says. No
// datapoints is no samples, which the SDK reports as unknown, not zero.
func readCloudWatchSeries(
	ctx context.Context,
	cli *shell.Client,
	namespace, metric, statistic string,
	since time.Time,
	dimensions []string,
) ([]provider.Sample, error) {
	args := []string{
		"cloudwatch", "get-metric-statistics",
		"--namespace", namespace,
		"--metric-name", metric,
		"--start-time", since.UTC().Format(time.RFC3339),
		"--end-time", time.Now().UTC().Format(time.RFC3339),
		"--period", "60",
		"--statistics", statistic,
		"--query", "Datapoints[*].[Timestamp," + statistic + "]",
		"--output", "text",
	}
	if len(dimensions) > 0 {
		args = append(args, "--dimensions")
		args = append(args, dimensions...)
	}
	out, err := cli.Run(ctx, args...)
	if err != nil {
		return nil, err
	}
	var samples []provider.Sample
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || fields[0] == "None" {
			continue
		}
		if len(fields) != 2 {
			return nil, fmt.Errorf("%w: unexpected aws output line %q", provider.ErrProtocol, line)
		}
		at, err := time.Parse(time.RFC3339, fields[0])
		if err != nil {
			return nil, fmt.Errorf("%w: datapoint timestamp %q: %v", provider.ErrProtocol, fields[0], err)
		}
		v, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			return nil, fmt.Errorf("%w: datapoint value %q", provider.ErrProtocol, fields[1])
		}
		samples = append(samples, provider.Sample{At: at, Value: v})
	}
	return samples, nil
}
