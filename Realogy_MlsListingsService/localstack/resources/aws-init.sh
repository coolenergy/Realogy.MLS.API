#!/bin/bash

LOCALSTACK_SSM=`curl -sS http://127.0.0.1:4566/health | jq '.services.ssm?'`

if [[ "$LOCALSTACK_SSM" = "\"running\"" ]] ; then
  awslocal ssm put-parameter --name "/realogy/services/mls-listings-service/local/mls-mongodb-user" --value "root" --type SecureString --region us-west-2
  awslocal ssm put-parameter --name "/realogy/services/mls-listings-service/local/mls-mongodb-url" --value "mongo:27017" --type SecureString --region us-west-2
  awslocal ssm put-parameter --name "/realogy/services/mls-listings-service/local/mls-mongodb-pass" --value "example" --type SecureString --region us-west-2
  echo "Completed creating mongodb credentials in AWS SSM" >> aws-init.txt
  # awslocal ssm get-parameters --names "/realogy/services/mls-listings-service/local/mls-mongodb-url" "/realogy/services/mls-listings-service/local/mls-mongodb-user" "/realogy/services/mls-listings-service/local/mls-mongodb-pass"
else
    exit
fi