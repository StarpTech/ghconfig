# ghconfig

ghconfig is a CLI library to manage (.github) repository configurations as a fleet.

Github CI Workflow files are in organizations very similiar. If you need to update a single Job you have to update every
single repository manually. Ghconfig helps you to automate such tasks. Clone your workflow directory (.github/workflows) to (.ghconfig/workflows). Use [Go templates](https://golang.org/pkg/text/template/) and sync it to all your repositories. 

## Getting started ##

Ensure that your personal access token is exported with `GITHUB_TOKEN`. 

Generate workflow files based on your local templates and create a PR's per repository.

```
$ go run main.go workflow
? Please select all target repositories. StarpTech/shikaka

Repository         Changes              Url
----------         -------              ---
StarpTech/shikaka  ci.yaml,release.yml  https://github.com/StarpTech/shikaka/pull/X
```

You can check your generated templates without creating a PR with `--dry-run`. The output is saved to `ghconfig-debug.yml`.

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

### Help ###

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