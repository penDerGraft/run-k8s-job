package main

import (
	"testing"
)

func TestBuildK8sConfig(t *testing.T) {
	testCases := []struct {
		desc    string
		input   ActionInput
		wantErr bool
	}{
		{
			desc:    "missing input 'image'",
			input:   ActionInput{image: ""},
			wantErr: true,
		},
		{
			desc:    "missing input 'cluster-token'",
			input:   ActionInput{image: "test-image", clusterURL: "localhost:8000"},
			wantErr: true,
		},
		{
			desc:    "missing input 'clusterURL'",
			input:   ActionInput{image: "test-image", clusterURL: ""},
			wantErr: true,
		},
		{
			desc: "missing input 'caFile' without 'allow-insecure'",
			input: ActionInput{
				image:        "test-image",
				clusterURL:   "localhost:8000",
				clusterToken: "test-token",
				caFile:       "",
			},
			wantErr: true,
		},
		{
			desc: "non-bool 'allow-insecure' input",
			input: ActionInput{
				image:         "test-image",
				clusterURL:    "localhost:8000",
				clusterToken:  "test-token",
				allowInsecure: "not a bool",
			},
			wantErr: true,
		},
		{
			desc:    "non-base64 'kubeconfig-file' input",
			input:   ActionInput{image: "test-image", kubeconfigFile: "not base64"},
			wantErr: true,
		},
		{
			desc: "'allow-insecure' without 'caFile'",
			input: ActionInput{
				image:         "test-image",
				clusterURL:    "localhost:8000",
				clusterToken:  "test-token",
				allowInsecure: "true",
			},
			wantErr: false,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.desc, func(t *testing.T) {

			_, err := BuildK8sConfig(tt.input)

			if didErr := err != nil; didErr != tt.wantErr {
				t.Errorf("'wantErr' was %t, but err value was: '%v'", tt.wantErr, err)
			}
		})
	}
}
