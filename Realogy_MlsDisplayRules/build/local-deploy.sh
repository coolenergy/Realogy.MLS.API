#!/usr/bin/env bash
set -eu
ROOT="$(cd $(dirname $0)/.. && pwd)"
SERVICE_NAME="eapdenonprd"

# run local=deploy.sh -p
if [[ $0 == "d"]]; then
    echo "Building a docker image"
    ${ROOT}/build/docker-build.sh
    ${ROOT}/build/docker-push.sh

fi
echo "Deploy service to ${SERVICE_NAME}"
pipeline artifact deploy --s ${SERVICE_NAME} --path  ${ROOT}/target/artifact

echo "Service log"
cwlogs fetch  /eapdenonprd/mls-display-rules -f

echo "Done"