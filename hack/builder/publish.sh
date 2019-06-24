#!/usr/bin/env bash

SCRIPT_DIR="$(
    cd "$(dirname "$BASH_SOURCE[0]")"
    pwd
)"

docker tag alukiano/mhco-builder docker.io/alukiano/mhco-builder
docker push docker.io/alukiano/mhco-builder
