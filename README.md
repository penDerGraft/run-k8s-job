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



## Required Inputs

- `cluster-url` the base URL for your Kubernetes cluster.
- `cluster-token` the OAuth Bearer token for your Kubernetes cluster.
- `image` the docker image to be run as a job. The image must be publically accessible. 
- `ca-file` the path to the root CA certificates (in PEM format) for establishing a TLS connection to the Kubernetes server. Note: this is not is not strictly a required input, but the step will fail if one is not provided and the `disable-tls` input is not explicitly set to false. It is **highly recommended** that a CA file be specified.

## Optional Inputs
- `job-name` prefix for the auto-generated job name in Kubernetes. Defaults to the name of the repo. 
- `namespace` the Kubernetes namespace where the job should run. Defaults to `default`
- `disable-tls` connect to the Kubernetes server insecurely (without TLS). Should only be used for testing purposes as it leaves the connection vulnerable to [man-in-the-middle attacks](https://en.wikipedia.org/wiki/Man-in-the-middle_attack).




