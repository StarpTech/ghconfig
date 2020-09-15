package cmd

import (
	"encoding/json"
	"fmt"
	"ghconfig/internal/config"
	"ghconfig/internal/dependabot"
	gh "ghconfig/internal/github"
	"ghconfig/internal/helper"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/apex/log"
	"github.com/briandowns/spinner"
	"github.com/cheynewallace/tabby"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/google/go-github/v32/github"
	"gopkg.in/yaml.v2"
)

type (
	GetRepositoryList func(reposNames []string) []string
)

func StubRepositorySelection(result []string) func() {
	orig := Multiselect
	Multiselect = func(_ []string, r *[]string) error {
		*r = result
		return nil
	}
	return func() {
		Multiselect = orig
	}
}

var Multiselect = func(reposNames []string, targetRepos *[]string) error {
	targetRepoPrompt := &survey.MultiSelect{
		Message:  "Please select all repositories:",
		PageSize: 20,
		Options:  reposNames,
	}
	return survey.AskOne(targetRepoPrompt, targetRepos)
}

func NewSyncCmd(globalOptions *config.Config) error {
	workflowDirAbs := path.Join(globalOptions.RootDir, config.GhConfigBaseDir, config.GhWorkflowDir)
	templates, err := helper.FindWorkflows(workflowDirAbs)
	if err != nil {
		return err
	}

	workflowPatchesDirAbs := path.Join(globalOptions.RootDir, config.GhConfigBaseDir, config.GhWorkflowDir, config.GhPatchesDir)
	patches, err := helper.FindPatches(workflowPatchesDirAbs)
	if err != nil {
		return err
	}

	healthFilesAbs := path.Join(globalOptions.RootDir, config.GhConfigBaseDir)
	healthFiles, err := helper.FindHealthFiles(healthFilesAbs)
	if err != nil {
		return err
	}

	dependabotAbs := path.Join(globalOptions.RootDir, config.GhConfigBaseDir)
	dependabotTemplate, err := helper.FindDependabot(dependabotAbs)
	if err != nil {
		return err
	}

	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Collecting all available repositories..."
	s.Start()

	repos, err := helper.FetchAllRepos(globalOptions)
	if err != nil {
		return err
	}
	s.Stop()

	reposNames := []string{}

	for _, repo := range repos {
		reposNames = append(reposNames, *repo.FullName)
	}

	targetRepos := []string{}
	err = Multiselect(reposNames, &targetRepos)
	if err != nil {
		log.WithError(err).Error("could not create multi select for repository selection")
		return err
	}

	packages := []*config.RepositoryUpdate{}

	s = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Update workflow files and create pull requests..."

	s.Start()

	t := tabby.New()
	t.AddHeader("Repository", "Files", "Url")

	for _, repoFullName := range targetRepos {
		repo := getRepoByName(repos, repoFullName)
		branchName := globalOptions.BaseBranch

		ctx := log.WithFields(log.Fields{
			"repository": repoFullName,
		})

		if globalOptions.CreatePR {
			branchName = fmt.Sprintf(config.BranchNamePattern, globalOptions.Sid.MustGenerate())
		}

		updateOptions := &config.RepositoryUpdateOptions{
			Owner:       *repo.GetOwner().Login,
			Repo:        repo.GetName(),
			BaseRef:     globalOptions.BaseBranch,
			Branch:      branchName,
			PRBranchRef: "refs/heads/" + branchName,
		}

		update := &config.RepositoryUpdate{
			RepositoryOptions: updateOptions,
			Repository:        repo,
			TemplateVars:      map[string]interface{}{"Repo": repo},
		}

		files, err := prepareFileUpdates(globalOptions, update, templates, patches, healthFiles)
		if err != nil {
			ctx.WithError(err).Error("could not prepare workflow files")
			continue
		}

		if dependabotTemplate != nil {
			fileUpdate, err := prepareDependabot(globalOptions, update, dependabotTemplate)
			if err != nil {
				log.WithError(err).Error("could not prepare dependabot file")
				continue
			}

			update.Files = append(files, fileUpdate)
		}

		if len(update.Files) > 0 {
			pullRequestURL := ""
			if !globalOptions.DryRun {
				if globalOptions.CreatePR {
					pullRequestURL, err = helper.CreatePR(globalOptions, update)
					if err != nil {
						ctx.WithError(err).Error("could not create PR with changes")
						continue
					}
				} else {
					// update files directly on the base branch
					err := helper.UpdateRepositoryFiles(globalOptions, update.RepositoryOptions, update.Files)
					if err != nil {
						ctx.WithError(err).Error("could not update files on remote")
						continue
					}
				}
			}

			// build table for cli output
			for i, file := range update.Files {
				if i == 0 {
					url := ""
					if pullRequestURL != "" {
						url = pullRequestURL
					} else {
						url = file.RepositoryUpdateOptions.URL
					}
					t.AddLine(repo.GetFullName(), file.RepositoryUpdateOptions.DisplayName, url)
				} else {
					url := ""
					if pullRequestURL == "" {
						url = file.RepositoryUpdateOptions.URL
					}
					t.AddLine("", file.RepositoryUpdateOptions.DisplayName, url)
				}
			}

		}

		packages = append(packages, update)
	}

	s.Stop()
	fmt.Println()
	t.Print()

	if globalOptions.DryRun {
		file, err := os.Create(path.Join(globalOptions.RootDir, "ghconfig-debug.yml"))
		if err != nil {
			log.WithError(err).Error("could not create ghconfig-debug.yml")
		}
		defer file.Close()

		for _, wr := range packages {
			for _, f := range wr.Files {
				if f.Workflow != nil {
					y, err := yaml.Marshal(f.Workflow)
					if err != nil {
						log.WithError(err).Errorf("could not marshal %v", f.RepositoryUpdateOptions.Filename)
						continue
					}
					_, err = file.Write([]byte(fmt.Sprintf("\n# Repository: %v, File: %v\n%v\n---", wr.Repository.GetFullName(), f.RepositoryUpdateOptions.Path, string(y))))
					if err != nil {
						log.WithError(err).Error("could not write to ghconfig-debug.yml")
					}
				} else if f.Dependabot != nil {
					y, err := yaml.Marshal(f.Dependabot)
					if err != nil {
						log.WithError(err).Errorf("could not marshal %v", f.RepositoryUpdateOptions.Filename)
						continue
					}
					_, err = file.Write([]byte(fmt.Sprintf("\n# Repository: %v, File: %v\n%v\n---", wr.Repository.GetFullName(), f.RepositoryUpdateOptions.Path, string(y))))
					if err != nil {
						log.WithError(err).Error("could not write to ghconfig-debug.yml")
					}
				} else {
					yamlFile := config.FileYAML{
						Content: string(*f.RepositoryUpdateOptions.FileContent),
					}
					y, err := yaml.Marshal(yamlFile)
					if err != nil {
						log.WithError(err).Errorf("could not marshal %v", f.RepositoryUpdateOptions.Filename)
						continue
					}
					_, err = file.Write([]byte(fmt.Sprintf("\n# Repository: %v, File: %v\n%v\n---", wr.Repository.GetFullName(), f.RepositoryUpdateOptions.Path, string(y))))
					if err != nil {
						log.WithError(err).Error("could not write to ghconfig-debug.yml")
					}
				}
			}
		}
	}

	return nil
}

