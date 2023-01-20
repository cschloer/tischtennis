package helpers

import (
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"os"
	"strings"
	"tischtennis/database"
)

var ACCESS_KEY_HEADER_KEY = "x-wall-city-access-key"
var MASTER_ACCESS_KEY = os.Getenv("MASTER_ACCESS_KEY")
var PUBLIC_ACCESS_KEY = os.Getenv("PUBLIC_ACCESS_KEY")

func CheckAccessKey(request events.APIGatewayProxyRequest, personId string, publicAccesKey bool) (err error) {
	accessKey := ""
	for key, value := range request.Headers {
		if strings.ToLower(key) == ACCESS_KEY_HEADER_KEY {
			accessKey = value
			break
		}
	}

	if accessKey == MASTER_ACCESS_KEY {
		return nil
	}
	// A public access key that allows access to certain endpoints
	if publicAccesKey && PUBLIC_ACCESS_KEY != "" && accessKey == PUBLIC_ACCESS_KEY {
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
