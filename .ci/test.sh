#!/usr/bin/env bash
set -euxo pipefail

# Run the tests
set +e
export OUT_FILE="build/test-report.out"
mkdir -p build
go test -race -coverprofile=coverage.txt -covermode=atomic ./... 2>&1 | tee ${OUT_FILE}
status=$?

go get -v -u github.com/jstemmer/go-junit-report
go-junit-report > "build/junit-${GO_VERSION}.xml" < ${OUT_FILE}

exit ${status}