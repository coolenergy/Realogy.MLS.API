#!/bin/bash

VERSION="latest"

docker pull realogy/pipeline:${VERSION}
docker run --name tmp realogy/pipeline:${VERSION} /bin/true
docker cp tmp:/usr/local/bin/pipeline /usr/local/bin/pipeline
docker rm tmp

