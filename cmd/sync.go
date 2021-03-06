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
	"github.com/k0kubun/go-ansi"
	"github.com/pieterclaerhout/go-waitgroup"
	"github.com/schollz/progressbar/v3"
	"gopkg.in/yaml.v3"
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

	wg := waitgroup.NewWaitGroup(3)
	var results = make(chan *config.RepositoryUpdate, len(targetRepos))

	bar := progressbar.NewOptions(len(targetRepos),
		progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetWidth(15),
		progressbar.OptionSetDescription("Processing repositories..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	for _, name := range targetRepos {
		repositoryName := name
		wg.Add(func() {
			defer bar.Add(1)
			repo := getRepoByName(repos, repositoryName)
			branchName := globalOptions.BaseBranch

			ctx := log.WithFields(log.Fields{
				"repository": repositoryName,
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

			if !globalOptions.PatchOnly {
				files, err := prepareWorkflows(globalOptions, update, templates)
				if err != nil {
					ctx.WithError(err).Error("could not prepare workflow files")
					return
				}
				update.Files = append(update.Files, files...)

				if dependabotTemplate != nil {
					fileUpdate, err := prepareDependabot(globalOptions, update, dependabotTemplate)
					if err != nil {
						log.WithError(err).Error("could not prepare dependabot file")
						return
					}
					update.Files = append(update.Files, fileUpdate)
				}
			} else {
				files, err := preparePatches(globalOptions, update, patches)
				if err != nil {
					ctx.WithError(err).Error("could not prepare patch files")
					return
				}
				update.Files = append(update.Files, files...)
			}

			if len(update.Files) > 0 {
				if !globalOptions.DryRun {
					if globalOptions.CreatePR {
						pullRequestURL, err := helper.CreatePR(globalOptions, update)
						if err != nil {
							ctx.WithError(err).Error("could not create PR with changes")
							return
						}
						update.PullRequestURL = pullRequestURL
					} else {
						// update files directly on the base branch
						err := helper.UpdateRepositoryFiles(globalOptions, update.RepositoryOptions, update.Files)
						if err != nil {
							ctx.WithError(err).Error("could not update files on remote")
							return
						}
					}
				}
			}

			results <- update
		})
	}

	wg.Wait()
	close(results)

	s.Stop()

	updates := []*config.RepositoryUpdate{}
	for pkg := range results {
		updates = append(updates, pkg)
	}

	table := tabby.New()
	table.AddHeader("Repository", "Pull-Request")

	// build table for cli output
	for _, pkg := range updates {
		table.AddLine(pkg.Repository.GetFullName(), pkg.PullRequestURL)
	}

	fmt.Print("\n\n")
	table.Print()

	if globalOptions.DryRun {
		file, err := os.Create(path.Join(globalOptions.RootDir, "ghconfig-debug.yml"))
		if err != nil {
			log.WithError(err).Error("could not create ghconfig-debug.yml")
		}
		defer file.Close()

		for _, wr := range updates {
			for _, f := range wr.Files {
				if f.Workflow != nil {
					y, err := yaml.Marshal(f.Workflow)
					if err != nil {
						log.WithError(err).Errorf("could not marshal %v", f.RepositoryUpdateOptions.Filename)
						continue
					}
					_, err = file.Write([]byte(fmt.Sprintf("\n# Repository: %v, File: %v\n%v\n---", wr.Repository.GetFullName(), f.RepositoryUpdateOptions.DisplayName, string(y))))
					if err != nil {
						log.WithError(err).Error("could not write to ghconfig-debug.yml")
					}
				} else if f.Dependabot != nil {
					y, err := yaml.Marshal(f.Dependabot)
					if err != nil {
						log.WithError(err).Errorf("could not marshal %v", f.RepositoryUpdateOptions.Filename)
						continue
					}
					_, err = file.Write([]byte(fmt.Sprintf("\n# Repository: %v, File: %v\n%v\n---", wr.Repository.GetFullName(), f.RepositoryUpdateOptions.DisplayName, string(y))))
					if err != nil {
						log.WithError(err).Error("could not write to ghconfig-debug.yml")
					}
				}
			}
		}
		fmt.Printf("\nData has been written to ghconfig-debug.yml\n")
	}

	return nil
}

func preparePatches(opts *config.Config, update *config.RepositoryUpdate, patches []*config.PatchData) ([]*config.RepositoryFileUpdate, error) {
	files := []*config.RepositoryFileUpdate{}

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
			log.WithError(err).Errorf("could not download file: %v", content.GetDownloadURL())
			continue
		}
		if fileResp.StatusCode != 200 {
			log.WithError(fmt.Errorf("file: %v", content.GetDownloadURL())).Error("could not download file")
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
			log.WithError(err).Error("could not marshal patch")
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
	return files, nil
}

func prepareWorkflows(opts *config.Config, update *config.RepositoryUpdate, templates []*config.WorkflowTemplate) ([]*config.RepositoryFileUpdate, error) {
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

		locaTemplateData, err := yaml.Marshal(workflowTemplate.Workflow)
		if err != nil {
			log.WithError(err).Error("could not marshal template")
			continue
		}

		bytesCache, err := helper.ExecuteTemplate(workflowTemplate.Filename, string(locaTemplateData), update.TemplateVars)
		if err != nil {
			log.WithError(err).Error("could not template")
			continue
		}

		localTemplate := gh.GithubWorkflow{}
		templateBytes := bytesCache.Bytes()
		err = yaml.Unmarshal(templateBytes, &localTemplate)
		if err != nil {
			log.WithError(err).Error("could unmarshal template")
			continue
		}

		for _, content := range dirContent {
			if content.GetName() == workflowTemplate.Filename {
				content, _, resp, err := opts.GithubClient.Repositories.GetContents(
					opts.Context,
					update.RepositoryOptions.Owner,
					update.RepositoryOptions.Repo,
					workflowTemplate.RepositoryPath,
					&github.RepositoryContentGetOptions{
						Ref: update.RepositoryOptions.BaseRef,
					},
				)
				if err != nil {
					if resp.StatusCode == 404 {
						log.Debugf("worklfow file %v doesn't exist anymore on remote", workflowTemplate.RepositoryPath)
						continue
					}
					log.WithError(err).Error("could not list workflow file")
					continue
				}

				fileResp, err := http.Get(content.GetDownloadURL())
				if err != nil {
					log.WithError(err).Errorf("could not download file: %v", content.GetDownloadURL())
					continue
				}
				if fileResp.StatusCode != 200 {
					return nil, fmt.Errorf("could not download: %v", content.GetDownloadURL())
				}
				defer fileResp.Body.Close()

				remoteFileData, err := ioutil.ReadAll(fileResp.Body)
				if err != nil {
					log.WithError(err).Error("could not read file response")
					continue
				}

				remoteTemplate := gh.GithubWorkflow{}
				err = yaml.Unmarshal(remoteFileData, &remoteTemplate)
				if err != nil {
					log.WithError(err).Error("could not unmarshal template")
					continue
				}

				err = gh.MergeWorkflow(&remoteTemplate, localTemplate)
				if err != nil {
					log.WithError(err).Error("could not merge template")
					continue
				}

				output, err := yaml.Marshal(remoteTemplate)
				if err != nil {
					log.WithError(err).Error("could not marshal template")
					continue
				}

				file = &config.RepositoryFileUpdate{}
				file.RepositoryUpdateOptions = &config.RepositoryFileUpdateOptions{}
				file.RepositoryUpdateOptions.Filename = content.GetName()
				file.RepositoryUpdateOptions.DisplayName = file.RepositoryUpdateOptions.Filename
				file.Workflow = &remoteTemplate
				file.RepositoryUpdateOptions.FileContent = &output
				file.RepositoryUpdateOptions.Path = content.GetPath()
				file.RepositoryUpdateOptions.SHA = content.GetSHA()
				files = append(files, file)
				break
			}
		}
		if file == nil {
			file = &config.RepositoryFileUpdate{}
			file.RepositoryUpdateOptions = &config.RepositoryFileUpdateOptions{}
			file.RepositoryUpdateOptions.Filename = workflowTemplate.Filename
			file.RepositoryUpdateOptions.DisplayName = workflowTemplate.Filename
			file.Workflow = &localTemplate
			file.RepositoryUpdateOptions.FileContent = &templateBytes
			file.RepositoryUpdateOptions.Path = workflowTemplate.RepositoryPath
			files = append(files, file)
		}

	}
	return files, nil
}

func prepareDependabot(
	opts *config.Config, update *config.RepositoryUpdate, dependabotTemplate *config.DependabotTemplate) (*config.RepositoryFileUpdate, error) {

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
	localTemplate := dependabot.GithubDependabot{}
	localYAMLData := bytesCache.Bytes()
	err = yaml.Unmarshal(localYAMLData, &localTemplate)
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
			file.RepositoryUpdateOptions.FileContent = &localYAMLData
			file.RepositoryUpdateOptions.Path = remoteFilePath
			return file, nil
		}
		log.WithError(err).Error("could not list health file")
		return nil, err
	}
	url := content.GetDownloadURL()
	fileResp, err := http.Get(url)
	if err != nil {
		log.WithError(err).Errorf("could not download file: %v", content.GetDownloadURL())
		return nil, err
	}
	if fileResp.StatusCode != 200 {
		return nil, fmt.Errorf("could not download: %v", content.GetDownloadURL())
	}
	defer fileResp.Body.Close()

	remoteFileData, err := ioutil.ReadAll(fileResp.Body)
	if err != nil {
		log.WithError(err).Error("could not read file response")
		return nil, err
	}

	remoteTemplate := dependabot.GithubDependabot{}
	err = yaml.Unmarshal(remoteFileData, &remoteTemplate)
	if err != nil {
		log.WithError(err).Error("could unmarshal template")
		return nil, err
	}

	err = dependabot.MergeDependabot(&remoteTemplate, localTemplate)
	if err != nil {
		log.WithError(err).Error("could merge dependabot template")
		return nil, err
	}

	output, err := yaml.Marshal(remoteTemplate)
	if err != nil {
		log.WithError(err).Error("could not marshal template")
		return nil, err
	}

	file.RepositoryUpdateOptions = &config.RepositoryFileUpdateOptions{}
	file.RepositoryUpdateOptions.Filename = content.GetName()
	file.RepositoryUpdateOptions.DisplayName = file.RepositoryUpdateOptions.Filename
	file.Dependabot = &remoteTemplate
	file.RepositoryUpdateOptions.FileContent = &output
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
