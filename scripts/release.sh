#!/bin/bash

git tag $1
git push origin $1
goreleaser --rm-dist