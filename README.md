This program is a tiny Amazon API Gateway emulator, letting you invoke Go Lambda functions (running locally) over HTTP. At the moment, it proxies all requests to a single function.

I'm sure there are bugs, and it's not very customizable. If you have use cases that aren't covered yet, feel free to submit pull requests!

You need to set `_LAMBDA_SERVER_PORT` when running your lambda to make it listen for requests on a port.

Docker image available: https://hub.docker.com/r/stefansundin/go-lambda-gateway

Note: Beware of the capitalization of your headers. This program uses Go's `net/http` server, which will normalize the capitalization of your headers according to its own `CanonicalHeaderKey` function, whereas Amazon API Gateway does not manipulate the capitalization at all (but if you send the same header multiple times with different capitalization, it will use the first capitalization).

You may also be interested in [go-lambda-invoke](https://github.com/stefansundin/go-lambda-invoke).
