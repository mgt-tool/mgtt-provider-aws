package probes

import (
	"context"
	"strings"
	"time"

	"github.com/mgt-tool/mgtt/sdk/provider"
	"github.com/mgt-tool/mgtt/sdk/provider/shell"
)

func registerACMCertificate(r *provider.Registry, cli *shell.Client) {
	const typ = "acm_certificate"
	r.Register(typ, map[string]provider.ProbeFn{
		"issued": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			text, err := describeText(ctx, cli,
				"acm", "describe-certificate",
				"--certificate-arn", req.Name,
				"--query", "Certificate.Status",
				"--output", "text")
			if err != nil {
				return provider.Result{}, err
			}
			return provider.Result{
				Value:  strings.EqualFold(text, "ISSUED"),
				Raw:    text,
				Status: provider.StatusOk,
			}, nil
		},
		"in_use": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			text, err := describeText(ctx, cli,
				"acm", "describe-certificate",
				"--certificate-arn", req.Name,
				"--query", "length(Certificate.InUseBy)",
				"--output", "text")
			if err != nil {
				return provider.Result{}, err
			}
			n, err := parseIntOrZero(text)
			if err != nil {
				return provider.Result{}, err
			}
			return provider.BoolResult(n > 0), nil
		},
		"days_until_expiry": func(ctx context.Context, req provider.Request) (provider.Result, error) {
			if err := requireName(typ, req.Name); err != nil {
				return provider.Result{}, err
			}
			text, err := describeText(ctx, cli,
				"acm", "describe-certificate",
				"--certificate-arn", req.Name,
				"--query", "Certificate.NotAfter",
				"--output", "text")
			if err != nil {
				return provider.Result{}, err
			}
			t, err := parseAWSTimestamp(text)
			if err != nil {
				return provider.Result{}, err
			}
			if t.IsZero() {
				return provider.IntResult(0), nil
			}
			days := int(time.Until(t).Hours() / 24)
			return provider.IntResult(days), nil
		},
	})
}
