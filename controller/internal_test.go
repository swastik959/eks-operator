package controller

import (
	"github.com/aws/aws-sdk-go-v2/credentials"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	eksv1 "github.com/rancher/eks-operator/pkg/apis/eks.cattle.io/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("newAWSConfigV2", func() {
	const namespace = "default"

	createSecret := func(name string, data map[string][]byte) {
		secrets := coreFactory.Core().V1().Secret()
		_, err := secrets.Create(&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
			Data:       data,
		})
		if err != nil && !apierrors.IsAlreadyExists(err) {
			Expect(err).NotTo(HaveOccurred())
		}
	}

	It("uses static credentials when both keys are populated", func() {
		createSecret("creds-static", map[string][]byte{
			"amazonec2credentialConfig-accessKey": []byte("AKIAEXAMPLE"),
			"amazonec2credentialConfig-secretKey": []byte("secret"),
		})
		spec := eksv1.EKSClusterConfigSpec{
			Region:                 "us-east-1",
			AmazonCredentialSecret: namespace + ":creds-static",
		}
		cfg, err := newAWSConfigV2(ctx, coreFactory.Core().V1().Secret(), spec)
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.Credentials).NotTo(BeNil())
		creds, err := cfg.Credentials.Retrieve(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(creds.AccessKeyID).To(Equal("AKIAEXAMPLE"))
		Expect(creds.SecretAccessKey).To(Equal("secret"))
	})

	It("falls back to the default credential chain when both keys are empty", func() {
		createSecret("creds-empty", map[string][]byte{
			"amazonec2credentialConfig-accessKey": []byte(""),
			"amazonec2credentialConfig-secretKey": []byte(""),
		})
		spec := eksv1.EKSClusterConfigSpec{
			Region:                 "us-east-1",
			AmazonCredentialSecret: namespace + ":creds-empty",
		}
		cfg, err := newAWSConfigV2(ctx, coreFactory.Core().V1().Secret(), spec)
		Expect(err).NotTo(HaveOccurred())
		// No static provider override was applied.
		_, isStatic := cfg.Credentials.(credentials.StaticCredentialsProvider)
		Expect(isStatic).To(BeFalse())
	})

	It("falls back to the default credential chain when useInstanceProfile=true", func() {
		createSecret("creds-instance-profile", map[string][]byte{
			"amazonec2credentialConfig-useInstanceProfile": []byte("true"),
			// stray fields must be ignored when the explicit flag is set
			"amazonec2credentialConfig-accessKey": []byte("ignored"),
			"amazonec2credentialConfig-secretKey": []byte("ignored"),
		})
		spec := eksv1.EKSClusterConfigSpec{
			Region:                 "us-east-1",
			AmazonCredentialSecret: namespace + ":creds-instance-profile",
		}
		cfg, err := newAWSConfigV2(ctx, coreFactory.Core().V1().Secret(), spec)
		Expect(err).NotTo(HaveOccurred())
		// Stray static keys must not have been loaded into a static provider.
		_, isStatic := cfg.Credentials.(credentials.StaticCredentialsProvider)
		Expect(isStatic).To(BeFalse())
	})

	It("returns an error when only one of accessKey/secretKey is populated", func() {
		createSecret("creds-partial", map[string][]byte{
			"amazonec2credentialConfig-accessKey": []byte("AKIAEXAMPLE"),
			"amazonec2credentialConfig-secretKey": []byte(""),
		})
		spec := eksv1.EKSClusterConfigSpec{
			Region:                 "us-east-1",
			AmazonCredentialSecret: namespace + ":creds-partial",
		}
		_, err := newAWSConfigV2(ctx, coreFactory.Core().V1().Secret(), spec)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("invalid aws cloud credential"))
	})

	It("uses the default credential chain when AmazonCredentialSecret is empty", func() {
		spec := eksv1.EKSClusterConfigSpec{Region: "us-east-1"}
		cfg, err := newAWSConfigV2(ctx, coreFactory.Core().V1().Secret(), spec)
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.Region).To(Equal("us-east-1"))
	})
})
