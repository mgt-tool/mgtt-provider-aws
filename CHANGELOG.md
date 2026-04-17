# Changelog

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
