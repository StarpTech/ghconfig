package github

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions

type (
	JobEnv           = map[string]string
	ContainerEnv     = map[string]string
	ContainerVolumes = map[string]string
	ServiceEnv       = map[string]string
	ServiceVolumes   = map[string]string
	Outputs          = map[string]string
	With             = map[string]string
	Env              = map[string]string
	MatrixValue      = interface{}
	Services         = map[string]*Service

	Jobs = map[string]*Job

	Matrix         = map[string]MatrixValue
	GithubWorkflow struct {
		Name     string   `yaml:"name,omitempty" json:"name,omitempty"`
		On       On       `yaml:"on,omitempty" json:"on,omitempty"`
		Env      Env      `yaml:"env,omitempty" json:"env,omitempty"`
		Defaults Defaults `yaml:"defaults,omitempty" json:"defaults,omitempty"`
		Jobs     Jobs     `yaml:"jobs,omitempty" json:"jobs,omitempty"`
	}
	Schedule struct {
		Cron string `yaml:"cron,omitempty" json:"cron,omitempty"`
	}
	Push struct {
		PathsIgnore []string `yaml:"paths-ignore,omitempty" json:"paths-ignore,omitempty"`
		TagsIgnore  []string `yaml:"tags-ignore,omitempty" json:"tags-ignore,omitempty"`
		Branches    []string `yaml:"branches,omitempty" json:"branches,omitempty"`
		Tags        []string `yaml:"tags,omitempty" json:"tags,omitempty"`
	}
	PullRequest struct {
		PathsIgnore []string `yaml:"paths-ignore,omitempty" json:"paths-ignore,omitempty"`
		TagsIgnore  []string `yaml:"tags-ignore,omitempty" json:"tags-ignore,omitempty"`
		Branches    []string `yaml:"branches,omitempty" json:"branches,omitempty"`
		Tags        []string `yaml:"tags,omitempty" json:"tags,omitempty"`
	}
	Release struct {
		Types []string `yaml:"types" json:"types"`
	}
	On struct {
		Schedule    []Schedule  `yaml:"schedule,omitempty" json:"schedule,omitempty"`
		Push        Push        `yaml:"push,omitempty" json:"push,omitempty"`
		Release     Release     `yaml:"release,omitempty" json:"release,omitempty"`
		PageBuild   string      `yaml:"page_build,omitempty" json:"page_build,omitempty"`
		PullRequest PullRequest `yaml:"pull_request,omitempty" json:"pull_request,omitempty"`
	}

	Run struct {
		Shell            string `yaml:"shell,omitempty" json:"shell,omitempty"`
		WorkingDirectory string `yaml:"working-directory,omitempty" json:"working-directory,omitempty"`
	}
	Defaults struct {
		Run Run `yaml:"run,omitempty" json:"run,omitempty"`
	}

	Strategy struct {
		Matrix      Matrix `yaml:"matrix,omitempty" json:"matrix,omitempty"`
		MaxParallel int    `yaml:"max-parallel,omitempty" json:"max-parallel,omitempty"`
		FailFast    bool   `yaml:"fail-fast,omitempty" json:"fail-fast,omitempty"`
	}

	Step struct {
		Uses            string `yaml:"uses,omitempty" json:"uses,omitempty"`
		ID              string `yaml:"id,omitempty" json:"id,omitempty"`
		If              string `yaml:"if,omitempty" json:"if,omitempty"`
		Name            string `yaml:"name,omitempty" json:"name,omitempty"`
		With            With   `yaml:"with,omitempty" json:"with,omitempty"`
		Run             string `yaml:"run,omitempty" json:"run,omitempty"`
		ContinueOnError bool   `yaml:"continue-on-error,omitempty" json:"continue-on-error,omitempty"`
		TimeoutMinutes  int    `yaml:"timeout-minutes,omitempty" json:"timeout-minutes,omitempty"`
	}
	Container struct {
		Image   string         `yaml:"image,omitempty" json:"image,omitempty"`
		Env     ContainerEnv   `yaml:"env,omitempty" json:"env,omitempty"`
		Ports   []string       `yaml:"ports,omitempty" json:"ports,omitempty"`
		Volumes ServiceVolumes `yaml:"volumes,omitempty" json:"volumes,omitempty"`
		Options string         `yaml:"options,omitempty" json:"options,omitempty"`
	}

	ContainerMarshaler struct {
		Image string `yaml:"image,omitempty" json:"image,omitempty"`
	}

	Service struct {
		Image   string           `yaml:"image,omitempty" json:"image,omitempty"`
		Env     ServiceEnv       `yaml:"env,omitempty" json:"env,omitempty"`
		Ports   []string         `yaml:"ports,omitempty" json:"ports,omitempty"`
		Volumes ContainerVolumes `yaml:"volumes,omitempty" json:"volumes,omitempty"`
		Options string           `yaml:"options,omitempty" json:"options,omitempty"`
	}

	Job struct {
		RunsOn          string      `yaml:"runs-on,omitempty" json:"runs-on,omitempty"`
		Strategy        Strategy    `yaml:"strategy,omitempty" json:"strategy,omitempty"`
		Name            string      `yaml:"name,omitempty" json:"name,omitempty"`
		Env             JobEnv      `yaml:"env,omitempty" json:"env,omitempty"`
		ContinueOnError bool        `yaml:"continue-on-error,omitempty" json:"continue-on-error,omitempty"`
		If              string      `yaml:"if,omitempty" json:"if,omitempty"`
		Defaults        Defaults    `yaml:"defaults,omitempty" json:"defaults,omitempty"`
		Outputs         Outputs     `yaml:"outputs,omitempty" json:"outputs,omitempty"`
		Steps           []*Step     `yaml:"steps,omitempty" json:"steps,omitempty"`
		Needs           StringArray `yaml:"needs,omitempty" json:"needs,omitempty"`         // string or array
		Container       Container   `yaml:"container,omitempty" json:"container,omitempty"` // can be string if only an image name is passed
		Services        Services    `yaml:"services,omitempty" json:"services,omitempty"`
	}

	StringArray []string
)

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
