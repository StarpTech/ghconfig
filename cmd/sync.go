package cmd

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/Masterminds/sprig"
	"github.com/briandowns/spinner"
	"github.com/cheynewallace/tabby"
	"github.com/google/go-github/v32/github"
	"github.com/teris-io/shortid"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
)

var (
	branchName   = "ghconfig/workflows/%s"
	workflowsDir = ".ghconfig/workflows"
)

type WorkflowTemplate struct {
	Workflow *Workflow
	FileName string
	FilePath string
}

type UpdateWorkflowIntentOptions struct {
	Branch      string
	PRBranchRef string
	BaseRef     string
	Owner       string
	Repo        string
	PRExists    bool
}

type UpdateWorkflowIntent struct {
	RepositoryName string
	WorkflowDrafts []*WorkflowFileDraft
	Options        UpdateWorkflowIntentOptions
}

type WorkflowFileDraft struct {
	Workflow    *Workflow
	FileContent *[]byte
	Filename    string
	SHA         string
	FilePath    string
}

type Config struct {
	GithubClient *github.Client
	Context      context.Context
	DryRun       bool
	BaseBranch   string
	Sid          *shortid.Shortid
}

func NewSyncCmd(opts *Config) error {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Collecting all available repositories..."
	s.Start()
	repos, err := fetchAllRepos(opts.Context, opts.GithubClient)
	s.Stop()
	if err != nil {
		return err
	}

	pDir, err := os.Getwd()
	if err != nil {
		return err
	}

	reposNames := []string{}

	for _, repo := range repos {
		reposNames = append(reposNames, *repo.FullName)
	}
	workflowDir := path.Join(pDir, workflowsDir)
	workflowFiles, err := ioutil.ReadDir(workflowDir)
	if err != nil {
		return err
	}
	templates := []WorkflowTemplate{}

	for _, workflowFile := range workflowFiles {
		if workflowFile.IsDir() {
			continue
		}
		filePath := path.Join(workflowDir, workflowFile.Name())
		bytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			return err
		}
		t := Workflow{}
		err = yaml.Unmarshal(bytes, &t)
		if err != nil {
			return err
		}
		templates = append(templates, WorkflowTemplate{
			Workflow: &t,
			FilePath: filePath,
			FileName: workflowFile.Name(),
		})
	}

	targetRepoPrompt := &survey.MultiSelect{
		Message:  "Please select all target repositories.",
		PageSize: 20,
		Options:  reposNames,
	}

	targetRepos := []string{}
	survey.AskOne(targetRepoPrompt, &targetRepos)

	intents := []*UpdateWorkflowIntent{}

	s = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Update workflow files and create pull requests..."

	s.Start()

	t := tabby.New()
	t.AddHeader("Repository", "Changes", "Url")

	for _, repoFullName := range targetRepos {
		repo := getRepoByName(repos, repoFullName)
		intent, err := collectWorkflowFiles(opts, repo, templates)
		if err != nil {
			kingpin.Errorf("could not update workflow files, Repo: %v, error: %v", repoFullName, err)
			continue
		}

		if len(intent.WorkflowDrafts) > 0 {
			url := "dry-run"
			if !opts.DryRun {
				url, err = createPR(opts, intent)
				if err != nil {
					kingpin.Errorf("could not create PR with changes: %v", err)
					return err
				}
			}

			changelog := []string{}
			for _, draft := range intent.WorkflowDrafts {
				changelog = append(changelog, draft.Filename)
			}
			t.AddLine(repo.GetFullName(), strings.Join(changelog, ","), url)

		} else {
			fmt.Printf("\nSkip: No templates found in %v.\n", workflowsDir)
		}

		intents = append(intents, intent)
	}

	s.Stop()
	fmt.Println()
	t.Print()

	if opts.DryRun {
		file, err := os.Create(path.Join(pDir, "ghconfig-debug.yml"))
		if err != nil {
			kingpin.Errorf("could not create ghconfig-debug.yml: %v", err)
		}
		defer file.Close()

		for _, wr := range intents {
			for _, draft := range wr.WorkflowDrafts {
				y, err := yaml.Marshal(draft.Workflow)
				if err != nil {
					continue
				}
				_, err = file.Write([]byte(fmt.Sprintf("\n# Repository: %v, Workflow: %v\n%v\n---", wr.RepositoryName, draft.Filename, string(y))))
				if err != nil {
					kingpin.Errorf("could not write to ghconfig-debug.yml: %v", err)
				}
			}
		}
	}

	return nil
}

