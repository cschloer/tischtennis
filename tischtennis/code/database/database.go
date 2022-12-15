package database

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
	"os"

	"errors"
	"fmt"
)

type Person struct {
	Name   string
	Id     string
	FaIcon string
	Wins   int64
	Losses int64
	Score  float64
}
type Game struct {
	Id              string
	ReporterName    string
	ReporterId      int64
	OtherPersonName string
	OtherPersonId   int64
	Wins            int64
	Losses          int64
}

type DatabasePerson struct {
	Id     string
	Name   string
	FaIcon string
}

// Initialize a session that the SDK will use to load
// credentials from the shared credentials file ~/.aws/credentials
// and region from the shared configuration file ~/.aws/config.
var sess = session.Must(session.NewSessionWithOptions(session.Options{
	SharedConfigState: session.SharedConfigEnable,
}))

// Create DynamoDB client
var svc = dynamodb.New(sess)

var ENVIRONMENT = os.Getenv("ENVIRONMENT")

func GetPersonAccessKey(personId int) (personAccessKey string, err error) {
	return "", errors.New(fmt.Sprintf("No person found with id %d.", personId))

}

func _createId() string {
	return uuid.New().String()

}

func AdminDatabase() (res string, err error) {
	tableName := ENVIRONMENT + "_person"
	id := _createId()
	item := DatabasePerson{
		Id:     id,
		Name:   "Conrad",
		FaIcon: "fas fa-wave",
	}

	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		return "", err
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		return "", err
	}

	// return "Successfully added item to table " + tableName + " with id " + id, nil
	scanInput := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}
	result, err := svc.Scan(scanInput)
	if err != nil {
		return "", err
	}
	for _, i := range result.Items {
		person := DatabasePerson{}

		err = dynamodbattribute.UnmarshalMap(i, &person)

		if err != nil {
			return "", err
		}

		fmt.Println("Name: ", person.Name)
		fmt.Println("Id:", person.Id)
		fmt.Println()
	}

	return "", nil

}
