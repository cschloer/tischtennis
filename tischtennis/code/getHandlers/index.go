package main

import (
	"fmt"
	"tischtennis/database"
	"tischtennis/helpers"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type IndexPageData struct {
	Version          string
	BasePath         string
	Title            string
	People           []database.Person
	AlphSortedPeople []database.Person
}

/*
var IndexTemplate = template.Must(
	template.ParseFiles("templates/base.html"),
)
*/

// Handler function Using AWS Lambda Proxy Request
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	//Get the path parameter that was sent
	// name := request.PathParameters["name"]

	//Generate message that want to be sent as body
	// message := fmt.Sprintf(" { \"Message\" : \"Hello %s \" } ", name)

	people := []database.Person{
		database.Person{
			Name:   "Lucas",
			Id:     "adfa-adfadfa",
			FaIcon: "fas fa-wave",
			Wins:   5,
			Losses: 12,
			Score:  0.3573,
		},
		database.Person{
			Name:   "Conrad",
			Id:     "adfadf134-01",
			FaIcon: "fas fa-wave",
			Wins:   12,
			Losses: 5,
			Score:  0.6573,
		},
	}

	data := IndexPageData{
		Version:          helpers.VERSION,
		BasePath:         helpers.BASE_PATH,
		Title:            "Tischtennis",
		People:           people,
		AlphSortedPeople: helpers.AlphSortPeople(people),
	}
	fmt.Println("BASE PATH", helpers.BASE_PATH)

	// IndexTemplate.ExecuteTemplate(response, "base", data)

	//Returning response with AWS Lambda Proxy Response
	return events.APIGatewayProxyResponse{
		Headers:    map[string]string{"content-type": "text/html"},
		Body:       helpers.BuildPage("templates/index.html", data).String(),
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(Handler)
}