func prepareFileUpdates(
	opts *config.Config,
	update *config.RepositoryUpdate,
	templates []*config.WorkflowTemplate,
	patches []*config.PatchData,
	healthFiles []*config.GithubHealthFile) ([]*config.RepositoryFileUpdate, error) {

	directory := path.Join(config.GithubConfigBaseDir, "workflows")
	_, dirContent, resp, err := opts.GithubClient.Repositories.GetContents(
		opts.Context,
		update.RepositoryOptions.Owner,
		update.RepositoryOptions.Repo,
		directory,
		&github.RepositoryContentGetOptions{
			Ref: update.RepositoryOptions.BaseRef,
		},
	)

	files := []*config.RepositoryFileUpdate{}

	if err != nil && resp.StatusCode != 404 {
		log.Debugf("workflow directory %v doesn't exist on remote", directory)
		return nil, err
	}

	for _, workflowTemplate := range templates {
		var file *config.RepositoryFileUpdate

		y, err := yaml.Marshal(workflowTemplate.Workflow)
		if err != nil {
			log.WithError(err).Error("could not marshal template")
			continue
		}

		bytesCache, err := helper.ExecuteTemplate(workflowTemplate.Filename, string(y), update.TemplateVars)
		if err != nil {
			log.WithError(err).Error("could not template")
			continue
		}

		proceedTemplate := gh.GithubWorkflow{}
		fileContent := bytesCache.Bytes()
		err = yaml.Unmarshal(fileContent, &proceedTemplate)
		if err != nil {
			log.WithError(err).Error("could unmarshal template")
			continue
		}

		for _, content := range dirContent {
			if content.GetName() == workflowTemplate.Filename {
				file = &config.RepositoryFileUpdate{}
				file.RepositoryUpdateOptions = &config.RepositoryFileUpdateOptions{}
				file.RepositoryUpdateOptions.Filename = content.GetName()
				file.RepositoryUpdateOptions.DisplayName = file.RepositoryUpdateOptions.Filename
				file.Workflow = &proceedTemplate
				file.RepositoryUpdateOptions.FileContent = &fileContent
				file.RepositoryUpdateOptions.Path = content.GetPath()
				file.RepositoryUpdateOptions.SHA = content.GetSHA()
			}
		}
		if file == nil {
			file = &config.RepositoryFileUpdate{}
			file.RepositoryUpdateOptions = &config.RepositoryFileUpdateOptions{}
			file.RepositoryUpdateOptions.Filename = workflowTemplate.Filename
			file.RepositoryUpdateOptions.DisplayName = workflowTemplate.Filename
			file.Workflow = &proceedTemplate
			file.RepositoryUpdateOptions.FileContent = &fileContent
			file.RepositoryUpdateOptions.Path = workflowTemplate.RepositoryPath
		}

		files = append(files, file)
	}

	for _, patch := range patches {
		remoteFilePath := path.Join(config.GithubConfigBaseDir, "workflows", patch.Filename)
		content, _, resp, err := opts.GithubClient.Repositories.GetContents(
			opts.Context,
			update.RepositoryOptions.Owner,
			update.RepositoryOptions.Repo,
			remoteFilePath,
			&github.RepositoryContentGetOptions{
				Ref: update.RepositoryOptions.BaseRef,
			},
		)
		if err != nil {
			if resp.StatusCode == 404 {
				log.Debugf("worklfow file %v doesn't exist on remote", remoteFilePath)
				continue
			}
			log.WithError(err).Error("could not list workflow file")
			return nil, err
		}
		fileResp, err := http.Get(content.GetDownloadURL())
		if err != nil {
			log.WithError(err).Error("could not download file")
			continue
		}
		defer fileResp.Body.Close()

		data, err := ioutil.ReadAll(fileResp.Body)
		if err != nil {
			log.WithError(err).Error("could not read file response")
			continue
		}

		t := gh.GithubWorkflow{}
		err = yaml.Unmarshal(data, &t)
		if err != nil {
			log.WithError(err).Error("could not unmarshal workflow")
			continue
		}

		repositoryFileJSON, err := json.Marshal(t)
		if err != nil {
			log.WithError(err).Error("could not convert yaml to json")
			continue
		}

		y, err := yaml.Marshal(patch)
		if err != nil {
			log.WithError(err).Error("could not marshal template")
			return nil, err
		}

		jsonPatchData, err := helper.ExecuteTemplate(patch.Filename, string(y), update.TemplateVars)
		if err != nil {
			log.WithError(err).Error("could not template")
			continue
		}

		newPatchData := config.PatchData{}
		err = yaml.Unmarshal(jsonPatchData.Bytes(), &newPatchData)
		if err != nil {
			log.WithError(err).Error("could not unmarshal template")
			continue
		}

		templatedJSONPatchFile, err := json.Marshal(newPatchData.Patch)
		if err != nil {
			log.WithError(err).Error("could not marshal templated patch file to json")
			continue
		}

		jsonPatch, err := jsonpatch.DecodePatch(templatedJSONPatchFile)
		if err != nil {
			log.WithError(err).Error("invalid patch file")
			continue
		}

		data, err = jsonPatch.Apply(repositoryFileJSON)
		if err != nil {
			log.WithError(err).Error("could not apply patch")
			continue
		}

		t = gh.GithubWorkflow{}
		err = yaml.Unmarshal(data, &t)
		if err != nil {
			log.WithError(err).Error("could not unmarshal patched workflow")
			continue
		}
		repositoryFileJSON, err = yaml.Marshal(&t)
		if err != nil {
			log.WithError(err).Error("could not marshal patched workflow")
			continue
		}

		file := &config.RepositoryFileUpdate{}
		file.RepositoryUpdateOptions = &config.RepositoryFileUpdateOptions{}
		file.RepositoryUpdateOptions.Filename = content.GetName()
		file.RepositoryUpdateOptions.DisplayName = content.GetName() + " (patched)"
		file.Workflow = &t
		file.RepositoryUpdateOptions.FileContent = &repositoryFileJSON
		file.RepositoryUpdateOptions.Path = content.GetPath()
		file.RepositoryUpdateOptions.SHA = content.GetSHA()

		files = append(files, file)

	}

	for _, healthFile := range healthFiles {
		file := &config.RepositoryFileUpdate{}
		remoteFilePath := path.Join(config.GithubConfigBaseDir, healthFile.Filename)
		content, _, resp, err := opts.GithubClient.Repositories.GetContents(
			opts.Context,
			update.RepositoryOptions.Owner,
			update.RepositoryOptions.Repo,
			remoteFilePath,
			&github.RepositoryContentGetOptions{
				Ref: update.RepositoryOptions.BaseRef,
			},
		)
		if err != nil {
			if resp.StatusCode == 404 {
				log.Debugf("health file %v doesn't exist", remoteFilePath)
				file.RepositoryUpdateOptions = &config.RepositoryFileUpdateOptions{}
				file.RepositoryUpdateOptions.Filename = healthFile.Filename
				file.RepositoryUpdateOptions.DisplayName = healthFile.Filename
				file.RepositoryUpdateOptions.FileContent = healthFile.FileContent
				file.RepositoryUpdateOptions.Path = remoteFilePath
				files = append(files, file)
				continue
			}
			log.WithError(err).Error("could not list health file")
			return nil, err
		}

		bytesCache, err := helper.ExecuteTemplate(healthFile.Filename, string(*healthFile.FileContent), update.TemplateVars)
		if err != nil {
			log.WithError(err).Error("could not template")
			continue
		}

		newFileContent := bytesCache.Bytes()

		file.RepositoryUpdateOptions = &config.RepositoryFileUpdateOptions{}
		file.RepositoryUpdateOptions.Filename = content.GetName()
		file.RepositoryUpdateOptions.DisplayName = content.GetName()
		file.RepositoryUpdateOptions.FileContent = &newFileContent
		file.RepositoryUpdateOptions.Path = content.GetPath()
		file.RepositoryUpdateOptions.SHA = content.GetSHA()

		files = append(files, file)
	}

	return files, nil
}

