package awsclassify

import (
	"errors"
	"testing"

	"github.com/mgt-tool/mgtt/sdk/provider"
)

func TestClassify(t *testing.T) {
	runErr := errors.New("exit status 254")

	cases := []struct {
		name    string
		stderr  string
		wantIs  error
		wantSub string
	}{
		{"rds not found", "An error occurred (DBInstanceNotFound) when calling the DescribeDBInstances operation: DBInstance x not found.\n", provider.ErrNotFound, "DBInstanceNotFound"},
		{"ec2 not found", "An error occurred (InvalidInstanceID.NotFound) when calling the DescribeInstances operation: The instance ID 'i-x' does not exist\n", provider.ErrNotFound, "InvalidInstanceID.NotFound"},
		{"iam role not found", "An error occurred (NoSuchEntity) when calling the GetRole operation: Role r not found\n", provider.ErrNotFound, "NoSuchEntity"},
		{"ssm parameter not found", "An error occurred (ParameterNotFound) when calling the GetParameter operation: Parameter /x not found.\n", provider.ErrNotFound, "ParameterNotFound"},
		{"ecr repo not found", "An error occurred (RepositoryNotFoundException) when calling the DescribeRepositories operation: The repository with name 'r' does not exist in the registry with id '1'\n", provider.ErrNotFound, "RepositoryNotFoundException"},
		{"security group not found", "An error occurred (InvalidGroup.NotFound) when calling the DescribeSecurityGroups operation: The security group 'sg-x' does not exist\n", provider.ErrNotFound, "InvalidGroup.NotFound"},
		{"vpc not found", "An error occurred (InvalidVpcID.NotFound) when calling the DescribeVpcs operation: The vpc ID 'vpc-x' does not exist\n", provider.ErrNotFound, "InvalidVpcID.NotFound"},
		{"vpc endpoint not found", "An error occurred (InvalidVpcEndpointId.NotFound) when calling the DescribeVpcEndpoints operation: The vpcEndpoint ID 'vpce-x' does not exist\n", provider.ErrNotFound, "InvalidVpcEndpointId.NotFound"},
		{"nat gateway not found", "An error occurred (NatGatewayNotFound) when calling the DescribeNatGateways operation: The Nat Gateway 'nat-x' does not exist\n", provider.ErrNotFound, "NatGatewayNotFound"},
		{"eks cluster not found", "An error occurred (ResourceNotFoundException) when calling the DescribeCluster operation: No cluster found for name: prod.\n", provider.ErrNotFound, "ResourceNotFoundException"},
		{"acm cert not found", "An error occurred (ResourceNotFoundException) when calling the DescribeCertificate operation: Could not find certificate for arn arn:aws:acm:...:/x\n", provider.ErrNotFound, "ResourceNotFoundException"},
		{"elasticache rg not found", "An error occurred (ReplicationGroupNotFoundFault) when calling the DescribeReplicationGroups operation: Replication group c not found.\n", provider.ErrNotFound, "ReplicationGroupNotFoundFault"},
		{"mq broker not found", "An error occurred (BrokerNotFoundException) when calling the DescribeBroker operation: Can't find requested broker b-x.\n", provider.ErrNotFound, "BrokerNotFoundException"},
		{"cloudfront dist not found", "An error occurred (NoSuchDistribution) when calling the GetDistribution operation: The specified distribution does not exist.\n", provider.ErrNotFound, "NoSuchDistribution"},
		{"s3 head-bucket 404", "An error occurred (404) when calling the HeadBucket operation: Not Found\n", provider.ErrNotFound, "HeadBucket"},
		{"s3 head-bucket 403", "An error occurred (403) when calling the HeadBucket operation: Forbidden\n", provider.ErrForbidden, "HeadBucket"},
		{"access denied", "An error occurred (AccessDenied) when calling the DescribeDBInstances operation: User: arn:aws:iam::x:user/y is not authorized\n", provider.ErrForbidden, "AccessDenied"},
		{"expired token", "An error occurred (ExpiredToken) when calling the DescribeDBInstances operation\n", provider.ErrForbidden, "ExpiredToken"},
		{"unable to locate credentials", "Unable to locate credentials. You can configure credentials by running \"aws configure\".\n", provider.ErrForbidden, "locate credentials"},
		{"throttling", "An error occurred (Throttling) when calling the ListMetrics operation: Rate exceeded\n", provider.ErrTransient, "Throttling"},
		{"endpoint timeout", "Connect timeout on endpoint URL: https://rds.us-east-1.amazonaws.com/\n", provider.ErrEnv, ""}, // endpoint phrasing varies; this one lands in ErrEnv
		{"endpoint connection", "Could not connect to the endpoint URL: https://rds.us-east-1.amazonaws.com/\n", provider.ErrTransient, "Could not connect"},
		{"unknown", "An error occurred (WeirdThing) when calling the DescribeDBInstances operation\n", provider.ErrEnv, "WeirdThing"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := Classify(tc.stderr, runErr)
			if !errors.Is(err, tc.wantIs) {
				t.Fatalf("want errors.Is(%v), got %v", tc.wantIs, err)
			}
			if tc.wantSub != "" && !contains(err.Error(), tc.wantSub) {
				t.Errorf("want err.Error() to contain %q; got %q", tc.wantSub, err.Error())
			}
		})
	}
}

func TestClassify_NilRunError(t *testing.T) {
	if err := Classify("", nil); err != nil {
		t.Fatalf("nil runErr must yield nil; got %v", err)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
