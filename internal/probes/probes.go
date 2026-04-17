// Package probes wires the aws provider's per-fact probe functions to
// the SDK's Registry. Individual type implementations live in sibling
// files (rds_instance.go, etc.); this file is the single Register entry
// point consumed by main.go.
package probes

import (
	"github.com/mgt-tool/mgtt-provider-aws/internal/awsclassify"
	"github.com/mgt-tool/mgtt/sdk/provider"
	"github.com/mgt-tool/mgtt/sdk/provider/shell"
)

// newClient returns a shell.Client for `aws` with the aws-specific
// stderr classifier. Kept as a package-level helper so tests can swap
// shell.Client.Exec without constructing the whole CLI binary.
func newClient() *shell.Client {
	c := shell.New("aws")
	c.Classify = awsclassify.Classify
	return c
}

// Register wires every fact of every type.
func Register(r *provider.Registry) {
	registerRDSInstance(r, newClient())
}
