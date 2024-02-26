#!/bin/bash

set -eu

ROOT="$(cd $(dirname $0)/.. && pwd)"
TARGET="${ROOT}/target"
ARTIFACT="${TARGET}/artifact"

pipeline docker login
pipeline docker push -p ${ARTIFACT}

