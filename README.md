This program is a tiny Amazon API Gateway emulator, letting you invoke Go Lambda functions (running locally) over HTTP. At the moment, it proxies all requests to a single function.

I'm sure there are bugs, and it's not very customizable. If you have use cases that aren't covered yet, feel free to submit pull requests!

You need to set `_LAMBDA_SERVER_PORT` when running your lambda to make it listen for requests on a port.

Based on https://github.com/djhworld/go-lambda-invoke.
