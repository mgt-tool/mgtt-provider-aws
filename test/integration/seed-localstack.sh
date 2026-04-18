#!/usr/bin/env bash
# Seeds fixtures in a LocalStack Community instance for integration tests.
# Idempotent: existing resources are tolerated (ign helper swallows errors).
# EC2-returned IDs are written to $FIXTURES_FILE for the Go tests to source.
#
# Required env:
#   AWS_ENDPOINT_URL   http://localhost:4566 when LocalStack is on the runner
#   AWS_REGION         defaults to us-east-1
#   AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY  default to "test"
#
# LocalStack CE does not emulate EKS, CloudFront, Amazon MQ, RDS, ElastiCache,
# or ECR (all Pro-only as of LocalStack 3.6) — those types have no fixture
# here; their integration coverage stays on unit tests. CloudWatch metric
# endpoints are also unreliable on CE — metric-backed facts are excluded from
# the Go integration test by design.
set -euo pipefail

: "${AWS_ENDPOINT_URL:?AWS_ENDPOINT_URL must point at LocalStack}"
export AWS_REGION="${AWS_REGION:-us-east-1}"
export AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID:-test}"
export AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY:-test}"

FIXTURES_FILE="${FIXTURES_FILE:-$(pwd)/fixtures.env}"
: > "$FIXTURES_FILE"

log() { printf '[seed] %s\n' "$*" >&2; }
ign() { "$@" >/dev/null 2>&1 || true; }

log "S3 bucket"
ign aws s3api create-bucket --bucket mgtt-test-bucket
ign aws s3api put-bucket-versioning \
  --bucket mgtt-test-bucket \
  --versioning-configuration Status=Enabled

log "IAM role"
ign aws iam create-role \
  --role-name mgtt-test-role \
  --assume-role-policy-document \
  '{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":"ec2.amazonaws.com"},"Action":"sts:AssumeRole"}]}'

log "SSM parameter"
ign aws ssm put-parameter \
  --name /mgtt-test/param \
  --value hello \
  --type String \
  --overwrite

log "ACM certificate"
acm_arn="$(aws acm request-certificate \
  --domain-name example.test \
  --validation-method DNS \
  --query CertificateArn \
  --output text 2>/dev/null || true)"
if [[ -n "${acm_arn}" && "${acm_arn}" != "None" ]]; then
  echo "FIXTURE_ACM_ARN=${acm_arn}" >> "$FIXTURES_FILE"
fi

log "VPC + subnet"
vpc_id="$(aws ec2 create-vpc --cidr-block 10.100.0.0/16 --query Vpc.VpcId --output text)"
echo "FIXTURE_VPC_ID=${vpc_id}" >> "$FIXTURES_FILE"
subnet_id="$(aws ec2 create-subnet \
  --vpc-id "$vpc_id" \
  --cidr-block 10.100.1.0/24 \
  --query Subnet.SubnetId \
  --output text)"
echo "FIXTURE_SUBNET_ID=${subnet_id}" >> "$FIXTURES_FILE"

log "Security group"
sg_id="$(aws ec2 create-security-group \
  --vpc-id "$vpc_id" \
  --group-name mgtt-test-sg \
  --description 'mgtt test' \
  --query GroupId \
  --output text)"
echo "FIXTURE_SG_ID=${sg_id}" >> "$FIXTURES_FILE"
ign aws ec2 authorize-security-group-ingress \
  --group-id "$sg_id" \
  --protocol tcp \
  --port 443 \
  --cidr 0.0.0.0/0

log "NAT gateway"
eip_alloc="$(aws ec2 allocate-address --domain vpc --query AllocationId --output text)"
nat_id="$(aws ec2 create-nat-gateway \
  --subnet-id "$subnet_id" \
  --allocation-id "$eip_alloc" \
  --query NatGateway.NatGatewayId \
  --output text)"
echo "FIXTURE_NAT_ID=${nat_id}" >> "$FIXTURES_FILE"

log "VPC endpoint (S3 gateway — gateway endpoints don't support PrivateDnsEnabled; dns_enabled will read False)"
endpoint_id="$(aws ec2 create-vpc-endpoint \
  --vpc-id "$vpc_id" \
  --service-name "com.amazonaws.${AWS_REGION}.s3" \
  --query VpcEndpoint.VpcEndpointId \
  --output text)"
echo "FIXTURE_ENDPOINT_ID=${endpoint_id}" >> "$FIXTURES_FILE"

log "Fixtures ready"
cat "$FIXTURES_FILE"
