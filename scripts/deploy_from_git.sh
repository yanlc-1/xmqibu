#!/bin/sh

set -eu

APP_ROOT="${APP_ROOT:-/opt/xmqibu-frontier-site}"
REPO_DIR="${REPO_DIR:-$APP_ROOT/repo}"
REMOTE_NAME="${REMOTE_NAME:-origin}"
BRANCH_NAME="${BRANCH_NAME:-main}"
RELEASE_TAG="${RELEASE_TAG:-manual-$(date +%Y%m%d%H%M%S)}"
DEPLOY_SCRIPT="${DEPLOY_SCRIPT:-$REPO_DIR/scripts/deploy_release.sh}"
GOPROXY_VALUE="${GOPROXY_VALUE:-https://goproxy.cn,direct}"
RUNTIME_DOCKERFILE="${RUNTIME_DOCKERFILE:-$REPO_DIR/Dockerfile.release}"

mkdir -p "$APP_ROOT/bin" "$APP_ROOT/config" "$APP_ROOT/uploads" "$APP_ROOT/logs" "$APP_ROOT/releases"

if [ ! -d "$REPO_DIR/.git" ]; then
  echo "git repo not found: $REPO_DIR" >&2
  exit 1
fi

cd "$REPO_DIR"
git fetch "$REMOTE_NAME"
git checkout "$BRANCH_NAME"
git reset --hard "$REMOTE_NAME/$BRANCH_NAME"

GOPROXY="$GOPROXY_VALUE" CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "$APP_ROOT/frontier-site" .

if [ ! -f "$RUNTIME_DOCKERFILE" ]; then
  echo "runtime dockerfile not found: $RUNTIME_DOCKERFILE" >&2
  exit 1
fi

cp "$RUNTIME_DOCKERFILE" "$APP_ROOT/Dockerfile"
chmod +x "$DEPLOY_SCRIPT"
"$DEPLOY_SCRIPT" "$RELEASE_TAG"
