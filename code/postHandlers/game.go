package main

import (
	"tischtennis/database"
	"tischtennis/helpers"

	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type CreateGameRequest struct {
	ReporterId    string `json:"reporterId"`
	OtherPersonId string `json:"otherPersonId"`
	Wins          int32  `json:"wins"`
	Losses        int32  `json:"losses"`
}

// Handler function Using AWS Lambda Proxy Request
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	bodyRequest := CreateGameRequest{}

	err := json.Unmarshal([]byte(request.Body), &bodyRequest)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 404}, nil
	}
	/* Auth */
	err = helpers.CheckAccessKey(request, bodyRequest.ReporterId, true)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 401}, nil
	}

	createdId1, createdId2, created, err := database.CreateGame(
		bodyRequest.ReporterId,
		bodyRequest.OtherPersonId,
		bodyRequest.Wins,
		bodyRequest.Losses,
	)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 500}, nil
	}

	var rs = map[string]interface{}{
		"success": true,
		"id1":     createdId1,
		"id2":     createdId2,
		"created": created,
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
