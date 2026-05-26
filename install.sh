#!/bin/bash
set -e

git clone --depth 1 https://github.com/FacileStudio/capsule-cli.git /tmp/capsule-cli
cd /tmp/capsule-cli
go install .
rm -rf /tmp/capsule-cli

echo "capsule installed to $(go env GOPATH)/bin/capsule"
