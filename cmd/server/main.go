package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/vincent-petithory/dataurl"
)

func main() {
	lambda.Start(handle)
}

func handle(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var data string
	switch request.HTTPMethod {
	case "POST":
		data = request.Body
	case "GET":
		var ok bool
		data, ok = request.QueryStringParameters["d"]
		if !ok {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusBadRequest,
				Body:       "query parameter 'd' with dataurl value required",
			}, nil
		}
	default:
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusMethodNotAllowed,
			Body:       http.StatusText(http.StatusMethodNotAllowed),
		}, nil
	}

	d, err := dataurl.DecodeString(data)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       http.StatusText(http.StatusBadRequest),
		}, nil
	}

	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	_, err = encoder.Write(d.Data)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       http.StatusText(http.StatusInternalServerError),
		}, nil
	}
	err = encoder.Close()
	if err != nil {
		log.Printf("error closing base64 encoder: %s", err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"content-type": d.ContentType(),
		},
		Body:            buf.String(),
		IsBase64Encoded: true,
	}, nil
}
