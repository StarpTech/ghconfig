package workflow

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
	Run Run `yaml:"run,omitempty"`
}
type MatrixInclude struct {
	Node         string `yaml:"node,omitempty"`
	Os           string `yaml:"os,omitempty"`
	Experimental string `yaml:"experimental,omitempty"`
}
type MatrixExclude struct {
	Node         string `yaml:"node,omitempty"`
	Os           string `yaml:"os,omitempty"`
	Experimental string `yaml:"experimental,omitempty"`
}
type Matrix struct {
	NodeVersion []string        `yaml:"node-version,omitempty"`
	Node        []string        `yaml:"node,omitempty"`
	Os          []string        `yaml:"os,omitempty"`
	Include     []MatrixInclude `yaml:"include,omitempty"`
	Exclude     []MatrixExclude `yaml:"exclude,omitempty"`
}
type Strategy struct {
	Matrix      Matrix `yaml:"matrix,omitempty"`
	MaxParallel int    `yaml:"max-parallel,omitempty"`
	FailFast    bool   `yaml:"fail-fast,omitempty"`
}
type JobEnv = map[string]string

type ContainerEnv = map[string]string
type ContainerVolumes = map[string]string

type ServiceEnv = map[string]string
type ServiceVolumes = map[string]string

type Outputs = map[string]string

type With = map[string]string

type Step struct {
	Uses            string `yaml:"uses,omitempty"`
	ID              string `yaml:"id,omitempty"`
	If              string `yaml:"if,omitempty"`
	Name            string `yaml:"name,omitempty"`
	With            With   `yaml:"with,omitempty"`
	Run             string `yaml:"run,omitempty"`
	ContinueOnError bool   `yaml:"continue-on-error,omitempty"`
	TimeoutMinutes  int    `yaml:"timeout-minutes,omitempty"`
}
type Container struct {
	Image   string         `yaml:"image,omitempty"`
	Env     ContainerEnv   `yaml:"env,omitempty"`
	Ports   []string       `yaml:"ports,omitempty"`
	Volumes ServiceVolumes `yaml:"volumes,omitempty"`
	Options string         `yaml:"options,omitempty"`
}

type ContainerMarshaler struct {
	Image string `yaml:"image,omitempty"`
}

func (c *Container) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var container ContainerMarshaler
	err := unmarshal(&container)
	if err != nil {
		var imageName string
		err := unmarshal(&imageName)
		if err != nil {
			return err
		}
		*c = Container{Image: imageName}
	} else {
		*c = Container{
			Image: container.Image,
		}
	}
	return nil
}

type Service struct {
	Image   string           `yaml:"image,omitempty"`
	Env     ServiceEnv       `yaml:"env,omitempty"`
	Ports   []string         `yaml:"ports,omitempty"`
	Volumes ContainerVolumes `yaml:"volumes,omitempty"`
	Options string           `yaml:"options,omitempty"`
}
type Job struct {
	RunsOn          string             `yaml:"runs-on,omitempty"`
	Strategy        Strategy           `yaml:"strategy,omitempty"`
	Name            string             `yaml:"name,omitempty"`
	Env             JobEnv             `yaml:"env,omitempty"`
	ContinueOnError bool               `yaml:"continue-on-error,omitempty"`
	If              string             `yaml:"if,omitempty"`
	Defaults        Defaults           `yaml:"defaults,omitempty"`
	Outputs         Outputs            `yaml:"outputs,omitempty"`
	Steps           []Step             `yaml:"steps,omitempty"`
	Needs           StringArray        `yaml:"needs,omitempty"`     // string or array
	Container       Container          `yaml:"container,omitempty"` // can be string if only an image name is passed
	Services        map[string]Service `yaml:"services,omitempty"`
}

type StringArray []string

func (a *StringArray) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var multi []string
	err := unmarshal(&multi)
	if err != nil {
		var single string
		err := unmarshal(&single)
		if err != nil {
			return err
		}
		*a = []string{single}
	} else {
		*a = multi
	}
	return nil
}

type Jobs = map[string]Job
