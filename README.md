This program is a tiny Amazon API Gateway emulator, letting you invoke Go AWS Lambda functions (running locally) over HTTP. At the moment, it proxies all requests to a single function.

I'm sure there are bugs, and it's not very customizable. If you have use cases that aren't covered yet, feel free to submit pull requests!

You need to set `_LAMBDA_SERVER_PORT` when running your lambda to make it listen for requests on a port.

This project supports the HTTP API payload format, but you need to set `PAYLOAD_FORMAT_VERSION=2.0` to use it. See [examples/http-api](examples/http-api) and [the documentation](https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-develop-integrations-lambda.html).

Docker image available: https://hub.docker.com/r/stefansundin/go-lambda-gateway

You can also use the program without using Docker:

```shell
go install github.com/stefansundin/go-lambda-gateway@latest
```

Note: Beware of the capitalization of your headers. This program uses Go's `net/http` server, which will normalize the capitalization of your headers according to its own `CanonicalHeaderKey` function, whereas Amazon API Gateway does not manipulate the capitalization at all (but if you send the same header multiple times with different capitalization, it will use the first capitalization).

You may also be interested in [go-lambda-invoke](https://github.com/stefansundin/go-lambda-invoke).
