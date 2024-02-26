#!/bin/bash

LOCALSTACK_SSM=`curl -sS http://127.0.0.1:4566/health | jq '.services.ssm?'`

if [[ "$LOCALSTACK_SSM" = "\"running\"" ]] ; then
  awslocal ssm put-parameter --name "/realogy/services/local/mls-display-rules/mls-mongodb-host" --value "mongo:27017" --type SecureString --region us-west-2
  awslocal ssm put-parameter --name "/realogy/services/local/mls-display-rules/mls-mongodb-user" --value "root" --type SecureString --region us-west-2
  awslocal ssm put-parameter --name "/realogy/services/local/mls-display-rules/mls-mongodb-pw" --value "example" --type SecureString --region us-west-2
  echo "Completed creating mongodb credentials in AWS SSM"
  echo "Successfully created AWS SSM parameters" >> aws-init.txt
else
    exit
fi