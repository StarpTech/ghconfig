# Github config CLI

`ghconfig` is a CLI library to manage (.github) repository configurations as a fleet.

<p align="center">
  <a href="https://twitter.com/bcomnes/status/1303003812249174018">
  <img src="https://raw.githubusercontent.com/StarpTech/ghconfig/master/tweet.png">
  </a>
</p>

Github CI Workflow files can be in organizations very similiar. If you need to update a single Job you have to update every
single repository manually. Ghconfig helps you to automate such tasks.

Ghconfig looks for a folder `.ghconfig` in the root of your repository.

```
.
├── .github
│   └── workflows
│       ├── ci.yaml
│       └── release.yml
├── .ghconfig
│   └── workflows
│       ├── ci.yaml
│       └── release.yml
```

This directory must have the same structure as your `.github` folder. Any file in the in the folder is handled as a [Go template](https://golang.org/pkg/text/template/). Currently, only the command `workflow` is implemented and therefore only `.github/workflows` are respected. We generate new workflows files and create a PR in every selected repository. Every execution creates a new PR unless other specified with `--no-create-pr` the changes are commited directly on the base branch.

```
$ ghconfig workflow
? Please select all target repositories. StarpTech/shikaka

Repository         Changes              Url
----------         -------              ---
StarpTech/shikaka  ci.yaml,release.yml  https://github.com/StarpTech/shikaka/pull/X
```

## We want your feedback

We'd love to hear your feedback about ghconfig. If you spot bugs or have features that you'd really like to see in ghconfig, please check out the [contributing page](./CONTRIBUTING.md).

## Usage

- `ghconfig workflow`
- `ghconfig workflow --no-create-pr`
- `ghconfig workflow --query=MyOrganisation`
- `ghconfig workflow --base-branch=master`
- `ghconfig workflow --dry-run`

## Installation

Ensure that your personal access token is exported with `GITHUB_TOKEN`.
You can [download](https://github.com/starptech/ghconfig/releases) `ghconfig` from Github.

## Templating

In all workflow files you have access to the full [Repository](https://pkg.go.dev/github.com/google/go-github/v32/github?tab=doc#Repository) object of the [go-github](https://pkg.go.dev/github.com/google/go-github) library. We use [sprig](http://masterminds.github.io/sprig/) to provide common helper functions.

Example:

```go
env:
    A: $(( uuidv4 ))
    B: $(( .Repo.GetFullName ))
```

## Rate Limiting

GitHub imposes a rate limit on all API clients. Unauthenticated clients are
limited to 60 requests per hour, while authenticated clients can make up to
5,000 requests per hour. The Search API has a custom rate limit. Unauthenticated
clients are limited to 10 requests per minute, while authenticated clients
can make up to 30 requests per minute.

## Roadmap

This library is being initially developed for an internal application, so features will likely be implemented in the order that they are needed by that application. Feel free to create a feature request.

- [ ] Patch mode. Support `.patch.yml` extension to apply only specific changes.

## Versioning

In general, ghconfig follows [semver](https://semver.org/) as closely as we
can for tagging releases of the package. For self-contained libraries, the
application of semantic versioning is relatively straightforward and generally
understood.

## License

This library is distributed under the MIT-style license found in the [LICENSE](./LICENSE)
file.
