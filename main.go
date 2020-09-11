package main

import (
	"context"
	"os"

	"ghconfig/cmd"

	"github.com/google/go-github/v32/github"
	"github.com/teris-io/shortid"
	"golang.org/x/oauth2"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	appDesc = `ghconfig is a CLI library to manage (.github) repository configurations as a fleet.

Github CI Workflow files are in organizations very similiar. If you need to update a single Job you have to update every
single repository manually. Ghconfig helps you to automate such tasks.
	`
	app             = kingpin.New("ghconfig", appDesc)
	dryRun          = app.Flag("dry-run", "Runs the command without side-effects.").Bool()
	baseBranch      = app.Flag("base-branch", "The base branch.").Default("master").Short('b').String()
	githubToken     = app.Flag("github-token", "Your personal github access token.").OverrideDefaultFromEnvar("GITHUB_TOKEN").Short('t').Required().String()
	worfklowCommand = app.Command("workflow", "Generates new workflows files based on the local templates and create a PR (draft) in the repository with the changes.")
	createPR        = worfklowCommand.Flag("create-pr", "Create a new branch and PR for all changes.").Default("true").Short('p').Bool()
	repositoryQuery = worfklowCommand.Flag("query", "Search by repository name.").Short('f').String()
)

func main() {
	app.Version("0.0.4")
	app.Parse(os.Args[1:])

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	sid, err := shortid.New(1, shortid.DefaultABC, 2342)
	if err != nil {
		kingpin.Fatalf("could not create id generator, %v", err)
	}

	cfg := &cmd.Config{
		GithubClient:    client,
		Context:         ctx,
		DryRun:          *dryRun,
		BaseBranch:      *baseBranch,
		Sid:             sid,
		CreatePR:        *createPR,
		RepositoryQuery: *repositoryQuery,
	}

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case worfklowCommand.FullCommand():
		if err := cmd.NewSyncCmd(cfg); err != nil {
			kingpin.Fatalf("command error: %v", err)
		}
	}

}
