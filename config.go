package main

import (
	"encoding/base64"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	kubeconfigPath = "run-k8s-job-kubeconfig"
)

var (
	errNoAuth = errors.New("you must provide either 'kubeconfig-file' or both 'cluster-url' and 'cluster-token'")
)

type ActionInput struct {
	kubeconfigFile string
	image          string
	jobName        string
	namespace      string
	clusterURL     string
	clusterToken   string
	caFile         string
	allowInsecure  string
}

func BuildK8sConfig(input ActionInput) (*rest.Config, error) {
	if len(input.image) == 0 {
		return nil, errors.New("'image' is a required input but was empty")
	}

	if len(input.kubeconfigFile) == 0 {
		return buildConfigWithSecondaryAuth(input)
	}

	data, err := base64.StdEncoding.DecodeString(input.kubeconfigFile)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode kubeconfig file")
	}

	err = ioutil.WriteFile(kubeconfigPath, data, 0644)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode kubeconfig file")
	}
	defer os.Remove(kubeconfigPath)

	return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
}

func buildConfigWithSecondaryAuth(input ActionInput) (*rest.Config, error) {
	if len(input.clusterURL) == 0 || len(input.clusterToken) == 0 {
		return nil, errors.Wrap(errNoAuth, "missing input for cluster authentication")
	}

	allowInsecure, err := strconv.ParseBool(input.allowInsecure)
	if err != nil {
		return nil, errors.Errorf("'allow-insecure input must be either 'true' or 'false', was %s", input.allowInsecure)
	}

	if !allowInsecure {
		if len(input.caFile) == 0 {
			return nil, errors.New("you must either include 'caPath' or set 'allow-insecure' to true")
		}

		if _, err := os.Stat(input.caFile); os.IsNotExist(err) {
			return nil, errors.Errorf("could not locate file %s; please make sure the file is available in the runner's context", input.caFile)
		}
	}

	config, err := clientcmd.BuildConfigFromFlags(input.clusterURL, "")
	if err != nil {
		return nil, err
	}

	config.Insecure = allowInsecure
	config.BearerToken = input.clusterToken
	config.CAFile = input.caFile

	return config, nil
}
