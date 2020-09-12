package cmd

import (
	"encoding/json"
	"fmt"
	"ghconfig/internal"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/briandowns/spinner"
	"github.com/cheynewallace/tabby"
	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/google/go-github/v32/github"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
)

var (
	defaultWorkflowDir  = "workflows"
	branchNamePattern   = "ghconfig/workflows/%s"
	ghConfigWorkflowDir = ".ghconfig/%s"
	githubConfigDir     = ".github"
)

func NewSyncCmd(opts *internal.Config) error {
	pDir, err := os.Getwd()
	if err != nil {
		return err
	}

	workflowDirAbs := path.Join(pDir, fmt.Sprintf(ghConfigWorkflowDir, opts.WorkflowRoot))
	workflowFiles, err := ioutil.ReadDir(workflowDirAbs)
	if err != nil {
		return err
	}
	templates := []internal.WorkflowTemplate{}
	patches := []internal.WorkflowPatch{}

	for _, workflowFile := range workflowFiles {
		if workflowFile.IsDir() {
			continue
		}

		// check for patch
		if strings.HasSuffix(workflowFile.Name(), ".patch.json") {
			// without extension
			workflowName := strings.TrimSuffix(workflowFile.Name(), ".patch.json")
			filePath := path.Join(workflowDirAbs, workflowFile.Name())
			bytes, err := ioutil.ReadFile(filePath)
			if err != nil {
				kingpin.Errorf("could not read workflow patch file %v.", filePath)
				continue
			}
			if len(bytes) == 0 {
				continue
			}
			patch, err := jsonpatch.DecodePatch(bytes)
			if err != nil {
				kingpin.Errorf("invalid workflow patch file %v.", filePath)
				continue
			}
			patches = append(patches, internal.WorkflowPatch{
				FileName: workflowName,
				Patch:    patch,
			})
		} else {
			workflowName := workflowFile.Name()
			filePath := path.Join(workflowDirAbs, workflowName)
			bytes, err := ioutil.ReadFile(filePath)
			if err != nil {
				kingpin.Errorf("could not read workflow file %v.", filePath)
				continue
			}
			if len(bytes) == 0 {
				continue
			}
			t := internal.Workflow{}
			err = yaml.Unmarshal(bytes, &t)
			if err != nil {
				kingpin.Errorf("workflow file %v can't be parsed as workflow.", filePath)
				continue
			}
			templates = append(templates, internal.WorkflowTemplate{
				Workflow: &t,
				// restore to .github/workflows structure to update the correct file in the repo
				FilePath: path.Join(githubConfigDir, defaultWorkflowDir, workflowName),
				FileName: workflowName,
			})
		}
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

	intents := []*internal.UpdateWorkflowIntent{}

	s = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Update workflow files and create pull requests..."

	s.Start()

	t := tabby.New()
	t.AddHeader("Repository", "Files", "Url")

	for _, repoFullName := range targetRepos {
		repo := getRepoByName(repos, repoFullName)
		intent, err := collectWorkflowFiles(opts, repo, templates, patches)
		if err != nil {
			kingpin.Errorf("could not update workflow files, Repo: %v, error: %v", repoFullName, err)
			continue
		}

		if len(intent.WorkflowDrafts) > 0 {
			pullRequestURL := ""
			if !opts.DryRun {
				if opts.CreatePR {
					pullRequestURL, err = internal.CreatePR(opts, intent)
					if err != nil {
						kingpin.Errorf("could not create PR with changes: %v", err)
						continue
					}
				} else {
					err := internal.UpdateRepositoryFiles(opts, intent)
					if err != nil {
						kingpin.Errorf("could not update files only on remote: %v", err)
						continue
					}
				}
			}

			for i, draft := range intent.WorkflowDrafts {
				if i == 0 {
					url := ""
					if pullRequestURL != "" {
						url = pullRequestURL
					} else {
						url = draft.Url
					}
					t.AddLine(repo.GetFullName(), draft.DisplayName, url)
				} else {
					url := ""
					if pullRequestURL == "" {
						url = draft.Url
					}
					t.AddLine("", draft.DisplayName, url)
				}
			}

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
				_, err = file.Write([]byte(fmt.Sprintf("\n# Repository: %v, Workflow: %v\n%v\n---", wr.RepositoryName, draft.DisplayName, string(y))))
				if err != nil {
					kingpin.Errorf("could not write to ghconfig-debug.yml: %v", err)
				}
			}
		}
	}

	return nil
}

