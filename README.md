# provider-bitbucket-server

`provider-bitbucket-server` is a [Crossplane](https://crossplane.io/)
Provider that is meant to integrate with the bitbucket server APIs.

It does not support the API for Bitbucket Cloud.

The scope of the current feature set is to provide enough resources to provision CI/CD pipelines.

## Configure

Create a secret containing an API token (go to Profile, Manage account, Personal Access Token), and configure a Bitbucket Server ProviderConfig with a BaseURL pointing to your bitbucket server:
[embedmd]:# (examples/provider/config.yaml yaml)
```yaml
apiVersion: v1
kind: Secret
metadata:
  namespace: crossplane-system
  name: example-provider-secret
type: Opaque
stringData:
  credentials: "foo"
---
apiVersion: bitbucket-server.crossplane.io/v1alpha1
kind: ProviderConfig
metadata:
  name: example
spec:
  baseURL: https://bitbucket.company.example.com
  credentials:
    source: Secret
    secretRef:
      namespace: crossplane-system
      name: example-provider-secret
      key: credentials
```

## Usage

The following resources can be created:

### AccessKey

Set up access keys to git repositories. They can be read only or
read+write. The bitbucket server has strict validation of this
resource which you must know:
* All fields are immutable except permission
* You can't upload a key which is already used as a personal keys
* You can't upload a key to a repo if the key already has access (for
  example at the project level)

[embedmd]:# (examples/accesskey/accesskey.yaml yaml)
```yaml
apiVersion: accesskey.bitbucket-server.crossplane.io/v1alpha1
kind: AccessKey
metadata:
  name: example
spec:
  forProvider:
    projectKey: TEST
    repoName: test
    publicKey:
      key: "ssh-rsa 100"
      label: "test2"
      permission: "REPO_WRITE"
  providerConfigRef:
    name: example
```

### Webhook
The webhook resource is fully mutable and refers to an URL which will
be triggered when the configured events occur:

[embedmd]:# (examples/webhook/webhook.yaml yaml)
```yaml
apiVersion: webhook.bitbucket-server.crossplane.io/v1alpha1
kind: Webhook
metadata:
  name: example
spec:
  forProvider:
    projectKey: TEST
    repoName: test
    webhook:
      name: "build-trigger"
      configuration:
        secret: "123"
      events:
        - "repo:refs_changed"
        - "repo:modified"
      url: "https://hooks.example.com/test"
  providerConfigRef:
    name: example
```

## Developing


https://docs.atlassian.com/bitbucket-server/rest/7.10.0/bitbucket-rest.html
https://docs.atlassian.com/bitbucket-server/rest/7.10.0/bitbucket-ssh-rest.html

Run against a Kubernetes cluster:

```console
make run
```

Install `latest` into Kubernetes cluster where Crossplane is installed:

```console
make install
```

Install local build into [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/)
cluster where Crossplane is installed:

```console
make install-local
```

Build, push, and install:

```console
make all
```

Build image:

```console
make image
```

Push image:

```console
make push
```

Build binary:

```console
make build
```
