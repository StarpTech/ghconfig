package github

import (
	"fmt"
	"ghconfig/internal/common"
	"reflect"

	"github.com/imdario/mergo"
)

type workflowTransformer struct{}

func (t workflowTransformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	if typ == reflect.TypeOf(Env{}) {
		return func(dst, src reflect.Value) error {
			if dst.CanSet() {
				dstEnv, ok := dst.Interface().(Env)
				if !ok {
					return fmt.Errorf("expect dst to be type of Env, actual: %s", reflect.TypeOf(dstEnv).Name())
				}
				srcEnv, ok := src.Interface().(Env)
				if !ok {
					return fmt.Errorf("expect src to be type of Env, actual: %s", reflect.TypeOf(srcEnv).Name())
				}
				merge := common.MergeStringMap(srcEnv, dstEnv)
				dst.Set(reflect.ValueOf(merge))
			}
			return nil
		}
	} else if typ == reflect.TypeOf(On{}) {
		return func(dst, src reflect.Value) error {
			if dst.CanSet() {
				dstOn, ok := dst.Interface().(On)
				if !ok {
					return fmt.Errorf("expect dst to be type of *On, actual: %s", reflect.TypeOf(src).Name())
				}
				srcOn, ok := src.Interface().(On)
				if !ok {
					return fmt.Errorf("expect src to be type of *On, actual: %s", reflect.TypeOf(src).Name())
				}

				if srcOn.PageBuild == "" {
					srcOn.PageBuild = dstOn.PageBuild
				}

				srcOn.Release.Types = common.Unique(srcOn.Release.Types, dstOn.Release.Types)
				srcOn.Push.Branches = common.Unique(srcOn.Push.Branches, dstOn.Push.Branches)
				srcOn.Push.TagsIgnore = common.Unique(srcOn.Push.TagsIgnore, dstOn.Push.TagsIgnore)
				srcOn.Push.Tags = common.Unique(srcOn.Push.Tags, dstOn.Push.Tags)
				srcOn.Push.PathsIgnore = common.Unique(srcOn.Push.PathsIgnore, dstOn.Push.PathsIgnore)
				srcOn.PullRequest.Branches = common.Unique(srcOn.PullRequest.Branches, dstOn.PullRequest.Branches)
				srcOn.PullRequest.PathsIgnore = common.Unique(srcOn.PullRequest.PathsIgnore, dstOn.PullRequest.PathsIgnore)
				srcOn.PullRequest.Tags = common.Unique(srcOn.PullRequest.Tags, dstOn.PullRequest.Tags)
				srcOn.PullRequest.TagsIgnore = common.Unique(srcOn.PullRequest.TagsIgnore, dstOn.PullRequest.TagsIgnore)

				if len(srcOn.Schedule) == 0 {
					srcOn.Schedule = dstOn.Schedule
				}

				dst.Set(reflect.ValueOf(srcOn))
			}
			return nil
		}
	} else if typ == reflect.TypeOf(Jobs{}) {
		return func(dst, src reflect.Value) error {
			if dst.CanSet() {
				dstJobs, ok := dst.Interface().(Jobs)
				if !ok {
					return fmt.Errorf("expect dst to be type of Jobs, actual: %s", reflect.TypeOf(dst).Name())
				}
				srcJobs, ok := src.Interface().(Jobs)
				if !ok {
					return fmt.Errorf("expect src to be type of Jobs, actual: %s", reflect.TypeOf(src).Name())
				}

				for sk, sj := range srcJobs {
					for dk, dj := range dstJobs {
						if sk == dk {
							mergeJobs(sj, dj)

						}
					}
					dst.SetMapIndex(reflect.ValueOf(sk), reflect.ValueOf(sj))
				}
			}
			return nil
		}
	}
	return nil
}

func MergeWorkflow(dst *GithubWorkflow, src GithubWorkflow) error {
	return mergo.MergeWithOverwrite(dst, src,
		mergo.WithTypeCheck,
		mergo.WithTransformers(workflowTransformer{}),
	)
}

