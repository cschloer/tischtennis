package main

import (
	"tischtennis/database"
	"tischtennis/helpers"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type AdminPageData struct {
	Version           string
	BasePath          string
	StaticAssetsUrl   string
	AlphSortedPeople  []database.Person
	GamesMap          map[string][]database.Game
	PersonIdToNameMap map[string]string
}

// Handler function Using AWS Lambda Proxy Request
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	people, err := database.GetPeople()
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 500}, nil
	}

	gamesMap, err := database.GetGames(people, 15)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 500}, nil
	}

	// We can create a personIdToNameMap here without using the database

	personIdToNameMap := helpers.GetPersonIdToNameMap(people)

	data := AdminPageData{
		Version:           helpers.VERSION,
		BasePath:          helpers.BASE_PATH,
		StaticAssetsUrl:   helpers.STATIC_ASSETS_URL,
		AlphSortedPeople:  helpers.AlphSortPeople(people),
		GamesMap:          gamesMap,
		PersonIdToNameMap: personIdToNameMap,
	}

	body, err := helpers.BuildPage("templates/admin.html", data)
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
