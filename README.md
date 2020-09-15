<p align="center">
  <img alt="ghconfig Logo" src="https://raw.githubusercontent.com/StarpTech/ghconfig/master/docs/logo.png" />
  <h3 align="center">GitHub Config</h3>
  <p align="center">Manage Github CI workflow files as a fleet.</p>
</p>

---

`ghconfig` is a CLI library to manage Github CI workflow files as a fleet.

Managing Github CI Workflow and Dependabot files can be in organizations very exhausting because there is no functionality to apply changes in a batch. Ghconfig helps you to automate such tasks. You have two options:

- Apply [RFC7396 JSON merge patch](https://tools.ietf.org/html/rfc7396) based on your template and existing remote files.
- Apply a [RFC6902 JSON patch](http://tools.ietf.org/html/rfc6902) on an existing workflow file.

By default a Pull-Request is created for all changes on a repository.
Ghconfig looks for a folder `.ghconfig` in the root of your repository. 

**Example:** We will create an workflow `ci.yaml` and apply one patch to an existing workflow `release.yml`.

```
.
├── .github
│   └── workflows
│       ├── ci.yaml
│       └── release.json
├── .ghconfig
│   ├── workflows
│   │   ├── patches
│   │   │   └── release.yaml
│   │   ├── ci.yaml
│   └── dependabot.md
```

> _All other Community Health files like (SUPPORT.md, CONTRIBUTING.md, ISSUE templates) can be managed by a repository called [`.github`](https://docs.github.com/en/github/building-a-strong-community/creating-a-default-community-health-file#about-default-community-health-files) in the user or organization._

This directory follows the same structure as your `.github` folder. All files are handled as a [Go template](https://golang.org/pkg/text/template/) and in them you have access to the full [Repository](https://pkg.go.dev/github.com/google/go-github/v32/github?tab=doc#Repository) object of the [go-github](https://pkg.go.dev/github.com/google/go-github) library and additionally to all utility functions of [sprig](http://masterminds.github.io/sprig/).

## Example:

Sync your workflow and dependabot files:
```
$ ghconfig sync

? Please select all repositories:  [Use arrows to move, space to select, <right> to all, <left> to none, type to filter]
  [ ]  StarpTech/profiling-nodejs
  [ ]  StarpTech/go-web
  [ ]  StarpTech/next-localization
> [X]  StarpTech/shikaka

Repository         Files                 Url
----------         -----                 ---
StarpTech/shikaka  ci.yaml               https://github.com/StarpTech/shikaka/pull/X
                   dependabot.yml

sync took 1211.0006ms
```

Apply a single patch on the workflow `nodejs.yml`:
```
$ ghconfig patch

? Please select all repositories:  [Use arrows to move, space to select, <right> to all, <left> to none, type to filter]
  [ ]  StarpTech/profiling-nodejs
  [ ]  StarpTech/go-web
  [ ]  StarpTech/next-localization
> [X]  StarpTech/shikaka

? Please select all repositories: StarpTech/shikaka

Repository         Files                 Url        
----------         -----                 ---        
StarpTech/shikaka  nodejs.yml (patched)  https://github.com/StarpTech/shikaka/pull/X
      
sync took 400.1179ms
```

## We want your feedback

We'd love to hear your feedback about ghconfig. If you spot bugs or have features that you'd really like to see in ghconfig, please check out the [contributing page](./.github/CONTRIBUTING.md).

## Usage

- `ghconfig sync`
- `ghconfig patch`
- `ghconfig sync --no-create-pr`
- `ghconfig sync --query=MyOrganisation`
- `ghconfig sync --base-branch=master`
- `ghconfig sync --root-dir=different-ghconfig-root`
- `ghconfig sync --dry-run`

## Installation

Ensure that your personal access token is exported with `GITHUB_TOKEN`.
You can [download](https://github.com/starptech/ghconfig/releases) `ghconfig` from Github.

## Rate Limiting

GitHub imposes a rate limit on all API clients. Unauthenticated clients are
limited to 60 requests per hour, while authenticated clients can make up to
5,000 requests per hour. The Search API has a custom rate limit. Unauthenticated
clients are limited to 10 requests per minute, while authenticated clients
can make up to 30 requests per minute.

## Roadmap

This library is being initially developed for an internal application, so features will likely be implemented in the order that they are needed by that application. Feel free to create a feature request.

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

- [Github CI](https://docs.github.com/en/actions/getting-started-with-github-actions/core-concepts-for-github-actions)
- [About default community health files](https://docs.github.com/en/github/building-a-strong-community/creating-a-default-community-health-file)
- [Dependabot](https://github.blog/2020-06-01-keep-all-your-packages-up-to-date-with-dependabot/)
