[![Nightly e2e tests](https://github.com/rancher/eks-operator/actions/workflows/e2e-latest-rancher.yaml/badge.svg?branch=main)](https://github.com/rancher/eks-operator/actions/workflows/e2e-latest-rancher.yaml)

# rancher/eks-operator

EKS operator is a Kubernetes CRD controller that controls cluster provisioning in Elastic Kubernetes Service using an EKSClusterConfig defined by a Custom Resource Definition.

## Build

Operator binary can be built using the following command:

```bash
    make operator
```

## Deploy operator from source

You can use the following command to deploy a Kind cluster with Rancher manager and operator:

```bash
    make kind-deploy-operator
```

After this, you can also downscale operator deployment and run operator from a local binary.

## AWS credentials

By default the operator authenticates to AWS using the static access key and
secret access key stored in the Kubernetes secret referenced by
`EKSClusterConfigSpec.AmazonCredentialSecret` under the keys
`amazonec2credentialConfig-accessKey` and `amazonec2credentialConfig-secretKey`.

To avoid long-lived AWS credentials, the operator can also use the AWS SDK
default credential chain (IAM instance profile, IRSA, EKS Pod Identity, ECS task
role, environment variables, etc.). This mode is enabled when the referenced
secret either:

- contains `amazonec2credentialConfig-useInstanceProfile: "true"`, or
- has both `amazonec2credentialConfig-accessKey` and
  `amazonec2credentialConfig-secretKey` empty or absent.

When `amazonec2credentialConfig-useInstanceProfile` is set to `"true"`, any
`accessKey`/`secretKey` fields in the same secret are ignored.

When running in this mode, the identity used by the operator pod must be
authorized for the AWS APIs the operator calls
(`eks:*`, `cloudformation:*`, `ec2:*` for the resources it manages, and
`iam:PassRole` for the EKS cluster/node roles). For least privilege we
recommend providing this identity via **EKS Pod Identity** or **IRSA** rather
than the underlying node instance profile.

## Tests

Running unit tests can be done using the following command:

```bash
    make test
```

### E2E 

We run e2e tests after every merged PR and periodically every 24 hours. They are triggered by a [Github action](.github/workflows/e2e-latest-rancher.yaml)

For running e2e set the following variables and run:

```bash
    export AWS_ACCESS_KEY_ID="replace_with_your_value"
    export AWS_SECRET_ACCESS_KEY="replace_with_your_value"
    export AWS_REGION="replace_with_your_value"
    make kind-e2e-tests
```

A Kind cluster will be created, and the e2e tests will be run against it.

To delete the local Kind cluster once e2e tests are completed, run:

```bash
    make delete-local-kind-cluster
```

## Release

#### When should I release?

A KEv2 operator should be released if

* There have been several commits since the last release,
* You need to pull in an update/bug fix/backend code to unblock UI for a feature enhancement in Rancher
* The operator needs to be unRC for a Rancher release

#### How do I release?

Tag the latest commit on the `master` branch. For example, if latest tag is:
* `v1.0.8-rc1` you should tag `v1.0.8-rc2`.
* `v1.0.8` you should tag `v1.0.9-rc1`.

```bash
# Get the latest upstream changes
# Note: `upstream` must be the remote pointing to `git@github.com:rancher/eks-operator.git`.
git pull upstream master --tags

# Export the tag of the release to be cut, e.g.:
export RELEASE_TAG=v1.0.8-rc2

# Create tags locally
git tag -s -a ${RELEASE_TAG} -m ${RELEASE_TAG}

# Push tags
# Note: `upstream` must be the remote pointing to `git@github.com:rancher/eks-operator.git`.
git push upstream ${RELEASE_TAG}
```

After pushing the release tag, you need to run 2 Github actions. You can find them in the Actions tab of the repo:

* `Update EKS operator in rancher/rancher` - This action will update the EKS operator in rancher/rancher repo. It will bump go dependencies.
* `Update EKS Operator in rancher/charts` - This action will update the EKS operator in rancher/charts repo. It will bump the chart version.

#### How do I unRC?

UnRC is the process of removing the rc from a KEv2 operator tag and means the released version is stable and ready for use. Release the KEv2 operator but instead of bumping the rc, remove the rc. For example, if the latest release of EKS operator is:
* `v1.0.8-rc1`, release the next version without the rc which would be `v1.0.8`.
* `v1.0.8`, that has no rc so release that version or `v1.0.9` if updates are available.
