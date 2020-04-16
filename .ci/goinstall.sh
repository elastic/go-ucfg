#!/usr/bin/env bash

pkg=$1

set -exu pipefail

cd $(mktemp -d) && go mod init tempmod && go get -u $pkg
