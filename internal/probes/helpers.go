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
// AWS/<namespace>/<metric> for <statistic>. aws-cli emits "None" when
// no datapoint is available in the window; we normalize that to 0.
func readCloudWatchStatistic(
	ctx context.Context,
	cli *shell.Client,
	namespace, metric, statistic string,
	dimensions []string,
) (float64, error) {
	now := time.Now().UTC()
	start := now.Add(-metricsWindow)
	args := []string{
		"cloudwatch", "get-metric-statistics",
		"--namespace", namespace,
		"--metric-name", metric,
		"--start-time", start.Format(time.RFC3339),
		"--end-time", now.Format(time.RFC3339),
		"--period", "60",
		"--statistics", statistic,
		"--query", "Datapoints[0]." + statistic,
		"--output", "text",
	}
	if len(dimensions) > 0 {
		args = append(args, "--dimensions")
		args = append(args, dimensions...)
	}
	out, err := cli.Run(ctx, args...)
	if err != nil {
		return 0, err
	}
	return parseFloatOrZero(strings.TrimSpace(string(out)))
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
