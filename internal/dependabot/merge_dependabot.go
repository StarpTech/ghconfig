package dependabot

import (
	"fmt"
	"reflect"

	"github.com/imdario/mergo"
)

type dependabotTransformer struct{}

func (t dependabotTransformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	if typ == reflect.TypeOf([]*Updates{}) {
		return func(dst, src reflect.Value) error {
			if dst.CanSet() {
				dstUpdates, ok := dst.Interface().([]*Updates)
				if !ok {
					return fmt.Errorf("expect dst to be type of Jobs, actual: %s", reflect.TypeOf(dst).Name())
				}
				srcUpdates, ok := src.Interface().([]*Updates)
				if !ok {
					return fmt.Errorf("expect src to be type of Jobs, actual: %s", reflect.TypeOf(src).Name())
				}

				for _, sj := range srcUpdates {
					for _, dj := range dstUpdates {
						if (sj.Directory == dj.Directory) && (sj.PackageEcosystem == dj.PackageEcosystem) {
							mergeUpdates(sj, dj)
						}
					}
				}
				dst.Set(reflect.ValueOf(srcUpdates))
			}
			return nil
		}
	}
	return nil
}

func MergeDependabot(dst *GithubDependabot, src GithubDependabot) error {
	return mergo.MergeWithOverwrite(dst, src, mergo.WithTypeCheck, mergo.WithTransformers(dependabotTransformer{}))
}

func mergeUpdates(src, dst *Updates) {
	if src.PackageEcosystem == "" {
		src.PackageEcosystem = dst.PackageEcosystem
	}
	if src.Directory == "" {
		src.Directory = dst.Directory
	}
	if src.Schedule.Interval == "" {
		src.Schedule.Interval = dst.Schedule.Interval
	}
	if src.OpenPullRequestsLimit == "" {
		src.OpenPullRequestsLimit = dst.OpenPullRequestsLimit
	}
	if len(src.Ignore) == 0 {
		src.Ignore = dst.Ignore
	}
}
