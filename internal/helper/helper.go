package helper

import (
	"bytes"
	"fmt"
	"ghconfig/internal/config"
	"ghconfig/internal/dependabot"
	gh "ghconfig/internal/github"
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/Masterminds/sprig"
	"github.com/apex/log"
	"github.com/google/go-github/v32/github"
	"github.com/pieterclaerhout/go-waitgroup"
	"gopkg.in/yaml.v2"
)

func FindDependabot(dirPath string) (*config.DependabotTemplate, error) {
	tpl := &config.DependabotTemplate{}
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return tpl, nil
	}

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fileName := file.Name()

		switch file.Name() {
		case "dependabot.yml":
			fallthrough
		case "dependabot.yaml":
			filePath := path.Join(dirPath, fileName)
			bytes, err := ioutil.ReadFile(filePath)
			if err != nil {
				log.WithError(err).Errorf("could not read dependabot file: %v", filePath)
				continue
			}
			if len(bytes) == 0 {
				log.Infof("dependabot file is empty: %v", filePath)
				continue
			}
			dependabot := &dependabot.GithubDependabot{}
			err = yaml.Unmarshal(bytes, dependabot)
			if err != nil {
				log.WithError(err).Errorf("could not parse workflow file: %v", filePath)
				continue
			}
			tpl.RepositoryPath = path.Join(config.GithubConfigBaseDir, fileName)
			tpl.Dependabot = dependabot
			tpl.Filename = fileName

			return tpl, nil
		}
	}
	return nil, nil
}

// FindHealthFiles returns all github health files in the specified repository
// https://docs.github.com/en/github/building-a-strong-community/creating-a-default-community-health-file
func FindHealthFiles(dirPath string) ([]*config.GithubHealthFile, error) {
	healthFiles := []*config.GithubHealthFile{}
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return healthFiles, nil
	}

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fileName := file.Name()

		switch file.Name() {
		case "CODE_OF_CONDUCT.md":
			fallthrough
		case "FUNDING.md":
			fallthrough
		case "SECURITY.md":
			fallthrough
		case "CONTRIBUTING.md":
			fallthrough
		case "SUPPORT.md":
			filePath := path.Join(dirPath, fileName)
			bytes, err := ioutil.ReadFile(filePath)
			if err != nil {
				log.WithError(err).Errorf("could not read health file: %v", filePath)
				continue
			}
			if len(bytes) == 0 {
				log.Infof("health file is empty: %v", filePath)
				continue
			}
			healthFiles = append(healthFiles, &config.GithubHealthFile{
				Path:        path.Join(config.GithubConfigBaseDir, fileName),
				FileContent: &bytes,
				Filename:    fileName,
			})
		}
	}
	return healthFiles, nil
}

func FindWorkflows(dirPath string) ([]*config.WorkflowTemplate, error) {
	templates := []*config.WorkflowTemplate{}
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return templates, nil
	}

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
			log.WithError(err).Errorf("could not read workflow file: %v", filePath)
			continue
		}
		if len(bytes) == 0 {
			log.Infof("workflow file is empty: %v", filePath)
			continue
		}
		t := gh.GithubWorkflow{}
		err = yaml.Unmarshal(bytes, &t)
		if err != nil {
			log.WithError(err).Errorf("could not parse workflow file: %v", filePath)
			continue
		}
		templates = append(templates, &config.WorkflowTemplate{
			Workflow:       &t,
			RepositoryPath: path.Join(config.GithubConfigBaseDir, config.GhWorkflowDir, workflowName),
			Filename:       workflowName,
		})
	}
	return templates, nil
}

func FindPatches(dirPath string) ([]*config.PatchData, error) {
	patches := []*config.PatchData{}

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return patches, nil
	}

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
			log.WithError(err).Errorf("could not read patch file %v", filePath)
			continue
		}
		if len(bytes) == 0 {
			log.Infof("patch file: %v was empty.", filePath)
			continue
		}
		patchData := config.PatchData{}
		err = yaml.Unmarshal(bytes, &patchData)
		if err != nil {
			log.WithError(err).Errorf("could not parse patch file: %v", filePath)
			continue
		}
		patches = append(patches, &patchData)

	}
	return patches, nil
}

