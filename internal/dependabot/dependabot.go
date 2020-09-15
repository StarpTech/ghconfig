package dependabot

type GithubDependabot struct {
	Version int       `yaml:"version"`
	Updates []Updates `yaml:"updates"`
}
type Schedule struct {
	Interval string `yaml:"interval"`
}
type Ignore struct {
	DependencyName string   `yaml:"dependency-name"`
	Versions       []string `yaml:"versions,omitempty"`
}
type Updates struct {
	PackageEcosystem      string   `yaml:"package-ecosystem"`
	Directory             string   `yaml:"directory"`
	Schedule              Schedule `yaml:"schedule"`
	OpenPullRequestsLimit int      `yaml:"open-pull-requests-limit,omitempty"`
	Ignore                []Ignore `yaml:"ignore,omitempty"`
}
