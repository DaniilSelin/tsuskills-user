#!/bin/bash
set -e

if [ -z "$1" ]; then
  echo "Version needed"
  exit 1
fi

VERSION=$1
IMAGE_NAME="daniilselin/tsu-skills-user"

echo "build image $IMAGE_NAME:$VERSION ..."
docker build -t $IMAGE_NAME:$VERSION -f docker/Dockerfile .

echo "add tag latest ..."
docker tag $IMAGE_NAME:$VERSION $IMAGE_NAME:latest

echo "push $IMAGE_NAME:$VERSION ..."
docker push $IMAGE_NAME:$VERSION

echo "push $IMAGE_NAME:latest ..."
docker push $IMAGE_NAME:latest

echo "Compleate! Access tah: $VERSION and latest"
