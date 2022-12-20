package main

import (
	"tischtennis/database"
	"tischtennis/helpers"

	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Handler function Using AWS Lambda Proxy Request
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	/* Auth */
	err := helpers.CheckAccessKey(request, "")
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 401}, nil
	}

	res, err := database.AdminDatabase()
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 500}, nil
	}
	/*
		err = database.CreateDatabase()
		if err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = database.ComputeScores()
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
	*/

	var rs = map[string]interface{}{
		"success": true,
		"message": res,
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
