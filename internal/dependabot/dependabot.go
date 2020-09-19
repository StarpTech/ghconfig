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
type Updates struct {
	PackageEcosystem      string   `yaml:"package-ecosystem,omitempty" json:"package-ecosystem,omitempty"`
	Directory             string   `yaml:"directory,omitempty" json:"directory,omitempty"`
	Schedule              Schedule `yaml:"schedule,omitempty" json:"schedule,omitempty"`
	OpenPullRequestsLimit string   `yaml:"open-pull-requests-limit,omitempty" json:"open-pull-requests-limit,omitempty"`
	Ignore                []Ignore `yaml:"ignore,omitempty" json:"ignore,omitempty"`
}
