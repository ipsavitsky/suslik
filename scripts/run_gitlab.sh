#!/usr/bin/env bash
# Simple script to launch a local instance of Gitlab via podman
set -e

script_dir=$(dirname -- "$(readlink -f -- "${BASH_SOURCE[0]}")")

GITLAB_VERSION="17.6.2-ce.0"
GITLAB_HTTP_PORT=9999
GITLAB_INSECURE_PASSWORD=insecure1111

DOCKER="podman"
IMAGE_NAME="docker.io/gitlab/gitlab-ce:$GITLAB_VERSION"
GITLAB_HOME="$PWD/.gitlab-container"

mkdir -p "$GITLAB_HOME/config"
mkdir -p "$GITLAB_HOME/logs"
mkdir -p "$GITLAB_HOME/data"

$DOCKER run --rm -ti \
  --hostname gitlab.example.com \
  --env GITLAB_ROOT_PASSWORD=$GITLAB_INSECURE_PASSWORD \
  --env GITLAB_OMNIBUS_CONFIG="$(cat "${script_dir}"/gitlab_config.rb)" \
  --publish $GITLAB_HTTP_PORT:$GITLAB_HTTP_PORT \
  --name suslik-gitlab-integration-test \
  --volume "$GITLAB_HOME/config:/etc/gitlab" \
  --volume "$GITLAB_HOME/logs:/var/log/gitlab" \
  --volume "$GITLAB_HOME/data:/var/opt/gitlab" \
  --shm-size 256m \
  $IMAGE_NAME
