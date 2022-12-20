package main

import (
	"tischtennis/database"
	"tischtennis/helpers"

	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type CreatePersonRequest struct {
	Name      string `json:"name"`
	FaIcon    string `json:"faIcon"`
	AccessKey string `json:"accessKey"`
}

// Handler function Using AWS Lambda Proxy Request
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	/* Auth */
	err := helpers.CheckAccessKey(request, "")
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 401}, nil
	}

	bodyRequest := CreatePersonRequest{}

	err = json.Unmarshal([]byte(request.Body), &bodyRequest)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 404}, nil
	}
	createdId, err := database.CreatePerson(
		bodyRequest.Name,
		bodyRequest.FaIcon,
		bodyRequest.AccessKey,
	)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 500}, nil
	}

	var rs = map[string]interface{}{
		"success": true,
		"id":      createdId,
	}

	response, err := json.Marshal(rs)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 500}, nil
	}
	return events.APIGatewayProxyResponse{Body: string(response), StatusCode: 200}, nil

}

func main() {
	lambda.Start(Handler)
}
