Start `go-lambda-gateway`:

```shell
docker compose up
```

Start the lambda function:

```shell
export _LAMBDA_SERVER_PORT=8001
go run toupperlambda.go
```

Open http://localhost:8002/world in a browser, or use `curl`:

```shell
curl http://localhost:8002/world
```

It should return `HELLO WORLD!`
