package controller

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	eksv1 "github.com/rancher/eks-operator/pkg/apis/eks.cattle.io/v1"
	"github.com/rancher/eks-operator/pkg/eks/services"
	"github.com/rancher/eks-operator/utils"
	wranglerv1 "github.com/rancher/wrangler/v3/pkg/generated/controllers/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newAWSConfigV2(ctx context.Context, secretClient wranglerv1.SecretClient, spec eksv1.EKSClusterConfigSpec) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithUseDualStackEndpoint(aws.DualStackEndpointStateEnabled))
	if err != nil {
		return cfg, fmt.Errorf("error loading default AWS config: %w", err)
	}

	if region := spec.Region; region != "" {
		cfg.Region = region
	}

	if amazonCredentialSecret := spec.AmazonCredentialSecret; amazonCredentialSecret != "" {
		ns, id := utils.Parse(spec.AmazonCredentialSecret)
		secret, err := secretClient.Get(ns, id, metav1.GetOptions{})
		if err != nil {
			return cfg, fmt.Errorf("error getting secret %s/%s: %w", ns, id, err)
		}

		// Opt-in to the AWS SDK default credential chain (IAM instance profile,
		// IRSA, EKS Pod Identity, ECS task role, env vars, etc.) when the secret
		// explicitly requests it via the useInstanceProfile flag. In that mode
		// any stray accessKey/secretKey fields are ignored.
		if useInstanceProfile := string(secret.Data["amazonec2credentialConfig-useInstanceProfile"]); useInstanceProfile == "true" {
			return cfg, nil
		}

		accessKey := string(secret.Data["amazonec2credentialConfig-accessKey"])
		secretKey := string(secret.Data["amazonec2credentialConfig-secretKey"])

		// Both fields empty/absent is also treated as opting into the default
		// credential chain. This keeps the behavior consistent with an empty
		// AmazonCredentialSecret reference and lets Rancher store a credential
		// resource that only carries metadata (e.g. region) without static keys.
		if accessKey == "" && secretKey == "" {
			return cfg, nil
		}

		// Exactly one of the two fields populated is a misconfiguration.
		if accessKey == "" || secretKey == "" {
			return cfg, fmt.Errorf("invalid aws cloud credential")
		}

		cfg.Credentials = credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")
	}

	return cfg, nil
}

func newAWSv2Services(ctx context.Context, secretClient wranglerv1.SecretClient, spec eksv1.EKSClusterConfigSpec) (*awsServices, error) {
	cfg, err := newAWSConfigV2(ctx, secretClient, spec)
	if err != nil {
		return nil, err
	}

	return &awsServices{
		eks:            services.NewEKSService(cfg),
		cloudformation: services.NewCloudFormationService(cfg),
		iam:            services.NewIAMService(cfg),
		ec2:            services.NewEC2Service(cfg),
	}, nil
}

func deleteStack(ctx context.Context, svc services.CloudFormationServiceInterface, newStyleName, oldStyleName string) error {
	name := newStyleName
	_, err := svc.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{
		StackName: aws.String(name),
	})
	if doesNotExist(err) {
		name = oldStyleName
	}

	_, err = svc.DeleteStack(ctx, &cloudformation.DeleteStackInput{
		StackName: aws.String(name),
	})
	if err != nil && !doesNotExist(err) {
		return fmt.Errorf("error deleting stack: %w", err)
	}

	return nil
}
