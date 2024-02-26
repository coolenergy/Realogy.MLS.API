#!/bin/bash

curl -X POST -d '{"commit_status":{"url":"https://bitbucket.org/'${BITBUCKET_REPO_OWNER}'/'${BITBUCKET_REPO_SLUG}'/addon/pipelines/home#!/results/'${BITBUCKET_BUILD_NUMBER}'","refname":"'${BITBUCKET_BRANCH}'","commit":{"hash":"'${BITBUCKET_COMMIT}'"},"state":"SUCCESSFUL"}}' https://bgmqyurr8l.execute-api.us-west-2.amazonaws.com/live/inbound/bitbucket
