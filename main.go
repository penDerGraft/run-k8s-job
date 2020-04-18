package main

import (
	"context"
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

	action = action.WithFieldsMap(map[string]string{
		"job": jobName,
	})

	config, err := clientcmd.BuildConfigFromFlags(clusterURL, "")
	if err != nil {
		action.Fatalf("%v", err)
	}

	config.CAFile = caFilePath
	config.BearerToken = token

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		action.Fatalf("%v", err)
	}

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
