package config

import (
	"context"
	"ghconfig/internal/dependabot"
	gh "ghconfig/internal/github"

	"github.com/google/go-github/v32/github"
)

var (
	BranchNamePattern   = "ghconfig/workflows/%s"
	GhWorkflowDir       = "workflows"
	GhConfigBaseDir     = ".ghconfig"
	GhPatchesDir        = "patches"
	GithubConfigBaseDir = ".github"
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
		RootDir         string
	}

	FileYAML struct {
		Content string `yaml:"content"`
	}

	TemplateVars = map[string]interface{}

	DependabotTemplate struct {
		Dependabot     *dependabot.GithubDependabot
		Filename       string
		RepositoryPath string
	}

	GithubHealthFile struct {
		Filename    string
		Path        string
		FileContent *[]byte
	}

	WorkflowTemplate struct {
		Workflow       *gh.GithubWorkflow
		Filename       string
		RepositoryPath string
	}

	RepositoryFileUpdateOptions struct {
		FileContent *[]byte
		Filename    string
		SHA         string
		Path        string
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

	RepositoryUpdate struct {
		Repository        *github.Repository
		Files             []*RepositoryFileUpdate
		RepositoryOptions *RepositoryUpdateOptions
		TemplateVars      TemplateVars
		PullRequestURL    string
	}

	RepositoryFileUpdate struct {
		Workflow                *gh.GithubWorkflow
		Dependabot              *dependabot.GithubDependabot
		RepositoryUpdateOptions *RepositoryFileUpdateOptions
	}
)
