// Package awsclassify maps aws-cli stderr phrasing to the provider SDK's
// sentinel errors. This is the one place in the provider that encodes
// aws-specific vocabulary; every probe consumes the SDK's backend-agnostic
// shell.Client with this classifier plugged in.
package awsclassify

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/mgt-tool/mgtt/sdk/provider"
	"github.com/mgt-tool/mgtt/sdk/provider/shell"
)

// Classify is a shell.ClassifyFn for `aws`. Recognizes the common stderr
// shapes that AWS CLI v2 produces; anything else falls through to ErrEnv
// so the operator sees the raw message instead of a misclassified one.
func Classify(stderr string, runErr error) error {
	if runErr == nil {
		return nil
	}
	if errors.Is(runErr, exec.ErrNotFound) {
		return shell.EnvOnlyClassify(stderr, runErr)
	}
	first := firstLine(stderr)
	lower := strings.ToLower(stderr)
	switch {
	// Resource not present / identifier unknown. We match both the
	// AWS error code literal (stable across aws-cli phrasing changes)
	// and common lowercase natural-language substrings that AWS uses
	// in messages where the code is absent — notably S3 head-bucket,
	// which returns a bare HTTP status code instead of a named code.
	case strings.Contains(stderr, "DBInstanceNotFound"),
		strings.Contains(stderr, "NoSuchEntity"),
		strings.Contains(stderr, "NoSuchBucket"),
		strings.Contains(stderr, "NoSuchDistribution"),
		strings.Contains(stderr, "InvalidInstanceID.NotFound"),
		strings.Contains(stderr, "InvalidGroup.NotFound"),
		strings.Contains(stderr, "InvalidVpcID.NotFound"),
		strings.Contains(stderr, "InvalidSubnetID.NotFound"),
		strings.Contains(stderr, "InvalidVpcEndpointId.NotFound"),
		strings.Contains(stderr, "NatGatewayNotFound"),
		strings.Contains(stderr, "ResourceNotFoundException"),
		strings.Contains(stderr, "ParameterNotFound"),
		strings.Contains(stderr, "RepositoryNotFoundException"),
		strings.Contains(stderr, "ReplicationGroupNotFoundFault"),
		strings.Contains(stderr, "BrokerNotFoundException"),
		strings.Contains(stderr, "404) when calling the HeadBucket"),
		strings.Contains(lower, "could not be found"),
		strings.Contains(lower, "could not find"),
		strings.Contains(lower, "does not exist"),
		strings.Contains(lower, "not found"):
		return fmt.Errorf("%w: %s", provider.ErrNotFound, first)

	// Auth/permission failures. S3 head-bucket returns a bare HTTP
	// status code on 403; we match the shape.
	case strings.Contains(stderr, "AccessDenied"),
		strings.Contains(stderr, "UnauthorizedOperation"),
		strings.Contains(stderr, "InvalidClientTokenId"),
		strings.Contains(stderr, "SignatureDoesNotMatch"),
		strings.Contains(stderr, "ExpiredToken"),
		strings.Contains(stderr, "403) when calling the HeadBucket"),
		strings.Contains(lower, "credentials could not be found"),
		strings.Contains(lower, "unable to locate credentials"),
		strings.Contains(lower, "forbidden"):
		return fmt.Errorf("%w: %s", provider.ErrForbidden, first)

	// Network / endpoint / timeout / throttling — retryable class.
	case strings.Contains(stderr, "RequestLimitExceeded"),
		strings.Contains(stderr, "Throttling"),
		strings.Contains(stderr, "ServiceUnavailable"),
		strings.Contains(lower, "could not connect to the endpoint"),
		strings.Contains(lower, "connection reset"),
		strings.Contains(lower, "endpoint request timed out"),
		strings.Contains(lower, "context deadline exceeded"):
		return fmt.Errorf("%w: %s", provider.ErrTransient, first)
	}
	return fmt.Errorf("%w: %s", provider.ErrEnv, first)
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return strings.TrimSpace(s)
}
