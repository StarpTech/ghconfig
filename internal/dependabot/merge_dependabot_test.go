package dependabot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	Description string
	Dst         GithubDependabot
	Src         GithubDependabot
	Output      GithubDependabot
}

func TestSync_MergeDependabot(t *testing.T) {

	// dst: Remote Workflow on github
	// src: templated local worklfow
	testcases := []testCase{
		{
			Description: "Dst is always overriden by Src",
			Dst: GithubDependabot{
				Version: "1",
				Updates: []*Updates{
					{
						Directory: "/foo",
						Ignore: []Ignore{
							{DependencyName: "dep", Versions: []string{"1.0.0", "2.0.0"}},
						},
					},
					{
						Directory: "/bar",
					},
					{
						Directory:        "/bar",
						PackageEcosystem: "docker",
					},
				},
			},
			Src: GithubDependabot{
				Version: "2",
				Updates: []*Updates{
					{
						Directory:             "/foo",
						OpenPullRequestsLimit: 5,
					},
					{
						Directory: "/hello",
					},
					{
						Directory: "/bar",
					},
				},
			},
			Output: GithubDependabot{
				Version: "2",
				Updates: []*Updates{
					{
						Directory:             "/foo",
						OpenPullRequestsLimit: 5,
						Ignore: []Ignore{
							{DependencyName: "dep", Versions: []string{"1.0.0", "2.0.0"}},
						},
					},
					{
						Directory: "/hello",
					},
					{
						Directory: "/bar",
					},
				},
			},
		},
	}

	for _, testcase := range testcases {
		err := MergeDependabot(&testcase.Dst, testcase.Src)
		assert.Nil(t, err, testcase.Description)
		assert.EqualValues(t, testcase.Dst, testcase.Output, testcase.Description)
	}
}
