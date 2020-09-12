package internal

import (
	"context"

	jsonpatch "github.com/evanphx/json-patch/v5"
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
	Workflow *Workflow
	FileName string
	FilePath string
}

type FilePatch struct {
	PatchData PatchData
	Patch     jsonpatch.Patch
}

type PatchData struct {
	FileName string               `yaml:"fileName,omitempty" json:"fileName,omitempty"`
	Patch    []JsonPatchOperation `yaml:"patch,omitempty" json:"patch,omitempty"`
}

type JsonPatchOperation struct {
	FileName string `yaml:"name,omitempty" json:"name,omitempty"`
	Op       string `yaml:"op,omitempty" json:"op,omitempty"`
	Path     string `yaml:"path,omitempty" json:"path,omitempty"`
	Value    string `yaml:"value,omitempty" json:"value,omitempty"`
}

type UpdateIntentOptions struct {
	Branch      string
	PRBranchRef string
	BaseRef     string
	Owner       string
	Repo        string
}

type UpdateWorkflowIntent struct {
	RepositoryName string
	WorkflowDrafts []*WorkflowUpdateDraft
	Options        UpdateIntentOptions
}

type WorkflowUpdateDraft struct {
	Workflow    *Workflow
	FileContent *[]byte
	Filename    string
	SHA         string
	FilePath    string
	DisplayName string
	Url         string
}
