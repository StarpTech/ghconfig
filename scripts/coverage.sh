#!/bin/bash

$(dirname "$0")/test.sh
go tool cover -html=coverage.out