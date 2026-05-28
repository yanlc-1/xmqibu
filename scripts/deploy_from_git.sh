#!/bin/sh

set -eu

APP_ROOT="${APP_ROOT:-/opt/xmqibu-frontier-site}"
REPO_DIR="${REPO_DIR:-$APP_ROOT/repo}"
REMOTE_NAME="${REMOTE_NAME:-origin}"
BRANCH_NAME="${BRANCH_NAME:-main}"
RELEASE_TAG="${RELEASE_TAG:-manual-$(date +%Y%m%d%H%M%S)}"
DEPLOY_SCRIPT="${DEPLOY_SCRIPT:-$REPO_DIR/scripts/deploy_release.sh}"

mkdir -p "$APP_ROOT/bin" "$APP_ROOT/config" "$APP_ROOT/uploads" "$APP_ROOT/logs" "$APP_ROOT/releases"

if [ ! -d "$REPO_DIR/.git" ]; then
  echo "git repo not found: $REPO_DIR" >&2
  exit 1
fi

cd "$REPO_DIR"
git fetch "$REMOTE_NAME"
git checkout "$BRANCH_NAME"
git reset --hard "$REMOTE_NAME/$BRANCH_NAME"

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "$APP_ROOT/frontier-site" .
cp Dockerfile "$APP_ROOT/Dockerfile"
chmod +x "$DEPLOY_SCRIPT"
"$DEPLOY_SCRIPT" "$RELEASE_TAG"
