#!/bin/bash -e
export GOOS=linux
export GOARCH=amd64
go build -ldflags="-s -w"

docker build -t stefansundin/go-lambda-gateway .
