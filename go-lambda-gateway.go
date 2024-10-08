package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/rpc"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda/messages"
	"github.com/mitchellh/mapstructure"
)

var port int
var protocol string = "http"
var lambdaHost string
var payloadFormatVersion string

func IsBinary(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII || !unicode.IsPrint(r) {
			return true
		}
	}
	return false
}

func invokeLambda(payload []byte) (*events.APIGatewayProxyResponse, error) {
	deadline := time.Now().Add(60 * time.Minute)
	invokeRequest := &messages.InvokeRequest{
		Payload:      payload,
		RequestId:    "0",
		XAmznTraceId: "",
		Deadline: messages.InvokeRequest_Timestamp{
			Seconds: int64(deadline.Unix()),
			Nanos:   int64(deadline.Nanosecond()),
		},
		InvokedFunctionArn:    "",
		CognitoIdentityId:     "",
		CognitoIdentityPoolId: "",
		ClientContext:         nil,
	}

	client, err := rpc.Dial("tcp", lambdaHost)
	if err != nil {
		return nil, err
	}
	var invokeResponse messages.InvokeResponse
	if err = client.Call("Function.Invoke", invokeRequest, &invokeResponse); err != nil {
		return nil, err
	}
	if invokeResponse.Error != nil {
		return nil, errors.New(invokeResponse.Error.Message)
	}

	var response events.APIGatewayProxyResponse
	if payloadFormatVersion == "2.0" {
		var jsonData map[string]interface{}
		err = json.Unmarshal(invokeResponse.Payload, &jsonData)
		if err != nil {
			return nil, err
		}
		if _, ok := jsonData["statusCode"]; ok {
			mapstructure.Decode(jsonData, &response)
		} else {
			response.IsBase64Encoded = false
			response.StatusCode = 200
			response.Body = string(invokeResponse.Payload)
			response.Headers = map[string]string{
				"Content-Type": "application/json",
			}
		}
	} else {
		err = json.Unmarshal(invokeResponse.Payload, &response)
		if err != nil {
			return nil, err
		}
	}

	return &response, nil
}

func handleRequestV1(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}

	portDelimiterIndex := strings.LastIndexByte(r.RemoteAddr, ':')
	remoteIP := r.RemoteAddr[0:portDelimiterIndex]

	request := &events.APIGatewayProxyRequest{
		Resource:   "/",
		Path:       r.URL.Path,
		HTTPMethod: r.Method,
		Headers: map[string]string{
			"Host": r.Host,
		},
		MultiValueHeaders: map[string][]string{
			"Host": {r.Host},
		},
		QueryStringParameters:           map[string]string{},
		MultiValueQueryStringParameters: map[string][]string{},
		PathParameters:                  nil,
		StageVariables:                  nil,
		RequestContext:                  events.APIGatewayProxyRequestContext{},
		Body:                            string(body),
		IsBase64Encoded:                 false,
	}

	if r.URL.Path != "/" {
		request.Resource = "/{proxy+}"
		request.PathParameters = map[string]string{
			"proxy": r.URL.Path[1:],
		}
	}

	for header, values := range r.Header {
		for _, value := range values {
			request.Headers[header] = value
			request.MultiValueHeaders[header] = append(request.MultiValueHeaders[header], value)
		}
	}

	request.Headers["x-amzn-trace-id"] = "Root=0-00000000-000000000000000000000000"
	request.Headers["x-forwarded-for"] = remoteIP
	request.Headers["x-forwarded-port"] = strconv.Itoa(port)
	request.Headers["x-forwarded-proto"] = protocol

	for key, values := range r.URL.Query() {
		for _, value := range values {
			request.QueryStringParameters[key] = value
			request.MultiValueQueryStringParameters[key] = append(request.MultiValueQueryStringParameters[key], value)
		}
	}

	if IsBinary(request.Body) {
		request.IsBase64Encoded = true
		request.Body = base64.StdEncoding.EncodeToString(body)
	}

	payload, err := json.Marshal(request)
	if err != nil {
		log.Printf("Error marshalling json: %v", err)
		http.Error(w, "Error marshalling json", http.StatusInternalServerError)
		return
	}

	response, err := invokeLambda(payload)
	if err != nil {
		log.Printf("Error invoking lambda: %v", err)
		http.Error(w, "Error invoking lambda", http.StatusInternalServerError)
		return
	}
	// fmt.Printf("Response: %v\n", response)

	for header, value := range response.Headers {
		w.Header().Set(header, value)
	}
	w.WriteHeader(response.StatusCode)
	if response.IsBase64Encoded {
		bytes, err := base64.StdEncoding.DecodeString(response.Body)
		if err != nil {
			log.Printf("Error base64-decoding response body: %v", err)
			http.Error(w, "Error base64-decoding response body", http.StatusInternalServerError)
			return
		}
		w.Write(bytes)
	} else {
		fmt.Fprintf(w, response.Body)
	}

	// Log something similar to the common log format
	// host [date] request status bytes
	fmt.Printf("%s [%v] \"%s %s\" %v\n", r.Host, time.Now().Format("2006-01-02 15:04:05"), r.Method, r.URL.Path, len(response.Body))
}

