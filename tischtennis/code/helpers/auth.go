package helpers

import (
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"os"
	"strings"
	"tischtennis/database"
)

var ACCESS_KEY_HEADER_KEY = "x-wall-city-access-key"
var MASTER_ACCESS_KEY = os.Getenv("MASTER_ACCESS_KEY")

func CheckAccessKey(request events.APIGatewayProxyRequest, personId string) (err error) {
	accessKey := ""
	for key, value := range request.Headers {
		fmt.Println("key: '", key, "', access_key_header: '", ACCESS_KEY_HEADER_KEY, "'")
		if strings.ToLower(key) == ACCESS_KEY_HEADER_KEY {
			accessKey = value
			break
		}
	}

	if accessKey == MASTER_ACCESS_KEY {
		return nil
	}
	fmt.Println(accessKey, " -- ", MASTER_ACCESS_KEY, " :!!:::")
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