func mergeJobs(src, dst *Job) {
	src.Env = common.MergeStringMap(dst.Env, src.Env)
	src.Needs = common.Unique(src.Needs, dst.Needs)
	src.Outputs = common.MergeStringMap(dst.Outputs, src.Outputs)

	if src.RunsOn == "" {
		src.RunsOn = dst.RunsOn
	}
	if src.RunsOn == "" {
		src.RunsOn = dst.RunsOn
	}
	if src.If == "" {
		src.If = dst.If
	}
	if !src.ContinueOnError {
		src.ContinueOnError = dst.ContinueOnError
	}
	if src.Defaults.Run.WorkingDirectory == "" {
		src.Defaults.Run.WorkingDirectory = dst.Defaults.Run.WorkingDirectory
	}
	if src.Defaults.Run.Shell == "" {
		src.Defaults.Run.Shell = dst.Defaults.Run.Shell
	}

	mergeJobStrategy(src, dst)

	mergeJobContainer(src, dst)

	if len(src.Services) == 0 {
		src.Services = dst.Services
	} else {
		mergeJobServices(src, dst)
	}

	if len(src.Steps) == 0 {
		src.Steps = dst.Steps
	} else {
		mergeJobSteps(src, dst)
	}
}

func isSameStep(a, b *Step) bool {
	if a.ID != "" && b.ID != "" {
		return a.ID == b.ID
	}
	if a.Name != "" && b.Name != "" {
		return a.Name == b.Name
	}
	if reflect.DeepEqual(a, b) {
		return true
	}
	return false
}

func mergeJobSteps(src, dst *Job) {
	for _, sStep := range src.Steps {
		for _, dStep := range dst.Steps {
			if isSameStep(sStep, dStep) {
				if len(sStep.With) == 0 {
					sStep.With = dStep.With
				}
				if sStep.Name == "" {
					sStep.Name = dStep.Name
				}
				if sStep.If == "" {
					sStep.If = dStep.If
				}
				if sStep.Run == "" {
					sStep.Run = dStep.Run
				}
				if sStep.ID == "" {
					sStep.ID = dStep.ID
				}
				if sStep.Uses == "" {
					sStep.Uses = dStep.Uses
				}
				if sStep.TimeoutMinutes == 0 {
					sStep.TimeoutMinutes = dStep.TimeoutMinutes
				}
				if !sStep.ContinueOnError {
					sStep.ContinueOnError = dStep.ContinueOnError
				}
				break
			}
		}
	}
}

func mergeJobStrategy(src, dst *Job) {
	if !src.Strategy.FailFast {
		src.Strategy.FailFast = dst.Strategy.FailFast
	}

	if src.Strategy.MaxParallel == 0 {
		src.Strategy.MaxParallel = dst.Strategy.MaxParallel
	}

	if len(src.Strategy.Matrix) == 0 {
		src.Strategy.Matrix = dst.Strategy.Matrix
		return
	}

	for srcKey, srcVal := range src.Strategy.Matrix {
		for dstKey, dstVal := range dst.Strategy.Matrix {
			if srcKey == dstKey {
				switch srcValType := srcVal.(type) {
				case []interface{}:
					stringArray := []interface{}{}
					newArray := []string{}
					if d, ok := dstVal.([]interface{}); ok {
						for _, a := range d {
							if av, ok := a.(string); ok {
								stringArray = append(stringArray, av)
							}
						}
						if len(stringArray) > 0 {
							stringArray = append(stringArray, srcValType...)
							for _, l := range stringArray {
								newArray = append(newArray, fmt.Sprintf("%v", l))
							}
							src.Strategy.Matrix[srcKey] = common.Unique(newArray, []string{})
						}
					}
				default:
					if !reflect.ValueOf(srcValType).IsZero() {
						src.Strategy.Matrix[srcKey] = dstVal
					}
				}
				break
			}
		}
	}
}

func mergeJobContainer(src, dst *Job) {
	src.Container.Env = common.MergeStringMap(dst.Container.Env, src.Container.Env)
	src.Container.Ports = common.Unique(src.Container.Ports, dst.Container.Ports)
	src.Container.Volumes = common.MergeStringMap(src.Container.Volumes, dst.Container.Volumes)
	if src.Container.Image == "" {
		src.Container.Image = dst.Container.Image
	}
}

func mergeJobServices(src, dst *Job) {
	for srcKey, srcSvc := range src.Services {
		for dstKey, dstSvc := range dst.Services {
			if srcKey == dstKey {
				srcSvc.Ports = common.Unique(srcSvc.Ports, dstSvc.Ports)
				srcSvc.Env = common.MergeStringMap(srcSvc.Env, dstSvc.Env)
				srcSvc.Volumes = common.MergeStringMap(srcSvc.Volumes, dstSvc.Volumes)
				if srcSvc.Image == "" {
					srcSvc.Image = dstSvc.Image
				}
				if srcSvc.Options == "" {
					srcSvc.Options = dstSvc.Options
				}
				break
			}
		}
	}
}
