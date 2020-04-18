package main

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

var (
	errJobNotCreated = errors.New("job not created")
	errJobNotFound   = errors.New("job not found")
	errPodNotFound   = errors.New("pod not found")
	errLogsNotFound  = errors.New("pod logs not found")
)

type logger interface {
	Debugf(msg string, args ...interface{})
	Errorf(msg string, args ...interface{})
	Fatalf(msg string, args ...interface{})
	Warningf(msg string, args ...interface{})
}

type jobClient interface {
	Create(*v1.Job) (*v1.Job, error)
	Get(name string, options metav1.GetOptions) (*v1.Job, error)
}

type podClient interface {
	GetLogs(name string, opts *corev1.PodLogOptions) *rest.Request
	List(opts metav1.ListOptions) (*corev1.PodList, error)
}

type JobRunner struct {
	jc           jobClient
	pc           podClient
	pollInterval time.Duration
	log          logger
}

func NewJobRunner(jc jobClient, pc podClient, pollInterval time.Duration, log logger) JobRunner {
	return JobRunner{
		jc:           jc,
		pc:           pc,
		pollInterval: pollInterval,
		log:          log,
	}
}

func (j *JobRunner) RunJob(ctx context.Context, jobPrefix, namespace, image string) (string, error) {
	job, err := j.jc.Create(&v1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", jobPrefix),
			Namespace:    namespace,
		},
		Spec: v1.JobSpec{
			BackoffLimit: intPtr(0),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("%s-pod", jobPrefix),
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:  fmt.Sprintf("%s-con", jobPrefix),
							Image: image,
						},
					},
				},
			},
		},
	})

	if err != nil {
		return "", errors.Wrapf(errJobNotCreated, "error starting job: %v", err)
	}

	err = j.pollJobStatus(ctx, job.GetName())
	if err != nil {
		if errors.Is(err, errJobNotFound) {
			return "", err
		}

		logString, _ := j.getLogs(ctx, job.GetName())
		return logString, err
	}

	return j.getLogs(ctx, job.GetName())
}

func (j *JobRunner) pollJobStatus(ctx context.Context, jobName string) error {
	ticker := time.NewTicker(j.pollInterval)
	for {
		select {
		case <-ticker.C:
			job, err := j.getJob(ctx, jobName)
			if err != nil {
				return errors.Wrapf(errJobNotFound, "could not find job: %v", err)
			}

			if job != nil {
				if i := findCondition(job.Status.Conditions, v1.JobFailed); i > -1 {
					return errors.Errorf("job failed with %s: %s", job.Status.Conditions[i].Reason, job.Status.Conditions[i].Message)
				}

				if i := findCondition(job.Status.Conditions, v1.JobComplete); i > -1 {
					return nil
				}
			}

			j.log.Debugf("job running, awaiting status")

		case <-ctx.Done():
			ticker.Stop()
			return errors.New("job exceeded timeout limit")
		}
	}
}

func (j *JobRunner) getJob(ctx context.Context, jobName string) (*v1.Job, error) {
	return j.jc.Get(jobName, metav1.GetOptions{})
}

func (j *JobRunner) getLogs(ctx context.Context, jobName string) (string, error) {
	pods, err := j.pc.List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", jobName),
	})
	if err != nil {
		return "", errors.Wrapf(errPodNotFound, "could not find pod: %v", err)
	}

	// Need to figure out how to handle jobs where
	// multiple pods are created because of backoffPolicy
	podName := pods.Items[0].GetName()
	logBytes, err := j.pc.GetLogs(podName, &corev1.PodLogOptions{}).DoRaw()
	if err != nil {
		return "", errors.Wrapf(errLogsNotFound, "could not read logs: %v", err)
	}

	return string(logBytes), err
}

func findCondition(conditions []v1.JobCondition, condition v1.JobConditionType) int {
	for i, c := range conditions {
		if c.Type == condition && c.Status == corev1.ConditionTrue {
			return i
		}
	}

	return -1
}

func intPtr(i int32) *int32 {
	return &i
}
