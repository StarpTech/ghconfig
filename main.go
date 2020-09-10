package main

import (
	"context"
	"os"

	"ghconfig/cmd"

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	appDesc = `ghconfig is a CLI library to manage (.github) repository configurations as a fleet.

E.g Github CI Workflow files are in many cases very similiar. Imagine a organization that has focused on delivering Node.js projects. If you need to update a single Job you have to update every
single repository manually. ghconfig helps you to automate such tasks. We are looking for a local folder ".ghconfig" which must have the
same structure as the ".github" folder. There you can put your templates.
	`
	app             = kingpin.New("ghconfig", appDesc)
	dryRun          = app.Flag("dry-run", "Runs the command without side-effects.").Default("true").Bool()
	githubToken     = app.Flag("githubToken", "Your personal github access token.").OverrideDefaultFromEnvar("GITHUB_TOKEN").Short('t').Required().String()
	commitMsg       = app.Flag("commit-msg", "Commit message.").Short('m').String()
	worfklowCommand = app.Command("workflow", "Generates new workflows files based on the local templates and create a PR (draft) in the repository with the changes.")
)

func main() {
	app.Version("0.0.1")
	app.Parse(os.Args[1:])

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	cfg := &cmd.Config{
		GithubClient: client,
		Context:      ctx,
		DryRun:       *dryRun,
		CommitMsg:    *commitMsg,
	}

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case worfklowCommand.FullCommand():
		if err := cmd.NewSyncCmd(cfg); err != nil {
			kingpin.Fatalf("command error: %v", err)
		}
	}

}
