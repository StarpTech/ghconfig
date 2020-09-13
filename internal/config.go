package internal

import (
	"context"

	"github.com/google/go-github/v32/github"
)

type (
	IDGenerator interface {
		MustGenerate() string
	}
	Config struct {
		GithubClient    *github.Client
		Context         context.Context
		DryRun          bool
		CreatePR        bool
		BaseBranch      string
		Sid             IDGenerator
		RepositoryQuery string
		WorkflowRoot    string
		RootDir         string
	}

	WorkflowTemplate struct {
		Workflow           *GithubWorkflow
		Filename           string
		RepositoryFilePath string
	}

	RepositoryFileUpdateOptions struct {
		FileContent *[]byte
		Filename    string
		SHA         string
		FilePath    string
		DisplayName string
		URL         string
	}

	PatchData struct {
		Filename string               `yaml:"filename,omitempty" json:"filename,omitempty"`
		Patch    []JsonPatchOperation `yaml:"patch,omitempty" json:"patch,omitempty"`
	}

	JsonPatchOperation struct {
		Op    string `yaml:"op,omitempty" json:"op,omitempty"`
		Path  string `yaml:"path,omitempty" json:"path,omitempty"`
		Value string `yaml:"value,omitempty" json:"value,omitempty"`
	}

	RepositoryUpdateOptions struct {
		Branch      string
		PRBranchRef string
		BaseRef     string
		Owner       string
		Repo        string
	}

	WorkflowUpdatePackage struct {
		Repository        *github.Repository
		Files             []*WorkflowUpdatePackageFile
		RepositoryOptions RepositoryUpdateOptions
	}

	WorkflowUpdatePackageFile struct {
		Workflow                *GithubWorkflow
		RepositoryUpdateOptions *RepositoryFileUpdateOptions
	}
)
