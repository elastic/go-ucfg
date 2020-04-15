#!/usr/bin/env bash
set -exui pipefail

go mod verify || exit 1
git diff --exit-code
