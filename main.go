package main

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/sethvargo/go-githubactions"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	action := githubactions.New()

	clusterURL := action.GetInput("cluster-url")
	token := action.GetInput("cluster-token")
	namespace := action.GetInput("namespace")
	image := action.GetInput("image")
	jobName := action.GetInput("job-name")
	caFilePath := action.GetInput("ca-file")
	tlsFlag := action.GetInput("disable-tls")

	if len(clusterURL) == 0 {
		action.Fatalf("'cluster-url' is a required input but was empty")
	}

	if len(token) == 0 {
		action.Fatalf("'cluster-token' is a required input but was empty")
	}

	if len(image) == 0 {
		action.Fatalf("'image' is a required input but was empty")
	}

	disableTLS, err := strconv.ParseBool(tlsFlag)
	if err != nil {
		action.Fatalf("'disable-tls input must be either 'true' or 'false', was %s", tlsFlag)
	}

	if !disableTLS {
		if len(caFilePath) == 0 {
			action.Fatalf("you must either specify the file path to the root ca or explicitly disable tls using the 'disable-tls' input")
		}

		if _, err := os.Stat(caFilePath); os.IsNotExist(err) {
			action.Fatalf("could not locate file %s; please make sure the file is available in the runner's context", caFilePath)
		}
	}

	config, err := clientcmd.BuildConfigFromFlags(clusterURL, "")
	if err != nil {
		action.Fatalf("%v", err)
	}

	config.Insecure = disableTLS
	config.CAFile = caFilePath
	config.BearerToken = token

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		action.Fatalf("%v", err)
	}

	action = action.WithFieldsMap(map[string]string{
		"job": jobName,
	})

	runner := NewJobRunner(clientset.BatchV1().Jobs(namespace), clientset.CoreV1().Pods(namespace), 5*time.Second, action)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	logs, err := runner.RunJob(ctx, jobName, namespace, image)
	defer cancel()

	if err != nil {
		if len(logs) == 0 {
			action.Fatalf("%v", err)
		} else {
			action.Fatalf("job failed - job logs:\n%s", logs)
		}
	}

	action.Debugf("job completed successfully - job logs:\n%s", logs)
}