func CreatePR(opts *config.Config, intent *config.RepositoryUpdate) (string, error) {
	// get ref to branch from
	refs, _, err := opts.GithubClient.Git.ListMatchingRefs(
		opts.Context,
		intent.RepositoryOptions.Owner,
		intent.RepositoryOptions.Repo,
		&github.ReferenceListOptions{
			Ref: "heads/" + intent.RepositoryOptions.BaseRef,
		},
	)

	if err != nil {
		return "", err
	}

	if len(refs) > 0 {
		// create branch
		_, _, err := opts.GithubClient.Git.CreateRef(
			opts.Context,
			intent.RepositoryOptions.Owner,
			intent.RepositoryOptions.Repo,
			&github.Reference{
				// the name of the new branch
				Ref: &intent.RepositoryOptions.PRBranchRef,
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

	err = UpdateRepositoryFiles(opts, intent.RepositoryOptions, intent.Files)
	if err != nil {
		return "", err
	}

	draft := true

	commitMsg := "Synchronize (.github) configurations by ghconfig"

	pr, _, err := opts.GithubClient.PullRequests.Create(
		opts.Context,
		intent.RepositoryOptions.Owner,
		intent.RepositoryOptions.Repo,
		&github.NewPullRequest{
			Base:  &intent.RepositoryOptions.BaseRef,
			Title: &commitMsg,
			Draft: &draft,
			Head:  &intent.RepositoryOptions.Branch,
		},
	)
	if err != nil {
		return "", err
	}

	return pr.GetHTMLURL(), nil
}

func UpdateRepositoryFiles(opts *config.Config, updateOptions *config.RepositoryUpdateOptions, files []*config.RepositoryFileUpdate) error {
	for _, file := range files {
		// commit message
		commitMsg := fmt.Sprintf("Update %v file by ghconfig", file.RepositoryUpdateOptions.DisplayName)

		rr, _, err := opts.GithubClient.Repositories.UpdateFile(
			opts.Context,
			updateOptions.Owner,
			updateOptions.Repo,
			file.RepositoryUpdateOptions.Path,
			&github.RepositoryContentFileOptions{
				Branch:  &updateOptions.Branch,
				Message: &commitMsg,
				Content: *file.RepositoryUpdateOptions.FileContent,
				SHA:     &file.RepositoryUpdateOptions.SHA,
			},
		)
		if err != nil {
			return err
		}
		file.RepositoryUpdateOptions.URL = rr.GetHTMLURL()
	}
	return nil
}

func FetchAllRepos(opts *config.Config) ([]*github.Repository, error) {
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

	wg := waitgroup.NewWaitGroup(4)
	var results = make(chan *[]*github.Repository, resp.LastPage)

	for i := 2; i <= resp.LastPage; i++ {
		cpage := i
		wg.Add(func() {
			defer wg.Done()
			repos, _, err := fetch(cpage)
			if err != nil {
				fmt.Printf("error fetching repositories from github: %v", err)
			}
			results <- &repos.Repositories
		})
	}

	wg.Wait()

	close(results)

	for repos := range results {
		allRepos = append(allRepos, *repos...)
	}

	return allRepos, nil
}

func ExecuteTemplate(name string, text string, templateVars config.TemplateVars) (*bytes.Buffer, error) {
	t := template.Must(template.New(name).
		Delims("$((", "))").
		Funcs(sprig.FuncMap()).
		Parse(text))

	bytesCache := new(bytes.Buffer)
	err := t.Execute(bytesCache, templateVars)
	if err != nil {
		log.WithError(err).Error("could not execute template")
		return nil, err
	}

	return bytesCache, nil
}
