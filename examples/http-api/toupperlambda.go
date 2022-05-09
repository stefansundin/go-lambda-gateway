package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handleRequest(ctx context.Context, request events.APIGatewayV2HTTPRequest) (map[string]interface{}, error) {
	// r, _ := json.Marshal(request)
	// fmt.Fprintf(os.Stderr, "%v\n", string(r))

	text := "Hello " + request.RawPath[1:] + "!"
	fmt.Println(text)

	// Here we have the option to return either events.APIGatewayProxyResponse or map[string]interface{}
	// See https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-develop-integrations-lambda.html

	var response = make(map[string]interface{})
	response["message"] = strings.ToUpper(text)
	return response, nil
}

func main() {
	lambda.Start(handleRequest)
}
