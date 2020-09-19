package github

import (
	"fmt"
	"reflect"

	"github.com/imdario/mergo"
)

type timeTransformer struct{}

func (t timeTransformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	if typ == reflect.TypeOf(Env{}) {
		return func(dst, src reflect.Value) error {
			if dst.CanSet() {
				dst.Set(src)
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

				if len(srcOn.Release.Types) == 0 {
					srcOn.Release.Types = dstOn.Release.Types
				}

				if len(srcOn.Push.Branches) == 0 {
					srcOn.Push.Branches = dstOn.Push.Branches
				}

				if len(srcOn.Push.TagsIgnore) == 0 {
					srcOn.Push.TagsIgnore = dstOn.Push.TagsIgnore
				}

				if len(srcOn.Push.Tags) == 0 {
					srcOn.Push.Tags = dstOn.Push.Tags
				}

				if len(srcOn.Push.PathsIgnore) == 0 {
					srcOn.Push.PathsIgnore = dstOn.Push.PathsIgnore
				}

				if len(srcOn.PullRequest.Branches) == 0 {
					srcOn.PullRequest.Branches = dstOn.PullRequest.Branches
				}

				if len(srcOn.PullRequest.PathsIgnore) == 0 {
					srcOn.PullRequest.PathsIgnore = dstOn.PullRequest.PathsIgnore
				}

				if len(srcOn.PullRequest.Tags) == 0 {
					srcOn.PullRequest.Tags = dstOn.PullRequest.Tags
				}

				if len(srcOn.PullRequest.TagsIgnore) == 0 {
					srcOn.PullRequest.TagsIgnore = dstOn.PullRequest.TagsIgnore
				}

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
	return mergo.MergeWithOverwrite(dst, src, mergo.WithTypeCheck, mergo.WithTransformers(timeTransformer{}))
}

func mergeJobs(src, dst *Job) {
	if len(src.Env) == 0 {
		src.Env = dst.Env
	}
	if len(src.Needs) == 0 {
		src.Needs = dst.Needs
	}
	if len(src.Outputs) == 0 {
		src.Outputs = dst.Outputs
	}
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
					for _, v := range srcValType {
						if sv, ok := v.([]string); ok {
							if len(sv) == 0 {
								if dstArray, ok := dstVal.([]string); ok {
									src.Strategy.Matrix[srcKey] = dstArray
								}
							}
						} else if _, ok := v.([]map[string]string); ok {
							if dstArrayMap, ok := dstVal.([]map[string]string); ok {
								if len(srcValType) == 0 {
									src.Strategy.Matrix[srcKey] = dstArrayMap
								}
							}
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
	if len(src.Container.Env) == 0 {
		src.Container.Env = dst.Container.Env
	}
	if len(src.Container.Ports) == 0 {
		src.Container.Ports = dst.Container.Ports
	}
	if len(src.Container.Volumes) == 0 {
		src.Container.Volumes = dst.Container.Volumes
	}
	if src.Container.Image == "" {
		src.Container.Image = dst.Container.Image
	}
}

func mergeJobServices(src, dst *Job) {
	for srcKey, srcSvc := range src.Services {
		for dstKey, dstSvc := range dst.Services {
			if srcKey == dstKey {
				if len(srcSvc.Ports) == 0 {
					srcSvc.Ports = dstSvc.Ports
				}
				if len(srcSvc.Env) == 0 {
					srcSvc.Env = dstSvc.Env
				}
				if len(srcSvc.Volumes) == 0 {
					srcSvc.Volumes = dstSvc.Volumes
				}
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
