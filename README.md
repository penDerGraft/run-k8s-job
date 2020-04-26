# Run Kubernetes Job Action

Run an arbitrary docker image as a job on a Kubernetes cluster and report the output to stdout.

## Why?

For services/apps running on Kubernetes the `run-k8s-job` action allows you to define an arbitrary task as an explict step in a GitHub workflow, without having to deal with a lot of Kubernetes-specific details (you just need a Docker image). This can be useful for creating automated stage gates in a deployment pipeline, or kicking off any task that may be repeated based on the GitHub [events that trigger workflows](https://help.github.com/en/actions/reference/events-that-trigger-workflows). 

It's also not always entirely straightforward to get the output of a previously executed Kubernetes job. This action will grab job status and any logs and output them to the actions console. 

Some example uses might be:
- smoke/integration tests against a live environment
- load tests
- dependency provisioning
- database migrations

## Usage

```yaml
    - uses: ./
      with:
        kubeconfig-file: '${{ secrets.KUBECONFIG_FILE }}'        
        image: 'penDerGraft/integration-test-job'        
```

### Auth Strategies

To create a new job in a cluster, `run-k8s-job` needs credentials to authenticate against the canonical cluster endpoint. This is the IP address of the Kubernetes server that handles API access for the cluster. There are two ways to do this.

#### Using `kubeconfig-file` 
This is a base64 encoded version of the config file that is stored at `$HOME/.kube/config`. If you can access your cluster with `kubectl` this file should already have the necessary data for authentication. 

#### Using `cluster-url`,  `cluster-token` and `ca-file` 
These values are included in your kubeconfig file, but there may be cases when it's more straightforward to specifiy them directly. Note that you can omit the `ca-file` if you set `allow-insecure` to true, but this should only be used in testing situations. 

### Inputs

- `kubeconfig-file` a base64 encoded kubeconfig file with credentials for the cluster. This file is saved to `$HOME/.kube/config` by default. Get a base64 encoded string of the file by running `cat <path-to-file> | base64 --encode` where `<path-to-file>` is the path to your kubeconfig file. 
- `cluster-url` the cluster endpoint for your Kubernetes cluster.
- `cluster-token` the OAuth Bearer token for your Kubernetes cluster.
- `image` the docker image to be run as a job. The image must be publically accessible. 
- `ca-file` the path to the root CA certificates (in PEM format) for establishing a TLS connection to the Kubernetes server. Note: the step will fail if a ca file is not provided and the `disable-tls` input is not explicitly set to false. It is **highly recommended** that a CA file be specified.
- `job-name` prefix for the auto-generated job name in Kubernetes. Defaults to the name of the repo. 
- `namespace` the Kubernetes namespace where the job should run. Defaults to `default`
- `allow-insecure` connect to the Kubernetes server insecurely (without verifying the certificate authority). Should only be used for testing purposes as it leaves the connection vulnerable to [man-in-the-middle attacks](https://en.wikipedia.org/wiki/Man-in-the-middle_attack).




