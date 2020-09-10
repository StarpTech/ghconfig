release:
  git tag $version
  git push origin $version
  goreleaser