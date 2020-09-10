## Release

We build and release with [goreleaser](https://goreleaser.com/install/).

### Test it locally:

```
goreleaser --snapshot --skip-publish --rm-dist
```

### Release:

```
make release version=v1.2.3
```