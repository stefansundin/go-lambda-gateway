FROM busybox
MAINTAINER stefansundin https://github.com/stefansundin/go-lambda-gateway

COPY go-lambda-gateway .

ENTRYPOINT ["./go-lambda-gateway"]
