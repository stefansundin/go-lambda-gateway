FROM golang AS builder
WORKDIR /root
COPY . .
RUN \
  CGO_ENABLED=0 \
  GOOS=linux \
  go build -mod=readonly -ldflags="-s -w"

FROM busybox
LABEL maintainer="stefansundin https://github.com/stefansundin/go-lambda-gateway"
COPY --from=builder /root/go-lambda-gateway .
ENTRYPOINT ["./go-lambda-gateway"]
