#!/bin/bash

mkdir -p bin 
rm -rf bin/*

VERSION="0.3.2"
BUILD_DATE=$(date +%Y-%m-%d_%H:%M:%S) 
COMMIT=$(git rev-parse HEAD)
gox -ldflags "-X main.Version=$VERSION -X main.Commit=$COMMIT -X main.BuildDate=$BUILD_DATE" -rebuild -output "./bin/{{.Dir}}_{{.OS}}_{{.Arch}}"

ghr \
    -t ${GITHUB_TOKEN}       \
    -u fernandezvara         \
    -r urltester-bot         \
    -c ${COMMIT}             \
    -n ${VERSION}            \
    -b "Version: ${VERSION}" \
    ${VERSION} ./bin/

