#!/usr/bin/env bash
set -exu pipefail

checkformat() {
	$@
	git diff --exit-code
}

go vet ./...


echo "Verify go modules"
checkformat go mod verify
checkformat go mod tidy

echo "Check format"
checkformat go fmt ./...

echo "Check for license headers"
go install github.com/elastic/go-licenser@latest
checkformat go-licenser -license ASL2

echo "Check notice file"
go install go.elastic.co/go-licence-detector@v0.8.0
checkformat dev-tools/generate_notice