func prepareDependabot(
	opts *config.Config,
	update *config.RepositoryUpdate,
	dependabotTemplate *config.DependabotTemplate) (*config.RepositoryFileUpdate, error) {

	y, err := yaml.Marshal(dependabotTemplate.Dependabot)
	if err != nil {
		log.WithError(err).Error("could not marshal template")
		return nil, err
	}
	bytesCache, err := helper.ExecuteTemplate(dependabotTemplate.Filename, string(y), update.TemplateVars)
	if err != nil {
		log.WithError(err).Error("could not template")
		return nil, err
	}
	proceedTemplate := dependabot.GithubDependabot{}
	fileContent := bytesCache.Bytes()
	err = yaml.Unmarshal(fileContent, &proceedTemplate)
	if err != nil {
		log.WithError(err).Error("could not unmarshal template")
		return nil, err
	}

	file := &config.RepositoryFileUpdate{}
	remoteFilePath := path.Join(config.GithubConfigBaseDir, dependabotTemplate.Filename)
	content, _, resp, err := opts.GithubClient.Repositories.GetContents(
		opts.Context,
		update.RepositoryOptions.Owner,
		update.RepositoryOptions.Repo,
		remoteFilePath,
		&github.RepositoryContentGetOptions{
			Ref: update.RepositoryOptions.BaseRef,
		},
	)

	if err != nil {
		if resp.StatusCode == 404 {
			log.Debugf("dependabot file %v doesn't exist on remote", remoteFilePath)
			file.RepositoryUpdateOptions = &config.RepositoryFileUpdateOptions{}
			file.RepositoryUpdateOptions.Filename = dependabotTemplate.Filename
			file.RepositoryUpdateOptions.DisplayName = dependabotTemplate.Filename
			file.RepositoryUpdateOptions.FileContent = &fileContent
			file.RepositoryUpdateOptions.Path = remoteFilePath
			return file, nil
		}
		log.WithError(err).Error("could not list health file")
		return nil, err
	}

	file.RepositoryUpdateOptions = &config.RepositoryFileUpdateOptions{}
	file.RepositoryUpdateOptions.Filename = content.GetName()
	file.RepositoryUpdateOptions.DisplayName = file.RepositoryUpdateOptions.Filename
	file.Dependabot = &proceedTemplate
	file.RepositoryUpdateOptions.FileContent = &fileContent
	file.RepositoryUpdateOptions.Path = content.GetPath()
	file.RepositoryUpdateOptions.SHA = content.GetSHA()

	return file, nil
}

func getRepoByName(repos []*github.Repository, name string) *github.Repository {
	for _, repo := range repos {
		if name == *repo.FullName {
			return repo
		}
	}
	return nil
}
