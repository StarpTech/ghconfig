package github

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/imdario/mergo"
)

type timeTransformer struct{}

func (t timeTransformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	if typ == reflect.TypeOf(On{}) {
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
				srcOn.Release.Types = uniqueStringArray(srcOn.Release.Types, dstOn.Release.Types)

				srcOn.Push.Branches = uniqueStringArray(srcOn.Push.Branches, dstOn.Push.Branches)
				srcOn.Push.PathsIgnore = uniqueStringArray(srcOn.Push.PathsIgnore, dstOn.Push.PathsIgnore)
				srcOn.Push.Tags = uniqueStringArray(srcOn.Push.Tags, dstOn.Push.Tags)
				srcOn.Push.TagsIgnore = uniqueStringArray(srcOn.Push.TagsIgnore, dstOn.Push.TagsIgnore)

				srcOn.PullRequest.Branches = uniqueStringArray(srcOn.PullRequest.Branches, dstOn.PullRequest.Branches)
				srcOn.PullRequest.PathsIgnore = uniqueStringArray(srcOn.PullRequest.PathsIgnore, dstOn.PullRequest.PathsIgnore)
				srcOn.PullRequest.Tags = uniqueStringArray(srcOn.PullRequest.Tags, dstOn.PullRequest.Tags)
				srcOn.PullRequest.TagsIgnore = uniqueStringArray(srcOn.PullRequest.TagsIgnore, dstOn.PullRequest.TagsIgnore)

				mergeMapOnSchedule(&dstOn, &srcOn)

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
					for _, dj := range dstJobs {
						if sj.Name == dj.Name {
							mergeMapJobs(sj, dj)
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

func mergeMapOnSchedule(dst, src *On) {
	schedules := []Schedule{}
	keys := make(map[string]struct{})

	for _, s := range dst.Schedule {
		if _, ok := keys[s.Cron]; !ok {
			schedules = append(schedules, s)
			keys[s.Cron] = struct{}{}
		}
	}

	for _, s := range src.Schedule {
		if _, ok := keys[s.Cron]; !ok {
			schedules = append(schedules, s)
			keys[s.Cron] = struct{}{}
		}
	}

	src.Schedule = schedules

}

func mergeMapJobs(src, dst *Job) {
	src.Needs = uniqueStringArray(src.Needs, dst.Needs)
	src.Env = mergeStringMap(src.Env, dst.Env)
	src.Needs = uniqueStringArray(src.Needs, dst.Needs)
	src.Outputs = mergeStringMap(src.Outputs, dst.Outputs)

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

	if len(src.Strategy.Matrix) == 0 {
		src.Strategy.Matrix = dst.Strategy.Matrix
	} else {
		mergeMapJobStrategy(src, dst)
	}

	mergeMapJobContainer(src, dst)

	if len(src.Services) == 0 {
		src.Services = dst.Services
	} else {
		mergeMapJobServices(src, dst)
	}

	if len(src.Steps) == 0 {
		src.Steps = dst.Steps
	} else {
		mergeMapJobSteps(src, dst)
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

func mergeMapJobSteps(src, dst *Job) {
	for _, sStep := range src.Steps {
		for _, dStep := range dst.Steps {
			if isSameStep(sStep, dStep) {
				sStep.With = mergeStringMap(sStep.With, dStep.With)
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
			}
		}
	}

	// add missing not duplicate steps from dst
	for _, djStep := range dst.Steps {
		dup := false
		for _, sjStep := range src.Steps {
			if isSameStep(djStep, sjStep) {
				dup = true
				break
			}
		}
		if !dup {
			src.Steps = append(src.Steps, djStep)
		}
	}
}

func mergeMapJobStrategy(src, dst *Job) {
	if !src.Strategy.FailFast {
		src.Strategy.FailFast = dst.Strategy.FailFast
	}

	if src.Strategy.MaxParallel == 0 {
		src.Strategy.MaxParallel = dst.Strategy.MaxParallel
	}

	for srcKey, srcVal := range src.Strategy.Matrix {
		for dstKey, dstVal := range dst.Strategy.Matrix {
			if srcKey == dstKey {
				switch srcValType := srcVal.(type) {
				case []string:
					if dstArray, ok := dstVal.([]string); ok {
						src.Strategy.Matrix[srcKey] = uniqueStringArray(srcValType, dstArray)
					}
				case []map[string]string:
					if dstArrayMap, ok := dstVal.([]map[string]string); ok {
						rr := []map[string]string{}
						// add missing not duplicate from src
						for _, srcMap := range srcValType {
							dup := false
							for _, dstMap := range dstArrayMap {
								if reflect.DeepEqual(srcMap, dstMap) {
									dup = true
									break
								}
							}
							if !dup {
								rr = append(rr, srcMap)
							}
						}
						// add missing not duplicate from dst
						for _, dstMap := range dstArrayMap {
							dup := false
							for _, rrMap := range rr {
								if reflect.DeepEqual(rrMap, dstMap) {
									dup = true
									break
								}
							}
							if !dup {
								rr = append(rr, dstMap)
							}
						}
						src.Strategy.Matrix[srcKey] = rr
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

func mergeMapJobContainer(src, dst *Job) {
	src.Container.Env = mergeStringMap(src.Container.Env, dst.Container.Env)
	src.Container.Ports = uniqueStringArray(src.Container.Ports, dst.Container.Ports)
	src.Container.Volumes = mergeStringMap(src.Container.Volumes, dst.Container.Volumes)
	if src.Container.Image == "" {
		src.Container.Image = dst.Container.Image
	}
}

func mergeMapJobServices(src, dst *Job) {
	for srcKey, srcSvc := range src.Services {
		for dstKey, dstSvc := range dst.Services {
			if srcKey == dstKey {
				srcSvc.Ports = uniqueStringArray(srcSvc.Ports, dstSvc.Ports)
				srcSvc.Env = mergeStringMap(srcSvc.Env, dstSvc.Env)
				srcSvc.Volumes = mergeStringMap(srcSvc.Volumes, dstSvc.Volumes)
				if srcSvc.Image == "" {
					srcSvc.Image = dstSvc.Image
				}
				if srcSvc.Options == "" {
					srcSvc.Options = dstSvc.Options
				}
			}
		}
	}

	// add missing not duplicate services from dst
	for k, dstSvc := range dst.Services {
		if _, ok := src.Services[k]; !ok {
			src.Services[k] = dstSvc
		}
	}
}

func uniqueStringArray(a, b []string) []string {
	stringSlice := append(a, b...)
	keys := make(map[string]struct{})
	list := []string{}

	for _, entry := range stringSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = struct{}{}
			list = append(list, entry)
		}
	}

	if len(list) == 0 {
		return nil
	}

	sort.Strings(list)

	return list
}

func mergeStringMap(a, b map[string]string) map[string]string {
	for k, v := range b {
		a[k] = v
	}
	return a
}
