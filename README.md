# mgtt-provider-aws

AWS provider for [mgtt](https://github.com/mgt-tool/mgtt) — the model-guided troubleshooting tool.

Version **0.2.0** — built on the [mgtt provider SDK](https://github.com/mgt-tool/mgtt/tree/main/sdk/provider) (requires mgtt ≥ 0.1.0).

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

See [`types/rds_instance.yaml`](./types/rds_instance.yaml) for the state machine, healthy predicate, and failure-mode declarations.

Scope for v0.2.0 is intentionally narrow — one type, two facts — to get the packaging and image-install path right. Additional types (`ec2_instance`, `s3_bucket`, `lambda_function`, etc.) land in subsequent minor releases.

## Install

Two equivalent paths — pick whichever fits your workflow:

```bash
# Git + host toolchain (requires Go 1.25+, warns if aws CLI not on PATH)
mgtt provider install aws

# Pre-built Docker image (ships aws-cli inside; digest-pinned)
mgtt provider install --image ghcr.io/mgt-tool/mgtt-provider-aws:0.2.0@sha256:...
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

Probes are `aws rds describe-db-instances` and `aws cloudwatch get-metric-statistics` — **read-only**. `manifest.yaml` omits `read_only:` (which defaults to `true`), so `mgtt provider validate` emits no write-related warnings. Operators should scope credentials to a read-only policy anyway; minimum permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["rds:DescribeDBInstances", "cloudwatch:GetMetricStatistics"],
      "Resource": "*"
    }
  ]
}
```

## Architecture

- `main.go` — 13 lines: registers types and calls `provider.Main`.
- `internal/probes/rds_instance.go` — one function per fact; shells out to `aws` via the SDK's `shell.Client`.
- `internal/awsclassify/` — the **only** place aws-cli stderr phrasing (`DBInstanceNotFound`, `AccessDenied`, `Throttling`, …) is recognized; maps to the SDK's sentinel errors.

Plumbing (argv parsing, timeouts, size caps, `status:not_found` translation, exit codes, `version` subcommand) comes from the SDK.

## Development

```bash
go build .                                        # compile locally
go test ./...                                     # unit tests
./bin/mgtt-provider-aws probe checkout_db available --type rds_instance  # invoke directly
```

### Integration test (optional, requires AWS creds)

Not yet — added when the type surface expands past one type. Unit tests with a stubbed `shell.Client.Exec` cover the probe-protocol contract.

## See also

- [mgtt](https://github.com/mgt-tool/mgtt) — the constraint engine that consumes this provider
- [Image Capabilities](https://github.com/mgt-tool/mgtt/blob/main/docs/reference/image-capabilities.md) — how `needs` maps to `docker run` flags
