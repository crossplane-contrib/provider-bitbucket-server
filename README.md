# provider-bitbucket-server

`provider-bitbucket-server` is a [Crossplane](https://crossplane.io/)
Provider that is meant to integrate with the bitbucket server APIs.

The scope is to provide enough resources to provision CI/CD pipelines.

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
