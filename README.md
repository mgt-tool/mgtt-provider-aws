# mgtt-provider-aws

AWS provider for [mgtt](https://github.com/mgt-tool/mgtt) — the model-guided troubleshooting tool.

Version **1.0.0** — built on the [mgtt provider SDK](https://github.com/mgt-tool/mgtt/tree/main/sdk/provider) (requires mgtt ≥ 0.2.0).

```yaml
checkout_db:
  type: rds_instance
  providers: [aws]
```

When `mgtt plan` walks this component, the provider asks AWS: "is the RDS instance `checkout_db` available, and how many connections is it holding?" — using the answer to drive root-cause reasoning about your upstream services.

## Types

| Type | Facts |
|------|-------|
| `rds_instance` | `available` (bool), `connection_count` (int) |
| `elasticache_cluster` | `available` (bool), `connection_count` (int), `cache_hit_ratio` (%) |
| `mq_broker` | `available` (bool), `queue_depth` (int), `consumer_count` (int) |
| `s3_bucket` | `accessible` (bool), `versioning_enabled` (bool), `object_count` (int) |
| `eks_cluster` | `active` (bool), `api_reachable` (bool), `node_count` (int) |
| `ecr_repository` | `exists` (bool), `image_count` (int), `latest_push_age_seconds` (int) |
| `cloudfront_distribution` | `deployed` (bool), `enabled` (bool), `error_rate_5xx` (%) |
| `iam_role` | `exists` (bool), `assumable` (bool), `attached_policy_count` (int) |
| `acm_certificate` | `issued` (bool), `in_use` (bool), `days_until_expiry` (int) |
| `ssm_parameter` | `exists` (bool), `version` (int), `last_modified_age_seconds` (int) |
| `vpc` | `available` (bool), `ip_utilization` (%) |
| `nat_gateway` | `available` (bool), `error_port_allocation_count` (int), `bytes_out_per_second` (int) |
| `vpc_endpoint` | `available` (bool), `dns_enabled` (bool) |
| `security_group` | `exists` (bool), `ingress_rule_count` (int), `egress_rule_count` (int) |

Each type's YAML in [`types/`](./types/) declares the state machine, healthy predicate, and failure-mode declarations; `internal/probes/<type>.go` implements the facts by shelling out to `aws-cli`.

## Install

Two equivalent paths — pick whichever fits your workflow:

```bash
# Git + host toolchain (requires Go 1.25+, warns if aws CLI not on PATH)
mgtt provider install aws

# Pre-built Docker image (ships aws-cli inside; digest-pinned)
mgtt provider install --image ghcr.io/mgt-tool/mgtt-provider-aws:1.0.0@sha256:...
```

The image is published by [this repo's CI](./.github/workflows/docker.yml) on every push to `main` and every `v*` tag. Find the current digest on the [GHCR package page](https://github.com/mgt-tool/mgtt-provider-aws/pkgs/container/mgtt-provider-aws).

## Capabilities

When installed as an image, this provider declares the following runtime capabilities in [`manifest.yaml`](./manifest.yaml) (top-level `needs:`):

| Capability | Effect at probe time |
|---|---|
| `aws` | Mounts `~/.aws` read-only; forwards `AWS_PROFILE`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`, `AWS_REGION`, `AWS_DEFAULT_REGION` (whichever are set in the caller) |

Plus `network: host` so the container reaches AWS API endpoints without depending on bridge-network DNS.

On EC2/ECS with an IAM role (no on-disk credentials), drop the `aws` capability from `needs:` — the SDK inside the container resolves instance-profile credentials via the metadata service. See the [capability reference](https://github.com/mgt-tool/mgtt/blob/main/docs/reference/image-capabilities.md) for operator overrides and the `MGTT_IMAGE_CAPS_DENY` opt-out.

## Auth

Uses your existing AWS credential chain:

| Source | Used when |
|---|---|
| `AWS_PROFILE` | a named profile is selected |
| `AWS_ACCESS_KEY_ID` + `AWS_SECRET_ACCESS_KEY` (+ optional `AWS_SESSION_TOKEN`) | env-configured creds |
| `~/.aws/credentials`, `~/.aws/config` | file-configured creds |
| EC2 / ECS instance profile | running inside AWS |

All probes are **read-only** describes and CloudWatch reads. `manifest.yaml` omits `read_only:` (which defaults to `true`), so `mgtt provider validate` emits no write-related warnings. Operators should scope credentials to a read-only policy anyway; the union of all actions this provider uses:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "rds:DescribeDBInstances",
        "elasticache:DescribeReplicationGroups",
        "mq:DescribeBroker",
        "s3:ListBucket",
        "s3:GetBucketVersioning",
        "eks:DescribeCluster",
        "eks:ListNodegroups",
        "ecr:DescribeRepositories",
        "ecr:DescribeImages",
        "cloudfront:GetDistribution",
        "iam:GetRole",
        "iam:ListAttachedRolePolicies",
        "acm:DescribeCertificate",
        "ssm:GetParameter",
        "ssm:DescribeParameters",
        "ec2:DescribeVpcs",
        "ec2:DescribeSubnets",
        "ec2:DescribeNatGateways",
        "ec2:DescribeVpcEndpoints",
        "ec2:DescribeSecurityGroups",
        "cloudwatch:GetMetricStatistics"
      ],
      "Resource": "*"
    }
  ]
}
```

Types you don't use don't require their actions — trim the list to match the types referenced in your model.

## Architecture

- `main.go` — registers types and calls `provider.Main`.
- `internal/probes/<type>.go` — one file per type, one function per fact; shells out to `aws` via the SDK's `shell.Client`.
- `internal/probes/helpers.go` — shared name-guard and CloudWatch-statistic helper consumed by every probe that reads a metric.
- `internal/awsclassify/` — the **only** place aws-cli stderr phrasing (`DBInstanceNotFound`, `NoSuchEntity`, `AccessDenied`, `Throttling`, …) is recognized; maps to the SDK's sentinel errors.

Plumbing (argv parsing, timeouts, size caps, `status:not_found` translation, exit codes, `version` subcommand) comes from the SDK.

## Development

```bash
go build .                                        # compile locally
go test ./...                                     # unit tests
./bin/mgtt-provider-aws probe checkout_db available --type rds_instance  # invoke directly
```

### Integration tests

```bash
go test -tags=integration ./test/integration/...
```

- `TestImageInstall_Capabilities` builds the provider image, pushes it to a local registry, installs via `mgtt provider install --image`, and verifies capability/network declarations surface as expected. Requires `docker` and `mgtt` on PATH; skipped otherwise.
- `TestYAMLTypesAllRegistered` cross-checks `types/*.yaml` against the in-memory registry so a missing `register…` call in `internal/probes/probes.go` fails fast. Fast, no external deps.
- `TestLocalStack_Probes` runs every CE-supported probe against a live [LocalStack Community](https://github.com/localstack/localstack) instance:

  ```bash
  # Terminal 1 — start LocalStack
  docker run --rm -p 4566:4566 \
    -e SERVICES=s3,iam,ssm,ec2,acm \
    localstack/localstack:3.6

  # Terminal 2 — seed + run
  export AWS_ENDPOINT_URL=http://localhost:4566
  export AWS_REGION=us-east-1
  export AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test
  bash test/integration/seed-localstack.sh
  set -a && . test/integration/fixtures.env && set +a
  go test -tags=integration -run TestLocalStack_Probes ./test/integration/... -v
  ```

  Covers 8 types end-to-end: `s3_bucket`, `iam_role`, `ssm_parameter`, `acm_certificate`, `vpc`, `nat_gateway`, `vpc_endpoint`, `security_group`. The remaining 6 (`rds_instance`, `elasticache_cluster`, `ecr_repository`, `eks_cluster`, `cloudfront_distribution`, `mq_broker`) are Pro-only in LocalStack and CloudWatch metric endpoints are unreliable on CE — those stay on unit-test coverage.

  CI runs this job automatically on every push and PR — see `.github/workflows/ci.yml` job `integration-localstack`.

## See also

- [mgtt](https://github.com/mgt-tool/mgtt) — the constraint engine that consumes this provider
- [Image Capabilities](https://github.com/mgt-tool/mgtt/blob/main/docs/reference/image-capabilities.md) — how `needs` maps to `docker run` flags
