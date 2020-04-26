package main

import (
	"context"
	"time"

	"github.com/sethvargo/go-githubactions"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	action := githubactions.New()

	input := ActionInput{
		kubeconfigFile: action.GetInput("kubeconfig-file"),
		clusterURL:     action.GetInput("cluster-url"),
		clusterToken:   action.GetInput("cluster-token"),
		namespace:      action.GetInput("namespace"),
		image:          action.GetInput("image"),
		jobName:        action.GetInput("job-name"),
		caFile:         action.GetInput("ca-file"),
		allowInsecure:  action.GetInput("allow-insecure"),
	}

	action.Debugf("kubeconfig input %s\n", input.kubeconfigFile)

	config, err := BuildK8sConfig(input)
	if err != nil {
		action.Fatalf("%v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		action.Fatalf("%v", err)
	}

	action = action.WithFieldsMap(map[string]string{
		"job": input.jobName,
	})

	runner := NewJobRunner(clientset.BatchV1().Jobs(input.namespace), clientset.CoreV1().Pods(input.namespace), 5*time.Second, action)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	logs, err := runner.RunJob(ctx, input.jobName, input.namespace, input.image)
	defer cancel()

	if err != nil {
		if len(logs) == 0 {
			action.Fatalf("%v", err)
		} else {
			action.Fatalf("job failed\njob logs:\n%s", logs)
		}
	}

	action.Debugf("job completed successfully\njob logs:\n%s", logs)
}
