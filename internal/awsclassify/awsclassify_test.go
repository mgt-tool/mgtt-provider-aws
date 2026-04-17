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
