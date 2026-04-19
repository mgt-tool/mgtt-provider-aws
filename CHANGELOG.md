# Changelog

## [Unreleased]

### Added

- **Per-type `requires.iam:` block** in every `types/*.yaml` file listing exactly which AWS API actions that type's probes touch, along with the resource-ARN pattern the action supports (or `*` with a `note:` where AWS doesn't support resource-scoping). Operators can now scope a probe-role policy per type instead of granting the union-of-all-types. mgtt-core doesn't parse the block yet — it's documentation today; a follow-up will render it automatically via `mgtt provider inspect`.

## [1.0.0] — 2026-04-18

### Changed (breaking)

- `manifest.yaml` migrated to the v1.0 mgtt schema: three top-level blocks (`meta`, `runtime`, `install`); `hooks:` retired; `needs:` + `network:` moved under `runtime:`; install methods declared via `install.source` + `install.image` subblocks. Requires mgtt ≥ 0.2.0.

## [0.3.0] — 2026-04-18

### Added

- **13 new types** derived from surveying the `magento-infrastructure` Terraform project and applying the mgtt type philosophy (runtime-observable, with facts / states / failure modes):
  - `elasticache_cluster`, `mq_broker`, `s3_bucket`, `eks_cluster`, `ecr_repository`, `cloudfront_distribution`, `iam_role`, `acm_certificate`, `ssm_parameter`, `vpc`, `nat_gateway`, `vpc_endpoint`, `security_group`.
- **Go probe implementations** (`internal/probes/<type>.go`, one file per type) — each fact shells out to `aws-cli` via the SDK's `shell.Client` and propagates errors so `awsclassify` can translate them to SDK sentinels.
- **Shared helpers** (`internal/probes/helpers.go`) — `requireName`, `readCloudWatchStatistic`, CIDR/float parsing — consolidate the common probe plumbing.
- **Unit tests** for every new type under `internal/probes/*_test.go` — happy path + missing-name + not-found handling where applicable, using the SDK's stubbable `shell.Client.Exec`.
- **Extended IAM policy in README** covering the union of describes/reads for all 14 types; operators are directed to trim to the types they actually reference.
- **`TestYAMLTypesAllRegistered` integration test** — cross-checks `types/*.yaml` against the in-memory registry so a missing `register…` call fails fast.

### Changed

- `meta.version` and image tag bumped to `0.3.0`.
- README Types table expanded from 1 row to 14; "Scope intentionally narrow" paragraph removed.

### Scope notes

The new probes prioritize correctness of the declarative model (YAML types, state machines, error classification) over exhaustive per-fact tuning:

- `eks_cluster.node_count` is approximated as the number of managed nodegroups (avoids kubectl round-trips; self-managed / Fargate fleets under-report).
- `iam_role.assumable` checks whether a trust policy document exists, not whether the current caller can actually assume the role — an `sts:AssumeRole` call would be a write-ish side effect we avoid in probes.
- `cloudfront_distribution.error_rate_5xx` reads the `Region=Global` CloudFront metric (CloudFront's region dimension is fixed by the service).

All three are documented at the call site and can be tightened in a subsequent minor release without a type-surface change.

## [0.2.0] — 2026-04-17

### Added

- **Proper Go binary provider.** Replaces the v0.1.0 shell-fallback stub. `mgtt-provider-aws` is now a standalone binary built on the [mgtt provider SDK](https://github.com/mgt-tool/mgtt/tree/main/sdk/provider), consistent with mgtt-provider-kubernetes / tempo / quickwit / terraform.
- **Dockerfile + CI publish.** Image based on `amazon/aws-cli:2.17.0` with the provider binary grafted in; `.github/workflows/docker.yml` publishes to `ghcr.io/mgt-tool/mgtt-provider-aws` on every push to `main` and every `v*` tag.
- **`image.needs: [aws, network]`** in `manifest.yaml`. mgtt forwards `~/.aws` as a read-only bind mount, the `AWS_*` env chain, and `--network host` at probe time — `mgtt provider install --image` now works end-to-end.
- **Backend-specific error classification** (`internal/awsclassify`). Maps `DBInstanceNotFound`, `AccessDenied`, `Throttling`, etc. to the SDK's sentinel errors so `mgtt plan` reasons correctly about missing resources, auth failures, and retryable conditions.
- **Uninstall hook** (`hooks/uninstall.sh`). Cleans up `bin/` so `mgtt provider uninstall aws` leaves nothing behind.

### Changed

- `manifest.yaml` no longer carries inline `probe.cmd` strings; probes live in `internal/probes/rds_instance.go` and share the SDK's `shell.Client` for timeout/size-cap/classification.
- `meta.command` is set to `$MGTT_PROVIDER_DIR/bin/mgtt-provider-aws`; git install builds the binary via `hooks/install.sh`.
- `meta.version` bumped to `0.2.0`.

### Scope

Same facts as v0.1.0 — `rds_instance.available` and `rds_instance.connection_count`. Type-surface expansion (ec2_instance, s3_bucket, lambda_function, etc.) will happen in subsequent minor releases; this release is the packaging rewrite, not a feature expansion.

## [0.1.0] — 2026-04-12

Initial release. Shell-fallback vocab-only provider covering `rds_instance` with two facts (`available`, `connection_count`) invoked via `aws-cli` at probe time.
