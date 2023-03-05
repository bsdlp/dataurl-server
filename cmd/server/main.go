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
	if request.HTTPMethod != "POST" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusMethodNotAllowed,
			Body:       http.StatusText(http.StatusMethodNotAllowed),
		}, nil
	}

	if request.IsBase64Encoded {
		bs, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusBadRequest,
				Body:       "invalid dataurl",
			}, nil
		}
		data = string(bs)
	}

	d, err := dataurl.DecodeString(data)
	if err != nil {
		log.Printf("dataurl: %s", data)
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
