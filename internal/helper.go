package internal

import (
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"sync"

	"github.com/google/go-github/v32/github"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
)

var (
	DefaultWorkflowDir  = "workflows"
	BranchNamePattern   = "ghconfig/workflows/%s"
	GhConfigWorkflowDir = ".ghconfig/%s"
	GithubConfigDir     = ".github"
)

func FindWorkflows(dirPath string) ([]*WorkflowTemplate, error) {
	templates := []*WorkflowTemplate{}
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		ext := filepath.Ext(file.Name())
		if ext != ".yml" && ext != ".yaml" {
			continue
		}
		workflowName := file.Name()
		filePath := path.Join(dirPath, workflowName)
		bytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			kingpin.Errorf("could not read workflow file %v.", filePath)
			continue
		}
		if len(bytes) == 0 {
			continue
		}
		t := Workflow{}
		err = yaml.Unmarshal(bytes, &t)
		if err != nil {
			kingpin.Errorf("workflow file %v can't be parsed as workflow.", filePath)
			continue
		}
		templates = append(templates, &WorkflowTemplate{
			Workflow: &t,
			// restore to .github/workflows structure to update the correct file in the repo
			FilePath: path.Join(GithubConfigDir, DefaultWorkflowDir, workflowName),
			FileName: workflowName,
		})
	}
	return templates, nil
}

func FindPatches(dirPath string) ([]*PatchData, error) {
	patches := []*PatchData{}
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		ext := filepath.Ext(file.Name())
		if ext != ".yml" && ext != ".yaml" {
			continue
		}
		filePath := path.Join(dirPath, file.Name())
		bytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			kingpin.Errorf("could not read patch file %v.", filePath)
			continue
		}
		if len(bytes) == 0 {
			kingpin.Errorf("file %v was empty.", filePath)
			continue
		}
		patchData := PatchData{}
		err = yaml.Unmarshal(bytes, &patchData)
		if err != nil {
			kingpin.Errorf("file %v can't be parsed as patch file.", filePath)
			continue
		}
		patches = append(patches, &patchData)

	}
	return patches, nil
}

func CreatePR(opts *Config, intent *UpdateWorkflowIntent) (string, error) {
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
		return "", fmt.Errorf("could not find a ref on the base branch")
	}

	err = UpdateRepositoryFiles(opts, intent)
	if err != nil {
		return "", err
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

func UpdateRepositoryFiles(opts *Config, intent *UpdateWorkflowIntent) error {
	for _, draft := range intent.WorkflowDrafts {
		// commit message
		commitMsg := "Update workflow files by ghconfig"

		rr, _, err := opts.GithubClient.Repositories.UpdateFile(
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
			return nil
		}
		draft.Url = rr.GetHTMLURL()
	}
	return nil
}

func FetchAllRepos(opts *Config) ([]*github.Repository, error) {
	me, _, err := opts.GithubClient.Users.Get(opts.Context, "")
	if err != nil {
		return nil, err
	}

	allRepos := []*github.Repository{}
	query := "user:" + *me.Login

	if opts.RepositoryQuery != "" {
		query = opts.RepositoryQuery + " in:name"
	}

	fetch := func(page int) (*github.RepositoriesSearchResult, *github.Response, error) {
		return opts.GithubClient.Search.Repositories(
			opts.Context,
			query,
			&github.SearchOptions{
				ListOptions: github.ListOptions{PerPage: 120, Page: page},
			},
		)
	}

	searchResult, resp, err := fetch(1)
	if err != nil {
		return nil, err
	}
	allRepos = append(allRepos, searchResult.Repositories...)

	var wg sync.WaitGroup
	var results = make(chan *[]*github.Repository, resp.LastPage)

	for i := 2; i <= resp.LastPage; i++ {
		wg.Add(1)
		cpage := i
		go func() {
			defer wg.Done()
			repos, _, err := fetch(cpage)
			if err != nil {
				fmt.Printf("error fetching repositories from github: %v", err)
			}
			results <- &repos.Repositories
		}()
	}

	wg.Wait()

	close(results)

	for repos := range results {
		allRepos = append(allRepos, *repos...)
	}

	return allRepos, nil
}
