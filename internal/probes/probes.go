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
// stderr classifier. Tests in this package construct their own client
// via fakeClient (rds_instance_test.go) rather than swapping here.
func newClient() *shell.Client {
	c := shell.New("aws")
	c.Classify = awsclassify.Classify
	return c
}

// Register wires every fact of every type.
func Register(r *provider.Registry) {
	cli := newClient()
	registerRDSInstance(r, cli)
	registerElasticacheCluster(r, cli)
	registerMQBroker(r, cli)
	registerS3Bucket(r, cli)
	registerEKSCluster(r, cli)
	registerECRRepository(r, cli)
	registerCloudFrontDistribution(r, cli)
	registerIAMRole(r, cli)
	registerACMCertificate(r, cli)
	registerSSMParameter(r, cli)
	registerVPC(r, cli)
	registerNATGateway(r, cli)
	registerVPCEndpoint(r, cli)
	registerSecurityGroup(r, cli)
}
