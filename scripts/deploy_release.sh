#!/bin/sh

set -eu

if [ "${1:-}" = "" ]; then
  echo "usage: $0 <release-tag>" >&2
  exit 1
fi

RELEASE_TAG="$1"
APP_ROOT="${APP_ROOT:-/opt/xmqibu-frontier-site}"
APP_NAME="${APP_NAME:-frontier-site-web}"
DOCKER_NETWORK="${DOCKER_NETWORK:-nginx-proxy}"
ENV_FILE="${ENV_FILE:-$APP_ROOT/config/frontier-site.env}"
HOST_MEDIA_DIR="${HOST_MEDIA_DIR:-$APP_ROOT/uploads}"
RELEASES_DIR="${RELEASES_DIR:-$APP_ROOT/releases}"
HOST_NGINX_BIN="${HOST_NGINX_BIN:-/usr/local/nginx/sbin/nginx}"
VHOST_FILE="${VHOST_FILE:-/usr/local/nginx/conf/vhost/xmqibu-https.conf}"
BINARY_PATH="${BINARY_PATH:-$APP_ROOT/frontier-site}"
DOCKERFILE_PATH="${DOCKERFILE_PATH:-$APP_ROOT/Dockerfile}"
CONTAINER_MEDIA_DIR="${CONTAINER_MEDIA_DIR:-/uploads}"
PORT="${PORT:-18092}"
IMAGE="frontier-site:${RELEASE_TAG}"
CONTAINER_NAME="${APP_NAME}-${RELEASE_TAG}"

if [ ! -f "$ENV_FILE" ]; then
  echo "env file not found: $ENV_FILE" >&2
  exit 1
fi

if [ ! -f "$BINARY_PATH" ]; then
  echo "binary not found: $BINARY_PATH" >&2
  exit 1
fi

if [ ! -f "$DOCKERFILE_PATH" ]; then
  echo "dockerfile not found: $DOCKERFILE_PATH" >&2
  exit 1
fi

mkdir -p "$RELEASES_DIR"

BUILD_DIR="$(mktemp -d)"
cleanup() {
  rm -rf "$BUILD_DIR"
}
trap cleanup EXIT

cp "$BINARY_PATH" "$BUILD_DIR/frontier-site"
cp "$DOCKERFILE_PATH" "$BUILD_DIR/Dockerfile"

docker build -t "$IMAGE" "$BUILD_DIR"
docker rm -f "$CONTAINER_NAME" >/dev/null 2>&1 || true
docker run -d \
  --name "$CONTAINER_NAME" \
  --env-file "$ENV_FILE" \
  -e "MEDIA_DIR=$CONTAINER_MEDIA_DIR" \
  -v "$HOST_MEDIA_DIR:$CONTAINER_MEDIA_DIR" \
  --network "$DOCKER_NETWORK" \
  "$IMAGE" >/dev/null

CONTAINER_IP="$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "$CONTAINER_NAME")"

if [ "$CONTAINER_IP" = "" ]; then
  echo "failed to resolve container ip" >&2
  docker logs "$CONTAINER_NAME" >&2 || true
  exit 1
fi

attempt=0
until curl -fsS "http://$CONTAINER_IP:$PORT/healthz" >/dev/null; do
  attempt=$((attempt + 1))
  if [ "$attempt" -ge 15 ]; then
    echo "health check failed for $CONTAINER_NAME" >&2
    docker logs "$CONTAINER_NAME" >&2 || true
    exit 1
  fi
  sleep 2
done

cp "$VHOST_FILE" "$VHOST_FILE.bak-$RELEASE_TAG"
perl -0pi -e "s#proxy_pass http://[^;]+:$PORT;#proxy_pass http://$CONTAINER_IP:$PORT;#" "$VHOST_FILE"

"$HOST_NGINX_BIN" -t
"$HOST_NGINX_BIN" -s reload

curl -kfsS --resolve xmqibu.com:443:127.0.0.1 https://xmqibu.com/ >/dev/null
curl -kfsS --resolve www.xmqibu.com:443:127.0.0.1 https://www.xmqibu.com/ >/dev/null

cat > "$RELEASES_DIR/$RELEASE_TAG.txt" <<EOF
release_tag=$RELEASE_TAG
container_name=$CONTAINER_NAME
image=$IMAGE
container_ip=$CONTAINER_IP
deployed_at=$(date '+%Y-%m-%d %H:%M:%S %z')
EOF

printf '%s\n' "$RELEASE_TAG" > "$RELEASES_DIR/current"

echo "deployed $CONTAINER_NAME at $CONTAINER_IP"
