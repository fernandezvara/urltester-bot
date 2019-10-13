#!/bin/bash

VERSION="0.2.1"
BUILD_DATE=$(date +%Y-%m-%d_%H:%M:%S) 
COMMIT=$(git rev-parse HEAD)
gox -osarch "linux/arm" -ldflags "-X main.Version=$VERSION -X main.Commit=$COMMIT -X main.BuildDate=$BUILD_DATE" -rebuild