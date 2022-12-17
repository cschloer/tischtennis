package main

import (
	"fmt"
	"tischtennis/database"
	"tischtennis/helpers"

	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"strconv"
)

type DeleteGameRequest struct {
	PersonId string `json:"personId"`
	Created  string `json:"created"`
}

// Handler function Using AWS Lambda Proxy Request
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	/* Auth */
	err := helpers.CheckAccessKey(request, "")
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 401}, nil
	}

	bodyRequest := DeleteGameRequest{}

	err = json.Unmarshal([]byte(request.Body), &bodyRequest)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 404}, nil
	}

	createdInt, err := strconv.ParseInt(bodyRequest.Created, 10, 0)

	deletedId1, deletedId2, created, err := database.DeleteGame(
		bodyRequest.PersonId,
		createdInt,
	)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 500}, nil
	}

	var rs = map[string]interface{}{
		"success": true,
		"id1":     deletedId1,
		"id2":     deletedId2,
		"created": fmt.Sprintf("%d", created),
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
