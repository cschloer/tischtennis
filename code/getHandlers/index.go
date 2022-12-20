package main

import (
	"tischtennis/database"
	"tischtennis/helpers"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type IndexPageData struct {
	Version           string
	BasePath          string
	StaticAssetsUrl   string
	Title             string
	ScoreSortedPeople []database.Person
	AlphSortedPeople  []database.Person
}

// Handler function Using AWS Lambda Proxy Request
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	people, err := database.GetPeople()
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 500}, nil
	}

	data := IndexPageData{
		Version:           helpers.VERSION,
		BasePath:          helpers.BASE_PATH,
		StaticAssetsUrl:   helpers.STATIC_ASSETS_URL,
		Title:             "Tischtennis",
		ScoreSortedPeople: helpers.ScoreSortPeople(people),
		AlphSortedPeople:  helpers.AlphSortPeople(people),
	}

	body, err := helpers.BuildPage("templates/index.html", data)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 500}, nil
	}
	//Returning response with AWS Lambda Proxy Response
	return events.APIGatewayProxyResponse{
		Headers:    map[string]string{"content-type": "text/html"},
		Body:       body.String(),
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(Handler)
}
