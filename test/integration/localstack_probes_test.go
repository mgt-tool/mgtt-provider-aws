//go:build integration

package integration

// TestLocalStack_Probes runs real probe calls against a running LocalStack
// Community instance. Requires AWS_ENDPOINT_URL to point at LocalStack and
// fixtures seeded by test/integration/seed-localstack.sh (EC2-generated
// IDs are sourced from fixtures.env via the CI job).
//
// LocalStack CE does not emulate EKS, CloudFront, Amazon MQ, RDS,
// ElastiCache, or ECR (all Pro-only as of LocalStack 3.6). CloudWatch's
// get-metric-statistics is also unreliable on CE. Metric-backed facts and
// Pro-only types are intentionally absent from the sub-tests below; they
// stay on unit-test coverage.
//
// Run with:
//
//	bash test/integration/seed-localstack.sh
//	set -a && . test/integration/fixtures.env && set +a
//	go test -tags=integration -run TestLocalStack_Probes ./test/integration/... -v

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/mgt-tool/mgtt-provider-aws/internal/probes"
	"github.com/mgt-tool/mgtt/sdk/provider"
)

func TestLocalStack_Probes(t *testing.T) {
	if os.Getenv("AWS_ENDPOINT_URL") == "" {
		t.Skip("AWS_ENDPOINT_URL not set; skipping LocalStack integration")
	}
	if _, err := exec.LookPath("aws"); err != nil {
		t.Skip("aws CLI not on PATH")
	}

	vpcID := os.Getenv("FIXTURE_VPC_ID")
	sgID := os.Getenv("FIXTURE_SG_ID")
	natID := os.Getenv("FIXTURE_NAT_ID")
	endpointID := os.Getenv("FIXTURE_ENDPOINT_ID")
	acmArn := os.Getenv("FIXTURE_ACM_ARN")

	r := provider.NewRegistry()
	probes.Register(r)

	cases := []struct {
		name, typ, component, fact string
		rejectNotFound             bool
	}{
		{"s3_bucket accessible", "s3_bucket", "mgtt-test-bucket", "accessible", true},
		{"s3_bucket versioning_enabled", "s3_bucket", "mgtt-test-bucket", "versioning_enabled", true},
		{"iam_role exists", "iam_role", "mgtt-test-role", "exists", true},
		{"iam_role assumable", "iam_role", "mgtt-test-role", "assumable", true},
		{"iam_role attached_policy_count", "iam_role", "mgtt-test-role", "attached_policy_count", true},
		{"ssm_parameter exists", "ssm_parameter", "/mgtt-test/param", "exists", true},
		{"ssm_parameter version", "ssm_parameter", "/mgtt-test/param", "version", true},
		{"vpc available", "vpc", vpcID, "available", true},
		{"vpc ip_utilization", "vpc", vpcID, "ip_utilization", true},
		{"security_group exists", "security_group", sgID, "exists", true},
		{"security_group ingress_rule_count", "security_group", sgID, "ingress_rule_count", true},
		{"security_group egress_rule_count", "security_group", sgID, "egress_rule_count", true},
		{"nat_gateway available", "nat_gateway", natID, "available", true},
		{"vpc_endpoint available", "vpc_endpoint", endpointID, "available", true},
		{"vpc_endpoint dns_enabled", "vpc_endpoint", endpointID, "dns_enabled", true},
		{"acm_certificate issued", "acm_certificate", acmArn, "issued", true},
	}

	ctx := context.Background()
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.component == "" {
				t.Skipf("fixture id not populated for %s — seed script may have skipped it", tc.typ)
			}
			res, err := r.Probe(ctx, provider.Request{
				Type: tc.typ,
				Name: tc.component,
				Fact: tc.fact,
			})
			if err != nil {
				t.Fatalf("probe failed: %v", err)
			}
			if tc.rejectNotFound && res.Status == provider.StatusNotFound {
				t.Fatalf("%s resolved to not_found — seed fixture missing in LocalStack", tc.name)
			}
			t.Logf("%s = %v (raw=%q, status=%s)", tc.name, res.Value, res.Raw, res.Status)
		})
	}
}
