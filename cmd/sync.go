package cmd

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/Masterminds/sprig"
	"github.com/briandowns/spinner"
	"github.com/google/go-github/v32/github"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
)

type Config struct {
	GithubClient *github.Client
	Context      context.Context
	DryRun       bool
	CommitMsg    string
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
	workflowDir := path.Join(pDir, ".ghconfig/workflows")
	workflowFiles, err := ioutil.ReadDir(workflowDir)
	if err != nil {
		return err
	}
	templates := map[string]*Workflow{}

	for _, workflowFile := range workflowFiles {
		if workflowFile.IsDir() {
			continue
		}
		bytes, err := ioutil.ReadFile(path.Join(workflowDir, workflowFile.Name()))
		if err != nil {
			return err
		}
		t := Workflow{}
		err = yaml.Unmarshal(bytes, &t)
		if err != nil {
			return err
		}
		templates[t.Name] = &t
	}

	targetRepoPrompt := &survey.MultiSelect{
		Message:  "Please select all target repositories.",
		PageSize: 20,
		Options:  reposNames,
	}

	targetRepos := []string{}
	survey.AskOne(targetRepoPrompt, &targetRepos)

	for _, repoFullName := range targetRepos {
		repo := getRepoByName(repos, repoFullName)
		owner := *repo.GetOwner().Login
		repoName := repo.GetName()

		branchName := "feature/ghconfig_update_workflows"
		prBranchRef := "refs/heads/" + branchName
		masterRef := "master"

		// check if PR branch already exists
		existingRef, _, err := opts.GithubClient.Git.GetRef(opts.Context,
			owner,
			repoName,
			prBranchRef,
		)
		if err != nil {
			kingpin.Errorf("could not check for ref, %v", err)
			continue
		}

		sourceContentRef := masterRef
		if existingRef != nil {
			sourceContentRef = *existingRef.Ref
		}

		_, dirContent, _, err := opts.GithubClient.Repositories.GetContents(
			opts.Context,
			owner,
			repoName,
			".github/workflows",
			&github.RepositoryContentGetOptions{
				Ref: sourceContentRef,
			},
		)
		if err != nil {
			kingpin.Errorf("could not list workflow directory, %v", err)
			return err
		}

		for _, content := range dirContent {
			fullFileName := ".github/workflows/" + content.GetName()

			r, err := opts.GithubClient.Repositories.DownloadContents(
				opts.Context,
				owner,
				repoName,
				".github/workflows/"+content.GetName(),
				&github.RepositoryContentGetOptions{
					Ref: branchName,
				},
			)
			if err != nil {
				kingpin.Errorf("could not download repo file, %v", err)
				continue
			}

			defer r.Close()

			dst := Workflow{}

			repoFileContent, err := ioutil.ReadAll(r)
			if err != nil {
				kingpin.Errorf("could not read repo file, %v", err)
				continue
			}

			err = yaml.Unmarshal(repoFileContent, &dst)
			if err != nil {
				kingpin.Errorf("could not unmarshal repo template, %v", err)
				continue
			}

			if masterTemplate, ok := templates[dst.Name]; ok {
				vars := map[string]interface{}{"Repo": repo}
				y, err := yaml.Marshal(masterTemplate)
				if err != nil {
					kingpin.Errorf("could not marshal template, %v", err)
					continue
				}

				t := template.Must(template.New(owner+"/"+repoName).
					Delims("$((", "))").
					Funcs(sprig.FuncMap()).
					Parse(string(y)))

				bytesCache := new(bytes.Buffer)
				err = t.Execute(bytesCache, vars)
				if err != nil {
					kingpin.Errorf("could not template, %v", err)
					continue
				}

				proceedTemplate := Workflow{}
				err = yaml.Unmarshal(bytesCache.Bytes(), &proceedTemplate)
				if err != nil {
					kingpin.Errorf("could not unmarshal template, %v", err)
					continue
				}

				if !opts.DryRun {
					s = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
					s.Suffix = " Update remote workflow files..."
					s.FinalMSG = "Complete!\n"

					newFileContent := bytesCache.Bytes()

					s.Start()

					// get ref to branch from
					refs, _, err := opts.GithubClient.Git.ListMatchingRefs(
						opts.Context,
						owner,
						repoName,
						&github.ReferenceListOptions{
							Ref: masterRef,
						},
					)

					if err != nil {
						kingpin.Errorf("could not list Refs, %v", err)
						continue
					}

					if existingRef == nil {
						_, _, err = opts.GithubClient.Git.CreateRef(
							opts.Context,
							owner,
							repoName,
							&github.Reference{
								Ref:    &prBranchRef,
								Object: &github.GitObject{SHA: refs[0].Object.SHA},
							},
						)
						if err != nil {
							kingpin.Errorf("could not create Ref, %v", err)
							continue
						}
					}

					// commit message
					commitMsg := "Update " + content.GetName() + " by ghconfig"

					if opts.CommitMsg != "" {
						commitMsg = opts.CommitMsg
					}

					sha := content.GetSHA()

					rcr, _, err := opts.GithubClient.Repositories.UpdateFile(
						opts.Context,
						owner,
						repoName,
						fullFileName,
						&github.RepositoryContentFileOptions{
							Branch:  &branchName,
							Message: &commitMsg,
							Content: newFileContent,
							SHA:     &sha,
						},
					)
					if err != nil {
						kingpin.Errorf("could not update file %v", err)
						continue
					}

					draft := true
					baseBranch := "master"

					prs, _, err := opts.GithubClient.PullRequests.ListPullRequestsWithCommit(
						opts.Context,
						owner,
						repoName,
						*rcr.Commit.SHA,
						&github.PullRequestListOptions{
							State: "open",
						},
					)

					if len(prs) == 0 {
						_, _, err = opts.GithubClient.PullRequests.Create(
							opts.Context,
							owner,
							repoName,
							&github.NewPullRequest{
								Base:  &baseBranch,
								Title: &commitMsg,
								Draft: &draft,
								Head:  &branchName,
							},
						)
						if err != nil {
							kingpin.Errorf("could not create PR %v", err)
							continue
						}
					}

					s.Stop()
				} else {

					y, err := yaml.Marshal(proceedTemplate)
					if err != nil {
						kingpin.Errorf("could not marshal template for dry-run, %v", err)
						continue
					}
					fmt.Println("// Repository: ", repo.GetFullName(), "workflow: .github/workflows/"+content.GetName())
					fmt.Println(string(y))
				}
			}

		}
	}

	return nil
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
