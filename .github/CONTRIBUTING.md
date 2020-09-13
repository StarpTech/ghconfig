## Contributing

[legal]: https://help.github.com/articles/github-terms-of-service/#6-contributions-under-repository-license
[license]: ../LICENSE
[bug issues]: https://github.com/StarpTech/ghconfig/issues?q=is%3Aopen+is%3Aissue+label%3Abug
[feature request issues]: https://github.com/StarpTech/ghconfig/issues?q=is%3Aopen+is%3Aissue+label%3Aenhancement

Hi! Thanks for your interest in contributing to the ghconfig CLI!

We accept pull requests for bug fixes and features where we've discussed the approach in an issue and given the go-ahead for a community member to work on it. We'd also love to hear about ideas for new features as issues.

Please do:

* check existing issues to verify that the [bug][bug issues] or [feature request][feature request issues] has not already been submitted
* open an issue if things aren't working as expected
* open an issue to propose a significant change
* open a pull request to fix a bug
* open a pull request to fix documentation about a command
* open a pull request if a member of the ghconfig CLI team has given the ok after discussion in an issue

Please avoid:

* adding installation instructions specifically for your OS/package manager

## Building the project

Prerequisites:
- Go 1.15+

Build with: `go build`

Run the new binary as: `ghconfig`

Run tests with: `go test ./...`

## Submitting a pull request

1. Create a new branch: `git checkout -b my-branch-name`
1. Make your change, add tests, and ensure tests pass
1. Submit a pull request: `gh pr create --web`

Contributions to this project are [released][legal] to the public under the [project's open source license][license].

Please note that this project adheres to a [Contributor Code of Conduct][code-of-conduct]. By participating in this project you agree to abide by its terms.

We generate manual pages from source on every release. You do not need to submit pull requests for documentation specifically; manual pages for commands will automatically get updated after your pull requests gets accepted.

## Release the project

We build and release with [goreleaser](https://goreleaser.com/install/).

1. Bump version in `main.go`
1. Execute `./release.sh v1.2.3`

### Generate builds locally

```
goreleaser --snapshot --skip-publish --rm-dist
```

## Integration Test

This repo is shipped with example templates located in `.ghconfig/workflows`. You can run the program and work with them. In future we will implement automatic integration tests.

## Debugging

For easier debugging use the [spew](https://github.com/davecgh/go-spew) package and the [Golang extension](https://code.visualstudio.com/docs/languages/go) for VSCode to debug a test with a click.

## Resources

- [How to Contribute to Open Source](https://opensource.guide/how-to-contribute/)
- [Using Pull Requests](https://help.github.com/articles/about-pull-requests/)
- [GitHub Help](https://help.github.com)