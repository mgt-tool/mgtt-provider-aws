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

## Auth

Uses your existing AWS credentials: `AWS_PROFILE`, `AWS_ACCESS_KEY_ID`+`AWS_SECRET_ACCESS_KEY`, `~/.aws/credentials`, or instance profile. Read-only AWS API access only.