func handleRequestV2(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}

	now := time.Now()
	portDelimiterIndex := strings.LastIndexByte(r.RemoteAddr, ':')
	remoteIP := r.RemoteAddr[0:portDelimiterIndex]
	domainName := r.Host
	if strings.ContainsRune(domainName, ':') {
		domainName = domainName[0:strings.IndexRune(domainName, ':')]
	}

	request := &events.APIGatewayV2HTTPRequest{
		Version:        "2.0",
		RouteKey:       "$default",
		RawPath:        r.URL.Path,
		RawQueryString: r.URL.RawQuery,
		Cookies:        []string{},
		Headers: map[string]string{
			"host": r.Host,
		},
		QueryStringParameters: map[string]string{},
		PathParameters:        map[string]string{},
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			RouteKey:     "$default",
			AccountID:    "anonymous",
			Stage:        "$default",
			RequestID:    "todo",
			APIID:        domainName,
			DomainName:   domainName,
			DomainPrefix: domainName,
			Time:         now.Format("02/Jan/2006:15:04:05 -0700"),
			TimeEpoch:    now.UnixMilli(),
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method:   r.Method,
				Path:     r.URL.Path,
				Protocol: r.Proto,
				SourceIP: remoteIP,
			},
		},
		StageVariables:  nil,
		Body:            string(body),
		IsBase64Encoded: false,
	}

	for _, c := range r.Cookies() {
		request.Cookies = append(request.Cookies, c.Name+"="+c.Value)
	}

	for header, values := range r.Header {
		h := strings.ToLower(header)
		for _, value := range values {
			request.Headers[h] = value
			if h == "user-agent" {
				request.RequestContext.HTTP.UserAgent = value
			}
		}
	}

	request.Headers["x-amzn-trace-id"] = "Root=0-00000000-000000000000000000000000"
	request.Headers["x-forwarded-for"] = remoteIP
	request.Headers["x-forwarded-port"] = strconv.Itoa(port)
	request.Headers["x-forwarded-proto"] = protocol

	for key, values := range r.URL.Query() {
		for _, value := range values {
			request.QueryStringParameters[key] = value
		}
	}

	if IsBinary(request.Body) {
		request.IsBase64Encoded = true
		request.Body = base64.StdEncoding.EncodeToString(body)
	}

	payload, err := json.Marshal(request)
	if err != nil {
		log.Printf("Error marshalling json: %v", err)
		http.Error(w, "Error marshalling json", http.StatusInternalServerError)
		return
	}

	response, err := invokeLambda(payload)
	if err != nil {
		log.Printf("Error invoking lambda: %v", err)
		http.Error(w, "Error invoking lambda", http.StatusInternalServerError)
		return
	}
	// fmt.Printf("Response: %v\n", response)

	for header, value := range response.Headers {
		w.Header().Set(header, value)
	}
	w.WriteHeader(response.StatusCode)
	if response.IsBase64Encoded {
		bytes, err := base64.StdEncoding.DecodeString(response.Body)
		if err != nil {
			log.Printf("Error base64-decoding response body: %v", err)
			http.Error(w, "Error base64-decoding response body", http.StatusInternalServerError)
			return
		}
		w.Write(bytes)
	} else {
		fmt.Fprintf(w, response.Body)
	}

	// Log something similar to the common log format
	// host [date] request status bytes
	fmt.Printf("%s [%v] \"%s %s\" %v\n", r.Host, now.Format("2006-01-02 15:04:05"), r.Method, r.URL.Path, len(response.Body))
}

func main() {
	lambdaHost = os.Getenv("LAMBDA_HOST")
	if lambdaHost == "" {
		lambdaHost = "localhost:8001"
	}
	fmt.Fprintf(os.Stderr, "Lambda address: %s\n", lambdaHost)

	var keyfile, certfile string
	keyfiles, err := filepath.Glob("*.key")
	if err == nil && len(keyfiles) > 0 {
		keyfile = keyfiles[0]
		certfile = keyfile[0:len(keyfile)-3] + "crt"
		_, err = os.Stat(certfile)
		if errors.Is(err, fs.ErrNotExist) {
			fmt.Fprintf(os.Stderr, "Warning: Found %s but not %s.\n", keyfile, certfile)
		} else {
			protocol = "https"
			fmt.Fprintf(os.Stderr, "TLS certificate files: %s and %s\n", certfile, keyfile)
		}
	}

	port, _ = strconv.Atoi(os.Getenv("PORT"))
	if port == 0 {
		if protocol == "https" {
			port = 443
		} else {
			port = 8002
		}
	}
	fmt.Fprintf(os.Stderr, "Listening on port: %d\n", port)
	u := url.URL{
		Scheme: protocol,
		Host:   "localhost",
		Path:   "/",
	}
	if (protocol == "http" && port != 80) || (protocol == "https" && port != 443) {
		u.Host += fmt.Sprintf(":%d", port)
	}
	fmt.Fprintf(os.Stderr, "URL: %s\n", u.String())

	payloadFormatVersion = os.Getenv("PAYLOAD_FORMAT_VERSION")
	if payloadFormatVersion == "" {
		payloadFormatVersion = "1.0"
	}
	if payloadFormatVersion == "1.0" {
		http.HandleFunc("/", handleRequestV1)
	} else if payloadFormatVersion == "2.0" {
		http.HandleFunc("/", handleRequestV2)
	} else {
		fmt.Fprintf(os.Stderr, "Error: unknown payload format version: %s\n", payloadFormatVersion)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Payload format version: %s\n", payloadFormatVersion)
	fmt.Fprintln(os.Stderr)

	addr := fmt.Sprintf(":%d", port)
	if protocol == "http" {
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	} else {
		err := http.ListenAndServeTLS(addr, certfile, keyfile, nil)
		if err != nil {
			log.Fatal("ListenAndServeTLS: ", err)
		}
	}
}
