package utils

import (
	"github.com/aws/aws-sdk-go/aws/endpoints"
)

// dualStackDNSSuffixes maps an AWS partition ID to the DNS suffix used by the
// dual-stack endpoints of that partition. It only contains the partitions in
// which every AWS service used by the operator (CloudFormation, EC2, EKS, IAM
// and STS) publishes dual-stack endpoints.
//
// The AWS endpoint rules treat dual-stack support as a partition wide property,
// so asking for dual-stack endpoints in a partition where a single service does
// not publish them resolves to a host that does not exist. This is the case for
// EC2 in the aws-cn partition, where only ec2.<region>.amazonaws.com.cn exists
// and ec2.<region>.api.amazonwebservices.com.cn does not.
var dualStackDNSSuffixes = map[string]string{
	endpoints.AwsPartitionID: "api.aws",
}

// AWSPartitionForRegion returns the AWS partition the given region belongs to,
// defaulting to the standard AWS partition for empty or unknown regions.
func AWSPartitionForRegion(region string) endpoints.Partition {
	if p, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), region); ok {
		return p
	}
	return endpoints.AwsPartition()
}

// AWSDNSSuffix returns the DNS suffix of the partition the given region belongs to.
func AWSDNSSuffix(region string) string {
	return AWSPartitionForRegion(region).DNSSuffix()
}

// AWSARNPrefix returns the ARN prefix of the partition the given region belongs to.
func AWSARNPrefix(region string) string {
	return "arn:" + AWSPartitionForRegion(region).ID()
}

// SupportsDualStackEndpoints returns true when all the AWS services used by the
// operator publish dual-stack endpoints in the partition the given region
// belongs to.
func SupportsDualStackEndpoints(region string) bool {
	_, found := dualStackDNSSuffixes[AWSPartitionForRegion(region).ID()]
	return found
}

// DualStackDNSSuffix returns the DNS suffix used by the dual-stack endpoints of
// the partition the given region belongs to. It returns an empty string for the
// partitions without usable dual-stack endpoints.
func DualStackDNSSuffix(region string) string {
	return dualStackDNSSuffixes[AWSPartitionForRegion(region).ID()]
}
