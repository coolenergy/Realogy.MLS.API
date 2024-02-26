#!/bin/bash

set -eu

ROOT="$(cd $(dirname $0)/.. && pwd)"
TARGET="${ROOT}/target"
ARTIFACT="${TARGET}/artifact"

(cd ${ROOT}; GOOS=linux CGO_ENABLED=0 go build -a -installsuffix cgo -o mls-display-rules .)
pipeline docker build -p ${ARTIFACT} ${ROOT}

