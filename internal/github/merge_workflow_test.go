package github

import (
	"testing"

	"github.com/tj/assert"
)

func TestSync_MergeWorkflow(t *testing.T) {

	// Remote Workflow on github
	src := GithubWorkflow{
		Env: map[string]string{
			"token": "123",
		},
		Name: "name_src",
		On: On{
			Push: Push{
				Branches: []string{"master"},
			},
			Schedule: []Schedule{
				{Cron: "123"},
				{Cron: "123"},
				{Cron: "123"},
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
				},
				Strategy: Strategy{
					Matrix: map[string]interface{}{
						"node-version": []string{"11.x", "12.x", "14.x", "14.x"},
						"include": []map[string]string{
							{"node": "12"},
						},
					},
				},
			},
			"test": {
				Name: "test",
			},
		},
	}

	// templated local worklfow
	dst := GithubWorkflow{
		Env: map[string]string{
			"existing": "223",
		},
		On: On{
			PageBuild: "pageBuild",
			Push: Push{
				PathsIgnore: []string{"docs/*"},
			},
			Schedule: []Schedule{
				{Cron: "123"},
				{Cron: "12345"},
			},
		},
		Name: "name_dst",
		Jobs: map[string]*Job{
			"build": {
				Name: "build",
				Services: map[string]*Service{
					"svc":  {Ports: []string{"8080", "9090", "4040"}},
					"svc3": {Ports: []string{"8081", "9090"}},
				},
				Steps: []*Step{
					{Name: "install", ContinueOnError: true},
					{Name: "install", ContinueOnError: true},
					{Run: "npm test"},
					{Name: "build"},
				},
				Strategy: Strategy{
					Matrix: map[string]interface{}{
						"node-version": []string{"11.x", "12.x", "13.x"},
						"include": []map[string]string{
							{"node": "12", "os": "windows-latest"},
							{"not_exist_in_src": "12"},
						},
					},
				},
			},
			"foo": {
				Name: "foo",
			},
		},
	}

	output := GithubWorkflow{
		Env: map[string]string{
			"token":    "123",
			"existing": "223",
		},
		Name: "name_src",
		On: On{
			PageBuild: "pageBuild",
			Push: Push{
				Branches:    []string{"master"},
				PathsIgnore: []string{"docs/*"},
			},
			Schedule: []Schedule{
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
					"svc2": {Ports: []string{"8081", "9090"}},
					"svc3": {Ports: []string{"8081", "9090"}},
				},
				Steps: []*Step{
					{Name: "install", ContinueOnError: true, If: "if"},
					{Run: "npm test"},
					{Name: "build"},
				},
				Strategy: Strategy{
					Matrix: Matrix{
						"node-version": []string{"11.x", "12.x", "13.x", "14.x"},
						"include": []map[string]string{
							{"node": "12"},
							{"node": "12", "os": "windows-latest"},
							{"not_exist_in_src": "12"},
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
	}

	err := MergeWorkflow(&dst, src)
	assert.Nil(t, err)

	err = MergeWorkflow(&dst, src)
	assert.Nil(t, err)

	// bb, _ := json.Marshal(&dst)
	// fmt.Println(string(bb))

	// spew.Dump(output.Jobs["build"].Steps)
	assert.EqualValues(t, dst, output)
}
