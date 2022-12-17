package main

import (
	"tischtennis/database"
	"tischtennis/helpers"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type PersonPageData struct {
	Version           string
	BasePath          string
	Person            database.Person
	AlphSortedPeople  []database.Person
	Games             []database.Game
	PersonIdToNameMap map[string]string
}

// Handler function Using AWS Lambda Proxy Request
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	//Get the personId parameter that was sent
	personId := request.PathParameters["personId"]

	people, err := database.GetPeople()
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 500}, nil
	}

	var person database.Person
	found := false
	// TODO: is it faster to make a DDB request here?
	for _, p := range people {
		if p.Id == personId {
			person = p
			found = true
			break
		}
	}
	if !found {
		return events.APIGatewayProxyResponse{
			Headers:    map[string]string{"content-type": "text/html"},
			Body:       "That person doesn't exist",
			StatusCode: 404,
		}, nil
	}

	gamesMap, err := database.GetGames([]database.Person{person}, 10)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 500}, nil
	}

	personIdToNameMap := helpers.GetPersonIdToNameMap(people)

	data := PersonPageData{
		Version:           helpers.VERSION,
		BasePath:          helpers.BASE_PATH,
		Person:            person,
		AlphSortedPeople:  helpers.AlphSortPeople(people),
		Games:             gamesMap[person.Id],
		PersonIdToNameMap: personIdToNameMap,
	}

	body, err := helpers.BuildPage("templates/person.html", data)
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
