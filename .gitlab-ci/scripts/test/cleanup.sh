#!/bin/bash
set -xe

TO_STOP=$(docker ps -q)
for i in $TO_STOP; do
  docker stop "$i"
done
docker system prune
docker network remove "1_mmctl_1"
docker network remove "1_mmctl_2"
