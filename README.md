<p align="center">
  <img alt="ghconfig Logo" src="https://raw.githubusercontent.com/StarpTech/ghconfig/master/docs/logo.png" />
  <h3 align="center">GitHub Config</h3>
  <p align="center">Manage Github Workflows and Dependabot files as a fleet.</p>
</p>

---

`ghconfig` is a CLI library to manage Github Workflows and Dependabot files as a fleet.

Managing Workflow and Dependabot files can be in organizations very exhausting because there is no way to apply changes in a batch.

- You need to adjust your compatibility matrix from `13.x` to `14.x` on 50 repositories?
- You want to standardize your CI by managing all workflows in a single repository?

No problem, `ghconfig` helps you to automate such tasks. You have two options:

- Strategic two-way merge of your local and remote files.
- Apply a [RFC6902 JSON patch](http://tools.ietf.org/html/rfc6902) on a remote workflow file.

By default a Pull-Request is created for all changes on a repository.
Ghconfig looks for a folder `.ghconfig` in the root of your repository. 

**Example:** We will create a workflow `ci.yaml` and apply one patch to an existing workflow `release.yml` on all repositories in the organization `foo`.

```
$ ghconfig sync --query=org:foo
$ ghconfig patch --query=org:foo
```

```
.
├── .github
│   └── workflows
│       ├── ci.yaml
│       └── release.yaml
├── .ghconfig
│   ├── workflows
│   │   ├── patches
│   │   │   └── release.yaml
│   │   ├── ci.yaml
│   └── dependabot.md
```

> _All other Community Health files like (SUPPORT.md, CONTRIBUTING.md, ISSUE templates) can be managed by a repository called [`.github`](https://docs.github.com/en/github/building-a-strong-community/creating-a-default-community-health-file#about-default-community-health-files) in the user or organization._

This directory follows the same structure as your `.github` folder. All files are handled as a [Go template](https://golang.org/pkg/text/template/) and you have access to the full [Repository](https://pkg.go.dev/github.com/google/go-github/v32/github?tab=doc#Repository) object of the [go-github](https://pkg.go.dev/github.com/google/go-github) library and to all utility functions of [sprig](http://masterminds.github.io/sprig/):

**Example:** 

- `$(( .Repo.GetFullName ))`
- `$(( uuidv4 ))`

## Usage

- `ghconfig sync`
- `ghconfig patch`
- `ghconfig sync --no-create-pr`
- `ghconfig sync --query=MyOrganisation`
- `ghconfig sync --base-branch=master`
- `ghconfig sync --root-dir=different-ghconfig-root`
- `ghconfig sync --dry-run`

## Merge semantic

- **Adding:** Fields present in the local template that are missing from the remote template will be added to the remote template.

- **Updating:** Fields present in the local template will be merged recursively until a primitive field is updated, or a field is added. Primitive Values, Maps and string arrays present in the remote template are merged without duplicates.

- **Deleting:** Fields present in the remote template that have been removed from the local template will not be deleted from the remote template when the remote template field can be used as fallback. If you want to delete a field permanently you have to delete it first on the remote template. You could apply a JSON-patch to do it.

> In all scenarios we try to merge lossless. This is the case for entire Jobs, Steps (with the same `name` or `id` field) and Maps, String Arrays.

## Installation

Ensure that your personal access token is exported with `GITHUB_TOKEN`.
You can [download](https://github.com/starptech/ghconfig/releases) `ghconfig` from Github.

## We want your feedback

We'd love to hear your feedback about ghconfig. If you spot bugs or have features that you'd really like to see in ghconfig, please check out the [contributing page](./.github/CONTRIBUTING.md).

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
