package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// r, _ := json.Marshal(request)
	// fmt.Fprintf(os.Stderr, "%v\n", string(r))

	text := "Hello " + request.Path[1:] + "!\n"
	fmt.Print(text)

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "text/plain",
		},
		Body: strings.ToUpper(text),
	}, nil
}

func main() {
	lambda.Start(handleRequest)
}
