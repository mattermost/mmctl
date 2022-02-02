#!/usr/bin/env bash

set -e
set -u
set -o pipefail

sleep 5
docker run --rm --net "$COMPOSE_PROJECT_NAME"_mm-test appropriate/curl:latest sh -c "until curl --max-time 5 --output - http://mysql:3306; do echo waiting for mysql; sleep 5; done;"
docker run --rm --net "$COMPOSE_PROJECT_NAME"_mm-test appropriate/curl:latest sh -c "until curl --max-time 5 --output - http://elasticsearch:9200; do echo waiting for elasticsearch; sleep 5; done;"
