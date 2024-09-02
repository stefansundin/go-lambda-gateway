Copy a real certificate to `cert.crt` and `cert.key`, or generate a self-signed one:

```shell
openssl req -x509 -sha256 -days 3650 -nodes -out cert.crt -newkey rsa:4096 -keyout cert.key -subj "/CN=localhost" -addext "subjectAltName=DNS:localhost"
```

Start `go-lambda-gateway`:

```shell
docker compose up
```

Start the lambda function:

```shell
export _LAMBDA_SERVER_PORT=8001
go run toupperlambda.go
```

Open https://localhost/world in a browser, or use `curl`:

```shell
curl https://localhost/world
```

It should return `HELLO WORLD!`
