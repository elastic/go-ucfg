#!/usr/bin/env bash
set -exu pipefail

cd $(mktemp -d) && go mod init tempmod && go get -u $@
