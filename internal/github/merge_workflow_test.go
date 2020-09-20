package github

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	Description string
	Dst         GithubWorkflow
	Src         GithubWorkflow
	Output      GithubWorkflow
}

func TestSync_MergeWorkflow(t *testing.T) {

	// dst: Remote file on github
	// src: templated local file
	testcases := []testCase{
		{
			Description: "Primitive and array values are overriden by Src",
			Dst: GithubWorkflow{
				Env: map[string]string{
					"token": "123",
				},
				Name: "name_dst",
			},
			Src: GithubWorkflow{
				Env: map[string]string{
					"existing": "223",
				},
				Name: "name_src",
			},
			Output: GithubWorkflow{
				Env: map[string]string{
					"existing": "223",
					"token":    "123",
				},
				Name: "name_src",
			},
		},
		{
			Description: "Dst fields are preserved when their fields reference to the same entity and Src is empty",
			Dst: GithubWorkflow{
				Jobs: map[string]*Job{
					"build": {
						If: "if",
					},
				},
				Name: "name_dst",
			},
			Src: GithubWorkflow{
				Jobs: map[string]*Job{
					"build": {},
				},
				Name: "name_src",
			},
			Output: GithubWorkflow{
				Jobs: map[string]*Job{
					"build": {
						If: "if",
					},
				},
				Name: "name_src",
			},
		},
		{
			Description: "Empty dst",
			Dst:         GithubWorkflow{},
			Src: GithubWorkflow{
				Jobs: map[string]*Job{
					"build": {},
				},
				Name: "name_src",
			},
			Output: GithubWorkflow{
				Jobs: map[string]*Job{
					"build": {},
				},
				Name: "name_src",
			},
		},
		{
			Description: "Merge Schedules",
			Dst: GithubWorkflow{
				On: On{
					Schedule: []Schedule{
						{Cron: "cron"},
						{Cron: "cron"},
					},
				},
				Name: "name_dst",
			},
			Src: GithubWorkflow{
				On: On{
					Schedule: []Schedule{
						{Cron: ""},
						{Cron: "cro2"},
					},
				},
				Name: "name_src",
			},
			Output: GithubWorkflow{
				On: On{
					Schedule: []Schedule{
						{Cron: "cron"},
						{Cron: "cro2"},
					},
				},
				Name: "name_src",
			},
		},
		{
			Description: "Steps from same Job with the same (name or id or properties) are merged. No steps from dst are appended.",
			Dst: GithubWorkflow{
				Jobs: map[string]*Job{
					"build": {
						Name: "build",
						Steps: []*Step{
							{Name: "install", ContinueOnError: true},
							{Run: "npm test"},
						},
					},
				},
			},
			Src: GithubWorkflow{
				Jobs: map[string]*Job{
					"build": {
						Name: "build",
						Steps: []*Step{
							{Name: "install"},
							{Run: "npm ci"},
						},
					},
				},
			},
			Output: GithubWorkflow{
				Jobs: map[string]*Job{
					"build": {
						Name: "build",
						Steps: []*Step{
							{Name: "install", ContinueOnError: true},
							{Run: "npm ci"},
						},
					},
				},
			},
		},
		{
			Description: "Jobs with the same name are merged. No jobs from dst are appended.",
			Dst: GithubWorkflow{
				Jobs: map[string]*Job{
					"build": {
						Name:  "build",
						Needs: StringArray{"a"},
					},
				},
			},
			Src: GithubWorkflow{
				Jobs: map[string]*Job{
					"build": {
						Name:            "build",
						ContinueOnError: true,
					},
				},
			},
			Output: GithubWorkflow{
				Jobs: map[string]*Job{
					"build": {
						Name:            "build",
						ContinueOnError: true,
						Needs:           StringArray{"a"},
					},
				},
			},
		},
		{
			Description: "Complex merge",
			Dst: GithubWorkflow{
				Env: map[string]string{
					"token": "123",
				},
				Name: "name_dst",
				On: On{
					Push: Push{
						Branches: []string{"master"},
					},
					PullRequest: PullRequest{
						Branches: []string{"master"},
					},
					Schedule: []Schedule{
						{Cron: "not_added"},
					},
				},
				Jobs: map[string]*Job{
					"build": {
						Name: "build",
						If:   "if",
						Services: map[string]*Service{
							"svc":  {Ports: []string{"8081", "9090"}},
							"svc2": {Ports: []string{"8081", "9090"}},
						},
						Steps: []*Step{
							{Name: "install", If: "if"},
							{Run: "npm test"},
							{Name: "not_added"},
						},
						Strategy: Strategy{
							Matrix: map[string]interface{}{
								"node-version": []MatrixValue{"12.x", "14.x"},
								"os":           []MatrixValue{"ubuntu-latest", "windows-latest", "macOS-latest"},
								"include": []MatrixValue{
									map[string]string{"node": "12"},
								},
							},
						},
					},
					"test": {
						Name: "test",
					},
				},
			},
			Src: GithubWorkflow{
				Env: map[string]string{
					"existing": "223",
				},
				On: On{
					PageBuild: "pageBuild",
					Push: Push{
						Branches:    []string{"feature"},
						PathsIgnore: []string{"docs/*"},
					},
					Schedule: []Schedule{
						{Cron: "123"},
						{Cron: "12345"},
					},
				},
				Name: "name_src",
				Jobs: map[string]*Job{
					"build": {
						Name: "build",
						Services: map[string]*Service{
							"svc":  {Ports: []string{"8080", "9090", "4040"}},
							"svc3": {Ports: []string{"8081", "9090"}},
						},
						Steps: []*Step{
							{Name: "install", ContinueOnError: true},
							{Run: "npm test"},
							{Name: "build"},
						},
						Strategy: Strategy{
							Matrix: map[string]interface{}{
								"node-version": []MatrixValue{"12.x", "13.x", "14.x"},
								"os":           []MatrixValue{"ubuntu-latest", "windows-latest"},
								"include": []MatrixValue{
									map[string]string{"node": "12"},
								},
							},
						},
					},
					"foo": {
						Name: "foo",
					},
				},
			},

			Output: GithubWorkflow{
				Env: map[string]string{
					"existing": "223",
					"token":    "123",
				},
				Name: "name_src",
				On: On{
					PageBuild: "pageBuild",
					Push: Push{
						Branches:    []string{"feature", "master"},
						PathsIgnore: []string{"docs/*"},
					},
					PullRequest: PullRequest{
						Branches: []string{"master"},
					},
					Schedule: []Schedule{
						{Cron: "not_added"},
						{Cron: "123"},
						{Cron: "12345"},
					},
				},
				Jobs: map[string]*Job{
					"build": {
						Name: "build",
						If:   "if",
						Services: map[string]*Service{
							"svc":  {Ports: []string{"4040", "8080", "8081", "9090"}},
							"svc3": {Ports: []string{"8081", "9090"}},
						},
						Steps: []*Step{
							{Name: "install", ContinueOnError: true, If: "if"},
							{Run: "npm test"},
							{Name: "build"},
						},
						Strategy: Strategy{
							Matrix: Matrix{
								"node-version": []string{"12.x", "13.x", "14.x"},
								"os":           []string{"macOS-latest", "ubuntu-latest", "windows-latest"},
								"include": []MatrixValue{
									map[string]string{"node": "12"},
								},
							},
						},
					},
					"test": {
						Name: "test",
					},
					"foo": {
						Name: "foo",
					},
				},
			},
		},
	}

	for _, testcase := range testcases {
		err := MergeWorkflow(&testcase.Dst, testcase.Src)
		assert.Nil(t, err, testcase.Description)
		assert.EqualValues(t, testcase.Dst, testcase.Output, testcase.Description)
	}
}
