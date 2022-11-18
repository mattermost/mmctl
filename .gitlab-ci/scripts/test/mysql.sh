#!/bin/bash
set -xe

if [[ "$GITLAB_CI" == "true" ]]; then
  export DIR_MATTERMOST_ROOT=$CI_PROJECT_DIR # e2e-ee
elif [[ "$GITLAB_CI" == "" ]]; then
  export DIR_MATTERMOST_ROOT=$PWD/..
  export CI_REGISTRY=registry.internal.mattermost.com
  export COMPOSE_PROJECT_NAME="1_mmctl_1"
  export IMAGE_BUILD_SERVER=$CI_REGISTRY/mattermost/ci/images/mattermost-build-server:20220415_golang-1.18.1
  # You need to log in to internal registry to run this script locally
fi

echo "$DOCKER_HOST"
docker ps
DOCKER_NETWORK=$COMPOSE_PROJECT_NAME
DOCKER_COMPOSE_FILE="gitlab-dc.mysql.yml"
CONTAINER_SERVER="${COMPOSE_PROJECT_NAME}_server_1"
docker network create $DOCKER_NETWORK
ulimit -n 8096
cd "$DIR_MATTERMOST_ROOT"/mattermost-server/build/
docker-compose -f $DOCKER_COMPOSE_FILE run -d --rm start_dependencies
sleep 5
docker-compose exec -T openldap bash -c 'ldapadd -x -D "cn=admin,dc=mm,dc=test,dc=com" -w mostest' < "$DIR_MATTERMOST_ROOT"/mattermost-server/tests/test-data.ldif
docker-compose exec -T minio sh -c 'mkdir -p /data/mattermost-test'
docker run --rm --name "${COMPOSE_PROJECT_NAME}_curl_mysql" --net $DOCKER_NETWORK $CI_REGISTRY/mattermost/ci/images/curl:7.59.0-1 sh -c "until curl --max-time 5 --output - http://mysql:3306; do echo waiting for mysql; sleep 5; done;"
docker run --rm --name "${COMPOSE_PROJECT_NAME}_curl_elasticsearch" --net $DOCKER_NETWORK $CI_REGISTRY/mattermost/ci/images/curl:7.59.0-1 sh -c "until curl --max-time 5 --output - http://elasticsearch:9200; do echo waiting for elasticsearch; sleep 5; done;"

docker run -d --name "$CONTAINER_SERVER" --net $DOCKER_NETWORK \
  --env-file="dotenv/test.env" \
  --env MM_SQLSETTINGS_DATASOURCE="mmuser:mostest@tcp(mysql:3306)/mattermost_test?charset=utf8mb4,utf8&multiStatements=true" \
  --env MM_SQLSETTINGS_DRIVERNAME=mysql \
  -v "$DIR_MATTERMOST_ROOT":/mattermost \
  -w /mattermost/mmctl \
  $IMAGE_BUILD_SERVER \
  bash -c 'ulimit -n 8096; ls -al; make test-all'

DIR_LOGS="$DIR_MATTERMOST_ROOT"/mmctl/logs
mkdir -p "$DIR_LOGS"
docker-compose logs --tail="all" -t --no-color > "$DIR_LOGS"/docker-compose_logs_$COMPOSE_PROJECT_NAME
docker ps -a --no-trunc > "$DIR_LOGS"/docker_ps_$COMPOSE_PROJECT_NAME
docker stats -a --no-stream > "$DIR_LOGS"/docker_stats_$COMPOSE_PROJECT_NAME
docker logs -f $CONTAINER_SERVER
tar -czvf "$DIR_LOGS"/docker_logs_$COMPOSE_PROJECT_NAME.tar.gz "$DIR_LOGS"/docker-compose_logs_$COMPOSE_PROJECT_NAME "$DIR_LOGS"/docker_ps_$COMPOSE_PROJECT_NAME "$DIR_LOGS"/docker_stats_$COMPOSE_PROJECT_NAME

DOCKER_EXIT_CODE=$(docker inspect $CONTAINER_SERVER --format='{{.State.ExitCode}}')
docker rm $CONTAINER_SERVER
echo "$DOCKER_EXIT_CODE"
exit "$DOCKER_EXIT_CODE"
