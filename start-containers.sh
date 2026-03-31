#!/usr/bin/env bash
set -euo pipefail

NETWORK_NAME="url-shortener-net"
PROJECT_ROOT="/Users/ronitboddu/Documents/Projects/system-design-url-shortener"

remove_container_if_exists() {
  local name="$1"
  if docker ps -a --format '{{.Names}}' | grep -Fxq "$name"; then
    docker rm -f "$name" >/dev/null 2>&1 || true
  fi
}

echo "Creating Docker network if needed..."
if ! docker network ls --format '{{.Name}}' | grep -Fxq "$NETWORK_NAME"; then
  docker network create "$NETWORK_NAME"
fi

echo "Removing old containers if they exist..."
for c in postgres-db db-service1 db-service2 db-service3 go-service1 go-service2 go-service3 python-lb go-lb; do
  remove_container_if_exists "$c"
done

echo "Building Python DB service image..."
docker build -t url-shortener-db-service "$PROJECT_ROOT/db-service"

echo "Building Go service image..."
docker build -t url-shortener-go-service "$PROJECT_ROOT/server"

echo "Starting Postgres..."
docker run -d \
  --name postgres-db \
  --network "$NETWORK_NAME" \
  -p 5432:5432 \
  -e POSTGRES_USER=ronitboddu \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=postgres \
  postgres:16

echo "Waiting a few seconds for Postgres startup..."
sleep 5

echo "Starting Python DB services..."
docker run -d \
  --name db-service1 \
  --network "$NETWORK_NAME" \
  -e DB_HOST=postgres-db \
  -e DB_PORT=5432 \
  -e DB_NAME=postgres \
  -e DB_USER=ronitboddu \
  -e DB_PASSWORD=postgres \
  -e SNOWFLAKE_NODE_ID=1 \
  url-shortener-db-service

docker run -d \
  --name db-service2 \
  --network "$NETWORK_NAME" \
  -e DB_HOST=postgres-db \
  -e DB_PORT=5432 \
  -e DB_NAME=postgres \
  -e DB_USER=ronitboddu \
  -e DB_PASSWORD=postgres \
  -e SNOWFLAKE_NODE_ID=2 \
  url-shortener-db-service

docker run -d \
  --name db-service3 \
  --network "$NETWORK_NAME" \
  -e DB_HOST=postgres-db \
  -e DB_PORT=5432 \
  -e DB_NAME=postgres \
  -e DB_USER=ronitboddu \
  -e DB_PASSWORD=postgres \
  -e SNOWFLAKE_NODE_ID=3 \
  url-shortener-db-service

echo "Starting Python load balancer..."
docker run -d \
  --name python-lb \
  --network "$NETWORK_NAME" \
  -p 8085:8000 \
  -v "$PROJECT_ROOT/nginx/python-lb.conf:/etc/nginx/conf.d/default.conf:ro" \
  nginx:latest

echo "Starting Go services..."
docker run -d \
  --name go-service1 \
  --network "$NETWORK_NAME" \
  -e DB_SERVICE_BASE_URL=http://python-lb:8000 \
  url-shortener-go-service

docker run -d \
  --name go-service2 \
  --network "$NETWORK_NAME" \
  -e DB_SERVICE_BASE_URL=http://python-lb:8000 \
  url-shortener-go-service

docker run -d \
  --name go-service3 \
  --network "$NETWORK_NAME" \
  -e DB_SERVICE_BASE_URL=http://python-lb:8000 \
  url-shortener-go-service

echo "Starting Go load balancer..."
docker run -d \
  --name go-lb \
  --network "$NETWORK_NAME" \
  -p 8080:8080 \
  -v "$PROJECT_ROOT/nginx/go-lb.conf:/etc/nginx/conf.d/default.conf:ro" \
  nginx:latest

echo
echo "All containers started."
echo "Go load balancer:     http://localhost:8080"
echo "Python load balancer: http://localhost:8085"
echo
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
