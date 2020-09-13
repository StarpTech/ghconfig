package cmd

import (
	"encoding/json"
	"fmt"
	"ghconfig/internal"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/briandowns/spinner"
	"github.com/cheynewallace/tabby"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/google/go-github/v32/github"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
)

func NewSyncCmd(opts *internal.Config) error {
	pDir, err := os.Getwd()
	if err != nil {
		return err
	}
	workflowDirAbs := path.Join(pDir, fmt.Sprintf(internal.GhConfigWorkflowDir, opts.WorkflowRoot))
	templates, err := internal.FindWorkflows(workflowDirAbs)
	if err != nil {
		return err
	}

	workflowPatchesDirAbs := path.Join(pDir, fmt.Sprintf(internal.GhConfigWorkflowDir, path.Join(opts.WorkflowRoot, "patches")))
	patches, err := internal.FindPatches(workflowPatchesDirAbs)
	if err != nil {
		return err
	}

	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Collecting all available repositories..."
	s.Start()

	repos, err := internal.FetchAllRepos(opts)
	if err != nil {
		return err
	}
	s.Stop()

	reposNames := []string{}

	for _, repo := range repos {
		reposNames = append(reposNames, *repo.FullName)
	}

	targetRepoPrompt := &survey.MultiSelect{
		Message:  "Please select all target repositories.",
		PageSize: 20,
		Options:  reposNames,
	}

	targetRepos := []string{}
	survey.AskOne(targetRepoPrompt, &targetRepos)

	packages := []*internal.WorkflowUpdatePackage{}

	s = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Update workflow files and create pull requests..."

	s.Start()

	t := tabby.New()
	t.AddHeader("Repository", "Files", "Url")

	for _, repoFullName := range targetRepos {
		repo := getRepoByName(repos, repoFullName)
		branchName := opts.BaseBranch

		if opts.CreatePR {
			branchName = fmt.Sprintf(internal.BranchNamePattern, opts.Sid.MustGenerate())
		}

		updateOptions := internal.RepositoryUpdateOptions{
			Owner:       *repo.GetOwner().Login,
			Repo:        repo.GetName(),
			BaseRef:     opts.BaseBranch,
			Branch:      branchName,
			PRBranchRef: "refs/heads/" + branchName,
		}

		pkg := &internal.WorkflowUpdatePackage{
			RepositoryOptions: updateOptions,
			Repository:        repo,
		}

		files, err := collectWorkflowChanges(opts, pkg, templates, patches)
		if err != nil {
			kingpin.Errorf("could not update workflow files, Repo: %v, error: %v", repoFullName, err)
			continue
		}
		pkg.Files = files

		if len(pkg.Files) > 0 {
			pullRequestURL := ""
			if !opts.DryRun {
				if opts.CreatePR {
					pullRequestURL, err = internal.CreatePR(opts, pkg)
					if err != nil {
						kingpin.Errorf("could not create PR with changes: %v", err)
						continue
					}
				} else {
					err := internal.UpdateRepositoryFiles(opts, &pkg.RepositoryOptions, pkg.Files)
					if err != nil {
						kingpin.Errorf("could not update files only on remote: %v", err)
						continue
					}
				}
			}

			for i, files := range pkg.Files {
				if i == 0 {
					url := ""
					if pullRequestURL != "" {
						url = pullRequestURL
					} else {
						url = files.RepositoryUpdateOptions.URL
					}
					t.AddLine(repo.GetFullName(), files.RepositoryUpdateOptions.DisplayName, url)
				} else {
					url := ""
					if pullRequestURL == "" {
						url = files.RepositoryUpdateOptions.URL
					}
					t.AddLine("", files.RepositoryUpdateOptions.DisplayName, url)
				}
			}

		}

		packages = append(packages, pkg)
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

		for _, wr := range packages {
			for _, files := range wr.Files {
				y, err := yaml.Marshal(files.Workflow)
				if err != nil {
					continue
				}
				_, err = file.Write([]byte(fmt.Sprintf("\n# Repository: %v, Workflow: %v\n%v\n---", wr.Repository.GetFullName(), files.RepositoryUpdateOptions.DisplayName, string(y))))
				if err != nil {
					kingpin.Errorf("could not write to ghconfig-debug.yml: %v", err)
				}
			}
		}
	}

	return nil
}

