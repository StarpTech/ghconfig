package dependabot

type GithubDependabot struct {
	Version string     `yaml:"version,omitempty" json:"version,omitempty"`
	Updates []*Updates `yaml:"updates,omitempty" json:"updates,omitempty"`
}
type Schedule struct {
	Interval string `yaml:"interval,omitempty" json:"interval,omitempty"`
}
type Ignore struct {
	DependencyName string   `yaml:"dependency-name,omitempty" json:"dependency-name,omitempty"`
	Versions       []string `yaml:"versions,omitempty" json:"versions,omitempty"`
}

type PullRequestBranchName struct {
	Separator string `yaml:"separator,omitempty" json:"separator,omitempty"`
}

type Allow struct {
	DependencyName string `yaml:"dependency-name,omitempty" json:"dependency-name,omitempty"`
	DependencyType string `yaml:"dependency-type,omitempty" json:"dependency-type,omitempty"`
}

type CommitMessage struct {
	Prefix            string `yaml:"prefix,omitempty" json:"prefix,omitempty"`
	PrefixDevelopment string `yaml:"prefix-development,omitempty" json:"prefix-development,omitempty"`
	Include           string `yaml:"include,omitempty" json:"include,omitempty"`
}

type Updates struct {
	PackageEcosystem      string        `yaml:"package-ecosystem,omitempty" json:"package-ecosystem,omitempty"`
	Directory             string        `yaml:"directory,omitempty" json:"directory,omitempty"`
	Schedule              Schedule      `yaml:"schedule,omitempty" json:"schedule,omitempty"`
	OpenPullRequestsLimit int           `yaml:"open-pull-requests-limit,omitempty" json:"open-pull-requests-limit,omitempty"`
	Ignore                []*Ignore     `yaml:"ignore,omitempty" json:"ignore,omitempty"`
	Labels                []string      `yaml:"labels,omitempty" json:"labels,omitempty"`
	Milestone             string        `yaml:"milestone,omitempty" json:"milestone,omitempty"`
	PullRequestBranchName string        `yaml:"pull-request-branch-name,omitempty" json:"pull-request-branch-name,omitempty"`
	RebaseStrategy        string        `yaml:"rebase-strategy,omitempty" json:"rebase-strategy,omitempty"`
	Allow                 []*Allow      `yaml:"allow,omitempty" json:"allow,omitempty"`
	CommitMessage         CommitMessage `yaml:"commit-message,omitempty" json:"commit-message,omitempty"`
	Assignees             []string      `yaml:"assignees,omitempty" json:"assignees,omitempty"`
	Reviewers             []string      `yaml:"reviewers,omitempty" json:"reviewers,omitempty"`
	TargetBranch          []string      `yaml:"target-branch,omitempty" json:"target-branch,omitempty"`
	VersioningStrategy    []string      `yaml:"versioning-strategy,omitempty" json:"versioning-strategy,omitempty"`
}
