#!/bin/bash

set -eu

ROOT="$(cd $(dirname $0)/.. && pwd)"
TARGET="${ROOT}/target"
ARTIFACT="${TARGET}/artifact"
source $(dirname $0)/env.sh

echo
echo "# monitoring build"
echo "#"

docker run -i --rm \
  -e AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID} \
  -e AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY} \
  -e AWS_REGION=${AWS_REGION:=us-west-2} \
  realogy/pipeline:latest build monitor \
    --branch "${GIT_BRANCH}" \
    --commit "${GIT_COMMIT}" \
    --repository ${GIT_DOMAIN}/${GIT_OWNER}/${GIT_SLUG}

echo "#"
echo "# ok"


