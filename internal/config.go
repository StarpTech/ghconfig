package internal

import (
	"context"

	"github.com/google/go-github/v32/github"
	"github.com/teris-io/shortid"
)

type Config struct {
	GithubClient    *github.Client
	Context         context.Context
	DryRun          bool
	CreatePR        bool
	BaseBranch      string
	Sid             *shortid.Shortid
	RepositoryQuery string
	WorkflowRoot    string
}

type WorkflowTemplate struct {
	Workflow           *GithubWorkflow
	Filename           string
	RepositoryFilePath string
}

type RepositoryFileUpdateOptions struct {
	FileContent *[]byte
	Filename    string
	SHA         string
	FilePath    string
	DisplayName string
	URL         string
}

type PatchData struct {
	Filename string               `yaml:"filename,omitempty" json:"filename,omitempty"`
	Patch    []JsonPatchOperation `yaml:"patch,omitempty" json:"patch,omitempty"`
}

type JsonPatchOperation struct {
	Op    string `yaml:"op,omitempty" json:"op,omitempty"`
	Path  string `yaml:"path,omitempty" json:"path,omitempty"`
	Value string `yaml:"value,omitempty" json:"value,omitempty"`
}

type RepositoryUpdateOptions struct {
	Branch      string
	PRBranchRef string
	BaseRef     string
	Owner       string
	Repo        string
}

type WorkflowUpdatePackage struct {
	Repository        *github.Repository
	Files             []*WorkflowUpdatePackageFile
	RepositoryOptions RepositoryUpdateOptions
}

type WorkflowUpdatePackageFile struct {
	Workflow                *GithubWorkflow
	RepositoryUpdateOptions *RepositoryFileUpdateOptions
}