func collectWorkflowFiles(opts *internal.Config, repo *github.Repository, templates []internal.WorkflowTemplate, patches []internal.WorkflowPatch) (*internal.UpdateWorkflowIntent, error) {
	branchName := opts.BaseBranch

	if opts.CreatePR {
		branchName = fmt.Sprintf(branchNamePattern, opts.Sid.MustGenerate())
	}

	updateOptions := internal.UpdateWorkflowIntentOptions{
		Owner:       *repo.GetOwner().Login,
		Repo:        repo.GetName(),
		BaseRef:     opts.BaseBranch,
		Branch:      branchName,
		PRBranchRef: "refs/heads/" + branchName,
	}

	intent := internal.UpdateWorkflowIntent{
		WorkflowDrafts: []*internal.WorkflowUpdateDraft{},
		Options:        updateOptions,
		RepositoryName: repo.GetFullName(),
	}

	_, dirContent, resp, err := opts.GithubClient.Repositories.GetContents(
		opts.Context,
		updateOptions.Owner,
		updateOptions.Repo,
		".github/workflows",
		&github.RepositoryContentGetOptions{
			Ref: updateOptions.BaseRef,
		},
	)

	if err != nil {
		// when repo has no .github/workflow directory
		if resp.StatusCode == 404 {
			return &intent, nil
		}
		kingpin.Errorf("could not list workflow directory, %v", err)
		return nil, err
	}

	templateVars := map[string]interface{}{"Repo": repo}

	// find matched workflows files in the repository to get git SHA
	for _, workflowTemplate := range templates {
		var draft *internal.WorkflowUpdateDraft

		bytesCache, err := internal.ExecuteTemplate(repo.GetFullName(), workflowTemplate.Workflow, templateVars)
		if err != nil {
			kingpin.Errorf("could not template, %v", err)
			continue
		}

		proceedTemplate := internal.Workflow{}
		fileContent := bytesCache.Bytes()
		err = yaml.Unmarshal(bytesCache.Bytes(), &proceedTemplate)
		if err != nil {
			kingpin.Errorf("could not unmarshal template, %v", err)
			continue
		}

		for _, content := range dirContent {
			// match based on filename without extension, workflow name is not unique
			// since we work only with yaml files we can omit the extension
			var remoteFileName, localName string = strings.TrimSuffix(
				content.GetName(),
				filepath.Ext(content.GetName()),
			), strings.TrimSuffix(
				workflowTemplate.FileName,
				filepath.Ext(workflowTemplate.FileName),
			)

			if remoteFileName == localName {
				draft = &internal.WorkflowUpdateDraft{}
				draft.Filename = content.GetName()
				draft.DisplayName = draft.Filename
				draft.Workflow = &proceedTemplate
				draft.FileContent = &fileContent
				draft.FilePath = content.GetPath()
				draft.SHA = content.GetSHA()
			}
		}
		if draft == nil {
			draft = &internal.WorkflowUpdateDraft{}
			draft.Filename = workflowTemplate.FileName
			draft.DisplayName = workflowTemplate.FileName
			draft.Workflow = &proceedTemplate
			draft.FileContent = &fileContent
			draft.FilePath = workflowTemplate.FilePath
		}

		intent.WorkflowDrafts = append(intent.WorkflowDrafts, draft)
	}

	for _, workflowPatch := range patches {
		var draft *internal.WorkflowUpdateDraft
		for _, content := range dirContent {
			// match based on filename without extension, workflow name is not unique
			// since we work only with yaml files we can omit the extension
			var remoteFileName, localName string = strings.TrimSuffix(
				content.GetName(),
				filepath.Ext(content.GetName()),
			), strings.TrimSuffix(
				workflowPatch.FileName,
				filepath.Ext(workflowPatch.FileName),
			)

			if remoteFileName == localName {
				r, err := opts.GithubClient.Repositories.DownloadContents(
					opts.Context,
					updateOptions.Owner,
					updateOptions.Repo,
					content.GetPath(),
					&github.RepositoryContentGetOptions{
						Ref: updateOptions.BaseRef,
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

				t := internal.Workflow{}
				err = yaml.Unmarshal(data, &t)
				if err != nil {
					kingpin.Errorf("could not unmarshal workflow, %v", err)
					continue
				}

				j, err := json.Marshal(t)
				if err != nil {
					kingpin.Errorf("could not convert yaml to json, %v", err)
					continue
				}

				data, err = workflowPatch.Patch.Apply(j)
				if err != nil {
					kingpin.Errorf("could not apply patch, %v", err)
					continue
				}

				t = internal.Workflow{}
				err = yaml.Unmarshal(data, &t)
				if err != nil {
					kingpin.Errorf("could not unmarshal patched workflow, %v", err)
					continue
				}
				j, err = yaml.Marshal(&t)
				if err != nil {
					kingpin.Errorf("could not marshal patched workflow, %v", err)
					continue
				}

				draft = &internal.WorkflowUpdateDraft{}
				draft.Filename = content.GetName()
				draft.DisplayName = content.GetName() + " (patched)"
				draft.Workflow = &t
				draft.FileContent = &j
				draft.FilePath = content.GetPath()
				draft.SHA = content.GetSHA()

				intent.WorkflowDrafts = append(intent.WorkflowDrafts, draft)
			}
		}
	}

	return &intent, nil
}

func getRepoByName(repos []*github.Repository, name string) *github.Repository {
	for _, repo := range repos {
		if name == *repo.FullName {
			return repo
		}
	}
	return nil
}
