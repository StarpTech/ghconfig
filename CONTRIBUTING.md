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