#!/usr/bin/env bash

set -e
set -u
set -o pipefail

docker run --net "$COMPOSE_PROJECT_NAME"_mm-test \
  --env-file=dotenv/test.env \
  --env MM_SQLSETTINGS_DATASOURCE="postgres://mmuser:mostest@postgres:5432/mattermost_test?sslmode=disable&connect_timeout=10" \
  --env MM_SQLSETTINGS_DATASOURCE=postgres \
  -v $GITHUB_WORKSPACE:/go/src \
  -w /go/src/mmctl \
  mattermost/mattermost-build-server:20220415_golang-1.18.1 \
  bash -c 'ulimit -n 8096 && make coverage'
