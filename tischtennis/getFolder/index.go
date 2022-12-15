package main

import (
	"bytes"
	"html/template"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type IndexPageData struct {
	Version string
}

/*
var IndexTemplate = template.Must(
	template.ParseFiles("templates/base.html"),
)
*/

func BuildPage(path string, data interface{}) *bytes.Buffer {
	var bodyBuffer bytes.Buffer
	t := template.Must(template.ParseFiles(path, "templates/base.html"))
	t.ExecuteTemplate(&bodyBuffer, "base", data)
	return &bodyBuffer
}

// Handler function Using AWS Lambda Proxy Request
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	//Get the path parameter that was sent
	// name := request.PathParameters["name"]

	//Generate message that want to be sent as body
	// message := fmt.Sprintf(" { \"Message\" : \"Hello %s \" } ", name)

	data := IndexPageData{
		Version: "0.1",
	}
	// IndexTemplate.ExecuteTemplate(response, "base", data)

	//Returning response with AWS Lambda Proxy Response
	return events.APIGatewayProxyResponse{
		Headers:    map[string]string{"content-type": "text/html"},
		Body:       BuildPage("templates/index.html", data).String(),
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(Handler)
}