func collectWorkflowChanges(opts *internal.Config, pkg *internal.WorkflowUpdatePackage, templates []*internal.WorkflowTemplate, patches []*internal.PatchData) ([]*internal.WorkflowUpdatePackageFile, error) {
	files := []*internal.WorkflowUpdatePackageFile{}
	_, dirContent, resp, err := opts.GithubClient.Repositories.GetContents(
		opts.Context,
		pkg.RepositoryOptions.Owner,
		pkg.RepositoryOptions.Repo,
		".github/workflows",
		&github.RepositoryContentGetOptions{
			Ref: pkg.RepositoryOptions.BaseRef,
		},
	)

	if err != nil {
		// when repo has no .github/workflow directory
		if resp.StatusCode == 404 {
			return files, nil
		}
		kingpin.Errorf("could not list workflow directory, %v", err)
		return nil, err
	}

	templateVars := map[string]interface{}{"Repo": pkg.Repository}

	// find matched workflows files in the repository to get git SHA
	for _, workflowTemplate := range templates {
		var file *internal.WorkflowUpdatePackageFile

		bytesCache, err := internal.ExecuteYAMLTemplate(workflowTemplate.FileName, workflowTemplate.Workflow, templateVars)
		if err != nil {
			kingpin.Errorf("could not template, %v", err)
			continue
		}

		proceedTemplate := internal.GithubWorkflow{}
		fileContent := bytesCache.Bytes()
		err = yaml.Unmarshal(bytesCache.Bytes(), &proceedTemplate)
		if err != nil {
			kingpin.Errorf("could not unmarshal template, %v", err)
			continue
		}

		for _, content := range dirContent {
			if content.GetName() == workflowTemplate.FileName {
				file = &internal.WorkflowUpdatePackageFile{}
				file.RepositoryUpdateOptions = &internal.RepositoryFileUpdateOptions{}
				file.RepositoryUpdateOptions.Filename = content.GetName()
				file.RepositoryUpdateOptions.DisplayName = file.RepositoryUpdateOptions.Filename
				file.Workflow = &proceedTemplate
				file.RepositoryUpdateOptions.FileContent = &fileContent
				file.RepositoryUpdateOptions.FilePath = content.GetPath()
				file.RepositoryUpdateOptions.SHA = content.GetSHA()
			}
		}
		if file == nil {
			file = &internal.WorkflowUpdatePackageFile{}
			file.RepositoryUpdateOptions = &internal.RepositoryFileUpdateOptions{}
			file.RepositoryUpdateOptions.Filename = workflowTemplate.FileName
			file.RepositoryUpdateOptions.DisplayName = workflowTemplate.FileName
			file.Workflow = &proceedTemplate
			file.RepositoryUpdateOptions.FileContent = &fileContent
			file.RepositoryUpdateOptions.FilePath = workflowTemplate.FilePath
		}

		files = append(files, file)
	}

	for _, patch := range patches {
		var file *internal.WorkflowUpdatePackageFile
		for _, content := range dirContent {
			if content.GetName() == patch.FileName {
				r, err := opts.GithubClient.Repositories.DownloadContents(
					opts.Context,
					pkg.RepositoryOptions.Owner,
					pkg.RepositoryOptions.Repo,
					content.GetPath(),
					&github.RepositoryContentGetOptions{
						Ref: pkg.RepositoryOptions.BaseRef,
					},
				)
				if err != nil {
					kingpin.Errorf("could not download workflow file for patch, %v", err)
					continue
				}
				data, err := ioutil.ReadAll(r)
				if err != nil {
					kingpin.Errorf("could not read from workflow file for patch, %v", err)
					continue
				}

				t := internal.GithubWorkflow{}
				err = yaml.Unmarshal(data, &t)
				if err != nil {
					kingpin.Errorf("could not unmarshal workflow, %v", err)
					continue
				}

				repositoryFileJSON, err := json.Marshal(t)
				if err != nil {
					kingpin.Errorf("could not convert yaml to json, %v", err)
					continue
				}

				jsonPatchData, err := internal.ExecuteYAMLTemplate(patch.FileName, patch, templateVars)
				if err != nil {
					kingpin.Errorf("could not template, %v", err)
					continue
				}

				newPatchData := internal.PatchData{}
				err = yaml.Unmarshal(jsonPatchData.Bytes(), &newPatchData)
				if err != nil {
					kingpin.Errorf("could not unmarshal template, %v", err)
					continue
				}

				templatedJSONPatchFile, err := json.Marshal(newPatchData.Patch)
				if err != nil {
					kingpin.Errorf("could not marshal templated patch file to json, %v", err)
					continue
				}

				jsonPatch, err := jsonpatch.DecodePatch(templatedJSONPatchFile)
				if err != nil {
					kingpin.Errorf("invalid patch file %v", err)
					continue
				}

				data, err = jsonPatch.Apply(repositoryFileJSON)
				if err != nil {
					kingpin.Errorf("could not apply patch, %v", err)
					continue
				}

				t = internal.GithubWorkflow{}
				err = yaml.Unmarshal(data, &t)
				if err != nil {
					kingpin.Errorf("could not unmarshal patched workflow, %v", err)
					continue
				}
				repositoryFileJSON, err = yaml.Marshal(&t)
				if err != nil {
					kingpin.Errorf("could not marshal patched workflow, %v", err)
					continue
				}

				file = &internal.WorkflowUpdatePackageFile{}
				file.RepositoryUpdateOptions = &internal.RepositoryFileUpdateOptions{}
				file.RepositoryUpdateOptions.Filename = content.GetName()
				file.RepositoryUpdateOptions.DisplayName = content.GetName() + " (patched)"
				file.Workflow = &t
				file.RepositoryUpdateOptions.FileContent = &repositoryFileJSON
				file.RepositoryUpdateOptions.FilePath = content.GetPath()
				file.RepositoryUpdateOptions.SHA = content.GetSHA()

				files = append(files, file)
			}
		}
	}

	return files, nil
}

func getRepoByName(repos []*github.Repository, name string) *github.Repository {
	for _, repo := range repos {
		if name == *repo.FullName {
			return repo
		}
	}
	return nil
}
