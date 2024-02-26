#!/bin/bash

set -eu

ROOT="$(cd $(dirname $0)/.. && pwd)"
TARGET="${ROOT}/target"
ARTIFACT="${TARGET}/artifact"
source $(dirname $0)/env.sh

echo "gitprovider domain=${GIT_DOMAIN},  branch name = ${GIT_BRANCH}, commit id = ${GIT_COMMIT}, owner/team = ${GIT_OWNER}, repository name = ${GIT_SLUG}, version = ${VERSION}"

echo
echo "# creating artifact directory"
echo "#"

mkdir -p ${ARTIFACT}
cp    ${ROOT}/manifest.hcl   ${ARTIFACT}
cp -r ${ROOT}/cloudformation ${ARTIFACT}

cat <<EOF > ${ARTIFACT}/build.hcl
version = "${VERSION}"
EOF

echo "#"
echo "# ok"


