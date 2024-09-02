#!/bin/bash -e

# vanilla build:
# docker build --pull -t stefansundin/go-lambda-gateway .

# multi-arch build:
docker buildx create --use --name multiarch --node multiarch0
docker buildx build -t stefansundin/go-lambda-gateway --platform linux/amd64,linux/arm64,linux/arm/v7 --pull --no-cache --push .
