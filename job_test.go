package main

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/rest/fake"
)

type jobClientMock struct {
	createFn func(*v1.Job) (*v1.Job, error)
	getFn    func(string, metav1.GetOptions) (*v1.Job, error)
	throwErr bool
}

func (j jobClientMock) Create(job *v1.Job) (*v1.Job, error) {
	return j.createFn(job)
}
func (j jobClientMock) Get(name string, options metav1.GetOptions) (*v1.Job, error) {
	return j.getFn(name, options)
}

type podClientMock struct {
	logsChecked bool
}

func (p *podClientMock) GetLogs(name string, opts *corev1.PodLogOptions) *rest.Request {

	c := fake.CreateHTTPClient(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader("pod log stream")),
		}, nil
	})
	fakeClient := fake.RESTClient{Client: c}

	p.logsChecked = true
	return fakeClient.Request()
}
func (p *podClientMock) List(opts metav1.ListOptions) (*corev1.PodList, error) {
	podMeta := metav1.ObjectMeta{Name: "test-job-pod"}
	return &corev1.PodList{Items: []corev1.Pod{{ObjectMeta: podMeta}}}, nil
}

type noopLogger struct{}

func (noopLogger) Debugf(msg string, args ...interface{})   {}
func (noopLogger) Errorf(msg string, args ...interface{})   {}
func (noopLogger) Fatalf(msg string, args ...interface{})   {}
func (noopLogger) Warningf(msg string, args ...interface{}) {}

func TestRunJob(t *testing.T) {
	testCases := []struct {
		desc           string
		condition      v1.JobConditionType
		createJobError bool
		getJobError    bool
		checkLogs      bool
		wantErr        bool
	}{
		{
			desc:      "job created succesfully, job succeeded",
			condition: v1.JobComplete,
			checkLogs: true,
			wantErr:   false,
		},
		{
			desc:      "job created successfully, job failed",
			condition: v1.JobFailed,
			checkLogs: true,
			wantErr:   true,
		},
		{
			desc:           "create job failure",
			condition:      v1.JobFailed,
			createJobError: true,
			checkLogs:      false,
			wantErr:        true,
		},
		{
			desc:        "get job failure",
			condition:   v1.JobFailed,
			getJobError: true,
			checkLogs:   false,
			wantErr:     true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			assert := assert.New(t)

			jm := newJobClientMock(tc.condition, tc.createJobError, tc.getJobError)
			pc := podClientMock{}
			r := NewJobRunner(jm, &pc, 1*time.Second, githubactions.New())
			_, err := r.RunJob(context.Background(), "test-job", "ns", "my-image")

			assert.Equal(tc.checkLogs, pc.logsChecked)

			if tc.wantErr {
				assert.NotNil(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func newJobClientMock(condition v1.JobConditionType, createError, getError bool) jobClient {
	jobMeta := metav1.ObjectMeta{Name: "test-job"}

	return jobClientMock{
		createFn: func(*v1.Job) (*v1.Job, error) {
			if createError {
				return nil, errors.New("error creating job")
			}

			return &v1.Job{ObjectMeta: jobMeta}, nil
		},
		getFn: func(string, metav1.GetOptions) (*v1.Job, error) {
			if getError {
				return nil, errors.New("error finding job")
			}

			return &v1.Job{
				ObjectMeta: jobMeta,
				Status: v1.JobStatus{
					Conditions: []v1.JobCondition{{
						Type:    condition,
						Status:  corev1.ConditionTrue,
						Reason:  "ConditionReason",
						Message: "This is what happened",
					}},
				},
			}, nil
		},
	}
}
