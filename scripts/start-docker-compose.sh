#!/usr/bin/env bash

set -e
set -u
set -o pipefail

REPO_ROOT=$(git rev-parse --show-toplevel)

docker-compose --no-ansi run --rm start_dependencies

cat "$REPO_ROOT/tests/test-data.ldif" | docker-compose --no-ansi exec -T openldap bash -c 'ldapadd -x -D "cn=admin,dc=mm,dc=test,dc=com" -w mostest';

docker-compose --no-ansi exec -T minio sh -c 'mkdir -p /data/mattermost-test';
docker-compose --no-ansi ps