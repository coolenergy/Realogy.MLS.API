#!/usr/bin/env bash

echo "starting..."
docker-compose -f docker-compose.test.yml run --rm listd_tests

echo "finished..."
