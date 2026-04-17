# mgtt-provider-aws

AWS provider for [mgtt](https://github.com/sajonaro/mgtt) — the model guided troubleshooting tool.

## Types

| Type | Description | Facts |
|------|-------------|-------|
| `rds_instance` | AWS RDS database instance | `available` (bool), `connection_count` (int) |

## Status

**Vocabulary-only provider.** All facts have inline probe commands using `aws-cli` — no compiled binary needed. mgtt executes the shell commands directly.

## Install

```bash
mgtt provider install aws
```

Requires `aws-cli` on `PATH` at probe time — facts are shell-fallback, so mgtt executes the `aws rds describe-db-instances …` commands in `provider.yaml` directly on the host.

### Docker image install is not yet supported

`mgtt provider install --image` expects the image's `ENTRYPOINT` to implement the probe protocol (argv: `probe <component> <fact> …`, JSON on stdout). This provider has no binary — it's a shell-fallback vocabulary only. An image would have nothing to invoke as entrypoint.

The clean path to image support is to port the shell-fallback `probe.cmd` commands into a tiny Go runner that uses the [mgtt provider SDK](https://github.com/mgt-tool/mgtt/tree/main/sdk/provider) and wraps `aws rds describe-db-instances` / `aws cloudwatch get-metric-statistics`. At that point the runner binary + `/provider.yaml` + an `aws-cli` layer compose into a standard image-installable provider, consistent with `mgtt-provider-kubernetes` and `mgtt-provider-tempo`. Tracked as a pre-v0.2.0 task.

## Auth

Uses your existing AWS credentials: `AWS_PROFILE`, `AWS_ACCESS_KEY_ID`+`AWS_SECRET_ACCESS_KEY`, `~/.aws/credentials`, or instance profile. Read-only AWS API access only.
