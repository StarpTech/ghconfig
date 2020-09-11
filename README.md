# ghconfig

ghconfig is a CLI library to manage (.github) repository configurations as a fleet.

Github CI Workflow files can be in organizations very similiar. If you need to update a single Job you have to update every
single repository manually. Ghconfig helps you to automate such tasks.

## How does it works? ##

Ghconfig looks for a folder `.ghconfig` in the root of your repository. This directory must have the same structure as your `.github` folder. Any file in the in the folder is handled as a [Go template](https://golang.org/pkg/text/template/). Currently, only the command `workflow` is implemented and therefore only `.github/workflows` are handled. We generate new workflows files and create a PR in every selected repository. Every execution creates a new PR.

```
$ ghconfig workflow
? Please select all target repositories. StarpTech/shikaka

Repository         Changes              Url
----------         -------              ---
StarpTech/shikaka  ci.yaml,release.yml  https://github.com/StarpTech/shikaka/pull/X
```

## Getting started ##

Ensure that your personal access token is exported with `GITHUB_TOKEN`.
You can [download](https://github.com/starptech/ghconfig/releases) `ghconfig` from Github.


### Dry-run

You can validate your generated templates without executing remote commands. Use the flag `--dry-run`. The output is saved to `ghconfig-debug.yml`.

```
ghconfig workflow --dry-run
```

## Templating

In all workflow files you have access to the full [Repository](https://pkg.go.dev/github.com/google/go-github/v32/github?tab=doc#Repository) object of the [go-github](https://pkg.go.dev/github.com/google/go-github) library. We use [sprig](http://masterminds.github.io/sprig/) to provide common helper functions.

Example:

```go
env:
    A: $(( uuidv4 ))
    B: $(( .Repo.GetFullName ))
```

## Help ##

List all available commands:
```
ghconfig --help
```

## Rate Limiting ##

GitHub imposes a rate limit on all API clients. Unauthenticated clients are
limited to 60 requests per hour, while authenticated clients can make up to
5,000 requests per hour. The Search API has a custom rate limit. Unauthenticated
clients are limited to 10 requests per minute, while authenticated clients
can make up to 30 requests per minute.

## Roadmap ##

This library is being initially developed for an internal application, so features will likely be implemented in the order that they are needed by that application. Feel free to create a feature request.

## Versioning ##

In general, ghconfig follows [semver](https://semver.org/) as closely as we
can for tagging releases of the package. For self-contained libraries, the
application of semantic versioning is relatively straightforward and generally
understood.

## License ##

This library is distributed under the MIT-style license found in the [LICENSE](./LICENSE)
file.