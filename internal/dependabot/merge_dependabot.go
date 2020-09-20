package dependabot

import (
	"fmt"
	"ghconfig/internal/common"
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
	if src.Milestone == "" {
		src.Milestone = dst.Milestone
	}
	if src.PullRequestBranchName == "" {
		src.PullRequestBranchName = dst.PullRequestBranchName
	}
	if src.RebaseStrategy == "" {
		src.RebaseStrategy = dst.RebaseStrategy
	}
	if src.PackageEcosystem == "" {
		src.PackageEcosystem = dst.PackageEcosystem
	}
	if src.Directory == "" {
		src.Directory = dst.Directory
	}
	if src.Schedule.Interval == "" {
		src.Schedule.Interval = dst.Schedule.Interval
	}
	if src.OpenPullRequestsLimit == 0 {
		src.OpenPullRequestsLimit = dst.OpenPullRequestsLimit
	}

	if len(src.Ignore) == 0 {
		src.Ignore = dst.Ignore
	}
	for _, srcIgnore := range src.Ignore {
		for _, dstIgnore := range dst.Ignore {
			if srcIgnore.DependencyName == dstIgnore.DependencyName {
				srcIgnore.Versions = common.Unique(srcIgnore.Versions, dstIgnore.Versions)
				break
			}
		}
	}

	src.Assignees = common.Unique(src.Assignees, dst.Assignees)
	src.Labels = common.Unique(src.Labels, dst.Labels)
	src.Reviewers = common.Unique(src.Reviewers, dst.Reviewers)
	src.TargetBranch = common.Unique(src.TargetBranch, dst.TargetBranch)
	src.VersioningStrategy = common.Unique(src.VersioningStrategy, dst.VersioningStrategy)

	if len(src.Allow) == 0 {
		src.Allow = dst.Allow
	}
	if src.CommitMessage.Include == "" {
		src.CommitMessage.Include = dst.CommitMessage.Include
	}
	if src.CommitMessage.Prefix == "" {
		src.CommitMessage.Prefix = dst.CommitMessage.Prefix
	}
	if src.CommitMessage.PrefixDevelopment == "" {
		src.CommitMessage.PrefixDevelopment = dst.CommitMessage.PrefixDevelopment
	}

}
