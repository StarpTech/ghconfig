<p align="center">
  <img alt="ghconfig Logo" src="https://raw.githubusercontent.com/StarpTech/ghconfig/master/docs/logo.png" />
  <h3 align="center">GitHub Config</h3>
  <p align="center">Manage (.github) repository configurations as a fleet.</p>
</p>

---

`ghconfig` is a CLI library to manage (.github) repository configurations as a fleet.

Github CI Workflow files can be in organizations very similiar. If you need to update a single Job you have to update every
single repository manually. Ghconfig helps you to automate such tasks. You can work in two modes.

- Create a new workflow file based on your templates.
- Apply a [RFC6902 JSON patche](http://tools.ietf.org/html/rfc6902) on an existing workflow file.

By default a Pull-Request is created for all changes on a repository.

Ghconfig looks for a folder `.ghconfig` in the root of your repository.

```
.
├── .github
│   └── workflows
│       ├── ci.yaml
│       └── release.json
├── .ghconfig
│   ├── workflows
│   │   ├── patches
│   │   │   └── nodejs.yaml
│   │   ├── ci.yaml
│   │   └── release.yaml
│   ├── CODE_OF_CONDUCT.md
│   ├── dependabot.yml
│   ├── FUNDING.yml
│   ├── SECURITY.md
│   └── SUPPORT.md
```

This directory must have the same structure as your `.github` folder. Any `yaml` file is handled as a [Go template](https://golang.org/pkg/text/template/).

## Example:

```
$ ghconfig workflow
? Please select all target repositories. StarpTech/shikaka

Repository         Files               Url
----------         -----               ---
StarpTech/shikaka  test.yaml           https://github.com/StarpTech/shikaka/pull/20
                   test2.yaml
                   nodejs.yml (patch)
```

## We want your feedback

We'd love to hear your feedback about ghconfig. If you spot bugs or have features that you'd really like to see in ghconfig, please check out the [contributing page](./.github/CONTRIBUTING.md).

## Usage

- `ghconfig workflow`
- `ghconfig workflow --no-create-pr`
- `ghconfig workflow --query=MyOrganisation`
- `ghconfig workflow --base-branch=master`
- `ghconfig workflow --root-dir=different-workflow-dir`
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

- [X] Workflows
- [X] JSON-Patch for Workflows
- [ ] Test suite
- [ ] Manage Github community health files

## Versioning

In general, ghconfig follows [semver](https://semver.org/) as closely as we
can for tagging releases of the package. For self-contained libraries, the
application of semantic versioning is relatively straightforward and generally
understood.

## License

This library is distributed under the MIT-style license found in the [LICENSE](./LICENSE)
file.

## Credits

The logo is provided by [icons8.de](https://icons8.de)

## References

- [About default community health files](https://docs.github.com/en/github/building-a-strong-community/creating-a-default-community-health-file)
- [Dependabot](https://github.blog/2020-06-01-keep-all-your-packages-up-to-date-with-dependabot/)