#!/usr/bin/env sh
# Build and tag the runner image.
#
#   ./build.sh                      → gofinity-runner:latest
#   RUNNER_IMAGE=foo:bar ./build.sh → foo:bar
#
# RUNNER_IMAGE is the same variable the API reads (see .env.example), so a
# custom tag built here is the one the API will run.
set -eu

IMAGE="${RUNNER_IMAGE:-gofinity-runner:latest}"
GO_VERSION="${RUNNER_GO_VERSION:-1.25}"

cd "$(dirname "$0")"

if ! command -v docker >/dev/null 2>&1; then
  echo "build.sh: docker is not installed" >&2
  exit 1
fi

echo "building $IMAGE (go $GO_VERSION)"
docker build --build-arg "GO_VERSION=$GO_VERSION" -t "$IMAGE" .
echo "built $IMAGE"
