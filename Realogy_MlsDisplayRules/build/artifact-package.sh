#!/bin/bash

set -eu

ROOT="$(cd $(dirname $0)/.. && pwd)"
TARGET="${ROOT}/target"
ARTIFACT="${TARGET}/artifact"
VERSION="${BITBUCKET_BUILD_NUMBER:=0}.$(echo ${BITBUCKET_COMMIT:=latest}| cut -c1-6)"

echo
echo "# creating artifact directory"
echo "#"

mkdir -p ${ARTIFACT}
cp    ${ROOT}/build/manifest.hcl   ${ARTIFACT}
cp -r ${ROOT}/build/cloudformation ${ARTIFACT}

cat <<EOF > ${ARTIFACT}/build.hcl
version = "${VERSION}"
EOF

echo "#"
echo "# ok"


