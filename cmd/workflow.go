package cmd

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions

type Workflow struct {
	Name     string   `yaml:"name"`
	On       On       `yaml:"on,omitempty"`
	Env      Env      `yaml:"env,omitempty"`
	Defaults Defaults `yaml:"defaults,omitempty"`
	Jobs     Jobs     `yaml:"jobs,omitempty"`
}
type Schedule struct {
	Cron string `yaml:"cron,omitempty"`
}
type Push struct {
	PathsIgnore []string `yaml:"paths-ignore,omitempty"`
	TagsIgnore  []string `yaml:"tags-ignore,omitempty"`
	Branches    []string `yaml:"branches,omitempty"`
	Tags        []string `yaml:"tags,omitempty"`
}
type PullRequest struct {
	PathsIgnore []string `yaml:"paths-ignore,omitempty"`
	TagsIgnore  []string `yaml:"tags-ignore,omitempty"`
	Branches    []string `yaml:"branches,omitempty"`
	Tags        []string `yaml:"tags,omitempty"`
}
type Release struct {
	Types []string `yaml:"types"`
}
type On struct {
	Schedule    []Schedule  `yaml:"schedule,omitempty"`
	Push        Push        `yaml:"push,omitempty"`
	Release     Release     `yaml:"release,omitempty"`
	PageBuild   string      `yaml:"page_build,omitempty"`
	PullRequest PullRequest `yaml:"pull_request,omitempty"`
}
type Env = map[string]string

type Run struct {
	Shell            string `yaml:"shell,omitempty"`
	WorkingDirectory string `yaml:"working-directory,omitempty"`
}
type Defaults struct {
	Run Run `yaml:"run"`
}
type Matrix struct {
	NodeVersion []string `yaml:"node-version,omitempty"`
	Os          []string `yaml:"os,omitempty"`
}
type Strategy struct {
	Matrix Matrix `yaml:"matrix,omitempty"`
}
type JobEnv = map[string]string

type Outputs = map[string]string

type With = map[string]string

type Step struct {
	Uses string `yaml:"uses,omitempty"`
	ID   string `yaml:"id,omitempty"`
	If   string `yaml:"if,omitempty"`
	Name string `yaml:"name,omitempty"`
	With With   `yaml:"with,omitempty"`
	Run  string `yaml:"run,omitempty"`
}
type Job struct {
	RunsOn   string   `yaml:"runs-on,omitempty"`
	Strategy Strategy `yaml:"strategy,omitempty"`
	Name     string   `yaml:"name,omitempty"`
	Env      JobEnv   `yaml:"env,omitempty"`
	If       string   `yaml:"if,omitempty"`
	Defaults Defaults `yaml:"defaults,omitempty"`
	Outputs  Outputs  `yaml:"outputs,omitempty"`
	Steps    []Step   `yaml:"steps,omitempty"`
}

type Jobs = map[string]Job