func collectWorkflowFiles(opts *Config, repo *github.Repository, templates []WorkflowTemplate) (*UpdateWorkflowIntent, error) {
	branchName := fmt.Sprintf(branchName, opts.Sid.MustGenerate())
	updateOptions := UpdateWorkflowIntentOptions{
		Owner:       *repo.GetOwner().Login,
		Repo:        repo.GetName(),
		BaseRef:     opts.BaseBranch,
		Branch:      branchName,
		PRBranchRef: "refs/heads/" + branchName,
	}

	_, dirContent, _, err := opts.GithubClient.Repositories.GetContents(
		opts.Context,
		updateOptions.Owner,
		updateOptions.Repo,
		".github/workflows",
		&github.RepositoryContentGetOptions{
			Ref: updateOptions.BaseRef,
		},
	)
	if err != nil {
		kingpin.Errorf("could not list workflow directory, %v", err)
		return nil, err
	}

	workflowDrafts := []*WorkflowFileDraft{}
	templateVars := map[string]interface{}{"Repo": repo}

	for _, workflowTemplate := range templates {
		var draft *WorkflowFileDraft

		bytesCache, err := templateWorkflow(repo.GetFullName(), workflowTemplate.Workflow, templateVars)
		if err != nil {
			kingpin.Errorf("could not template, %v", err)
			continue
		}

		proceedTemplate := Workflow{}
		fileContent := bytesCache.Bytes()
		err = yaml.Unmarshal(bytesCache.Bytes(), &proceedTemplate)
		if err != nil {
			kingpin.Errorf("could not unmarshal template, %v", err)
			continue
		}

		for _, content := range dirContent {
			// match based on filename, workflow name is not unique
			if content.GetName() == workflowTemplate.FileName {
				draft = &WorkflowFileDraft{}
				draft.Filename = content.GetName()
				draft.Workflow = &proceedTemplate
				draft.FileContent = &fileContent
				draft.FilePath = content.GetPath()
				draft.SHA = content.GetSHA()
			}
		}
		if draft == nil {
			draft = &WorkflowFileDraft{}
			draft.Filename = workflowTemplate.FileName
			draft.Workflow = &proceedTemplate
			draft.FileContent = &fileContent
			draft.FilePath = workflowTemplate.FilePath
		}

		workflowDrafts = append(workflowDrafts, draft)
	}

	return &UpdateWorkflowIntent{
		WorkflowDrafts: workflowDrafts,
		Options:        updateOptions,
		RepositoryName: repo.GetFullName(),
	}, nil
}

func templateWorkflow(name string, workflow *Workflow, templateVars map[string]interface{}) (*bytes.Buffer, error) {
	y, err := yaml.Marshal(workflow)
	if err != nil {
		kingpin.Errorf("could not marshal template, %v", err)
		return nil, err
	}

	t := template.Must(template.New(name).
		Delims("$((", "))").
		Funcs(sprig.FuncMap()).
		Parse(string(y)))

	bytesCache := new(bytes.Buffer)
	err = t.Execute(bytesCache, templateVars)
	if err != nil {
		kingpin.Errorf("could not template, %v", err)
		return nil, err
	}

	return bytesCache, nil
}

func createPR(opts *Config, intent *UpdateWorkflowIntent) (string, error) {
	// get ref to branch from
	refs, _, err := opts.GithubClient.Git.ListMatchingRefs(
		opts.Context,
		intent.Options.Owner,
		intent.Options.Repo,
		&github.ReferenceListOptions{
			Ref: "heads/" + intent.Options.BaseRef,
		},
	)

	if err != nil {
		return "", err
	}

	if len(refs) > 0 {
		// create branch
		_, _, err = opts.GithubClient.Git.CreateRef(
			opts.Context,
			intent.Options.Owner,
			intent.Options.Repo,
			&github.Reference{
				// the name of the new branch
				Ref: &intent.Options.PRBranchRef,
				// branch from master
				Object: &github.GitObject{SHA: refs[0].Object.SHA},
			},
		)
		if err != nil {
			return "", err
		}
	} else {
		return "", fmt.Errorf("could not find a ref on base branch")
	}

	for _, draft := range intent.WorkflowDrafts {
		// commit message
		commitMsg := "Update workflow files by ghconfig"

		_, _, err := opts.GithubClient.Repositories.UpdateFile(
			opts.Context,
			intent.Options.Owner,
			intent.Options.Repo,
			draft.FilePath,
			&github.RepositoryContentFileOptions{
				Branch:  &intent.Options.Branch,
				Message: &commitMsg,
				Content: *draft.FileContent,
				SHA:     &draft.SHA,
			},
		)
		if err != nil {
			return "", err
		}
	}

	draft := true

	commitMsg := "Update workflows by ghconfig"

	pr, _, err := opts.GithubClient.PullRequests.Create(
		opts.Context,
		intent.Options.Owner,
		intent.Options.Repo,
		&github.NewPullRequest{
			Base:  &intent.Options.BaseRef,
			Title: &commitMsg,
			Draft: &draft,
			Head:  &intent.Options.Branch,
		},
	)
	if err != nil {
		return "", err
	}

	return pr.GetHTMLURL(), nil
}

func fetchAllRepos(ctx context.Context, client *github.Client) ([]*github.Repository, error) {
	allRepos := []*github.Repository{}

	fetch := func(page int) ([]*github.Repository, *github.Response, error) {
		return client.Repositories.List(
			ctx,
			"",
			&github.RepositoryListOptions{
				ListOptions: github.ListOptions{PerPage: 100, Page: page},
			},
		)
	}

	repos, resp, err := fetch(1)
	if err != nil {
		return allRepos, nil
	}
	allRepos = append(allRepos, repos...)

	var wg sync.WaitGroup
	var results = make(chan *[]*github.Repository, resp.LastPage-1)

	for i := 2; i <= resp.LastPage; i++ {
		wg.Add(1)
		cpage := i
		go func() {
			defer wg.Done()
			repos, _, err := fetch(cpage)
			if err != nil {
				fmt.Printf("error fetching repositories from github: %v", err)
			}
			results <- &repos
		}()
	}

	wg.Wait()

	close(results)

	for repos := range results {
		allRepos = append(allRepos, *repos...)
	}

	return allRepos, nil
}

func getRepoByName(repos []*github.Repository, name string) *github.Repository {
	for _, repo := range repos {
		if name == *repo.FullName {
			return repo
		}
	}
	return nil
}
