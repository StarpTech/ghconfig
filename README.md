# ghconfig

ghconfig is a CLI library to manage (.github) repository configurations as a fleet.

E.g Github CI Workflow files are in many cases very similiar. Imagine a organization that has focused on delivering Node.js projects. If you need to update a single Job you have to update every single repository manually. Ghconfig helps you to automate such tasks. We are looking for a local folder ".ghconfig" which must have the same structure as the ".github" folder. Templates are basic [Go templates](https://golang.org/pkg/text/template/). They have access to helper functions and the informations about the repository.

## Help

```
ghconfig --help
```

### Rate Limiting ###

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