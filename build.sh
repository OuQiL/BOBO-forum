#!/bin/bash

set -e

REGISTRY=${REGISTRY:-"localhost:5000"}
VERSION=${VERSION:-"latest"}

echo "Building Docker images..."

docker build -t ${REGISTRY}/bobo-forum/gateway:${VERSION} ./api-gateway

echo "Building auth service..."
docker build -t ${REGISTRY}/bobo-forum/auth:${VERSION} ./service/auth

echo "Building post service..."
docker build -t ${REGISTRY}/bobo-forum/post:${VERSION} ./service/post

echo "Building search service..."
docker build -t ${REGISTRY}/bobo-forum/search:${VERSION} ./service/search

echo "Building interaction service..."
docker build -t ${REGISTRY}/bobo-forum/interaction:${VERSION} ./service/interaction

echo "Pushing images to registry..."

docker push ${REGISTRY}/bobo-forum/gateway:${VERSION}
docker push ${REGISTRY}/bobo-forum/auth:${VERSION}
docker push ${REGISTRY}/bobo-forum/post:${VERSION}
docker push ${REGISTRY}/bobo-forum/search:${VERSION}
docker push ${REGISTRY}/bobo-forum/interaction:${VERSION}

echo "Done! All images built and pushed."
