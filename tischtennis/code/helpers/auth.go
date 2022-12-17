package helpers

import (
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"os"
	"tischtennis/database"
)

var ACCESS_KEY_HEADER_KEY = "X-Wall-City-Access-Key"
var MASTER_ACCESS_KEY = os.Getenv("MASTER_ACCESS_KEY")

func CheckAccessKey(request events.APIGatewayProxyRequest, personId string) (err error) {
	accessKey := ""
	for key, value := range request.Headers {
		if key == ACCESS_KEY_HEADER_KEY {
			accessKey = value
			break
		}
	}

	if accessKey == MASTER_ACCESS_KEY {
		return nil
	}
	if personId != "" {
		personAccessKey, err := database.GetPersonAccessKey(personId)
		if err != nil {
			return err
		}
		if accessKey == personAccessKey {
			return nil
		}
	}
	return errors.New("Invalid access key")

}
