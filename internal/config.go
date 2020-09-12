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

type WorkflowPatch struct {
	Patch    jsonpatch.Patch
	FileName string
}

type UpdateWorkflowIntentOptions struct {
	Branch      string
	PRBranchRef string
	BaseRef     string
	Owner       string
	Repo        string
}

type UpdateWorkflowIntent struct {
	RepositoryName string
	WorkflowDrafts []*WorkflowUpdateDraft
	Options        UpdateWorkflowIntentOptions
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
