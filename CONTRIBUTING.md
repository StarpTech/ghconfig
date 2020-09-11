## Release

We build and release with [goreleaser](https://goreleaser.com/install/).

### Test it locally:

```
goreleaser --snapshot --skip-publish --rm-dist
```

### Release:

```
./release.sh v1.2.3
```

## Test

This repo is shipped with example templates located in `.ghconfig/workflows`. You can run the program and work with them.
In future we will implement integration tests.