package database

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/rs/xid"
	"os"
	"time"

	"errors"
	"fmt"
)

type Person struct {
	Id          string
	Score       float32
	Name        string
	FaIcon      string
	Wins        int32
	Losses      int32
	NumReported int32
	// TODO make sure person access key isn't being sent to the front end
}
type Game struct {
	PersonId      string
	Created       int64
	OtherPersonId string
	Reporter      bool
	Wins          int32
	Losses        int32
}

// Person with AccessKey
type Person_DANGEROUS struct {
	Id          string
	Score       float32
	Name        string
	FaIcon      string
	Wins        int32
	Losses      int32
	NumReported int32
	AccessKey   string
}

// Initialize a session that the SDK will use to load
// credentials from the shared credentials file ~/.aws/credentials
// and region from the shared configuration file ~/.aws/config.
var sess = session.Must(session.NewSessionWithOptions(session.Options{
	//SharedConfigState: session.SharedConfigEnable,
	Config: *aws.NewConfig().
		WithRegion("us-east-1").
		WithCredentials(credentials.NewStaticCredentials("DEFAULT_ACCESS_KEY", "DEFAULT_SECRET", "")).
		//WithEndpoint("http://0.0.0.0:8000"),
		WithEndpoint("http://172.17.0.1:8000"),
}))

// Create DynamoDB client
var svc = dynamodb.New(sess)

var ENVIRONMENT = os.Getenv("ENVIRONMENT")

func _createId() string {
	return xid.New().String()

}

func _getTableName(table string) string {
	return "tischtennis_" + ENVIRONMENT + "_" + table
}

func GetPeople() (people []Person, err error) {
	// TODO eventually set up pagination
	scanInput := &dynamodb.ScanInput{
		TableName: aws.String(_getTableName("person")),
	}
	result, err := svc.Scan(scanInput)
	if err != nil {
		return people, err
	}
	people = make([]Person, len(result.Items))
	for i, item := range result.Items {
		person := Person{}

		err = dynamodbattribute.UnmarshalMap(item, &person)

		if err != nil {
			return people, err
		}
		people[i] = person
	}
	return people, nil
}

func GetGames(people []Person, onlyReporter bool, limit int) (gamesMap map[string][]Game, err error) {
	// TODO add limit
	gamesMap = make(map[string][]Game)
	for _, person := range people {
		input := &dynamodb.QueryInput{
			TableName:              aws.String(_getTableName("game")),
			KeyConditionExpression: aws.String("PersonId = :hashKey"),
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":hashKey": {
					S: aws.String(person.Id),
				},
			},
		}

		result, err := svc.Query(input)
		if err != nil {
			return gamesMap, err
		}
		games := make([]Game, len(result.Items))
		for i, item := range result.Items {
			game := Game{}

			err = dynamodbattribute.UnmarshalMap(item, &game)
			fmt.Println("GAME", game)

			if err != nil {
				return gamesMap, err
			}
			games[i] = game
		}
		gamesMap[person.Id] = games
	}
	return gamesMap, nil
}

func GetPerson(personId string) (person Person, err error) {
	return Person{
		Name:   "Lucas",
		Id:     personId,
		FaIcon: "fas fa-chess-knight",
		Wins:   5,
		Losses: 12,
		Score:  0.3573,
	}, nil
}
func GetPersonAccessKey(personId int) (personAccessKey string, err error) {
	return "", errors.New(fmt.Sprintf("No person found with id %d.", personId))

}

func AdminDatabase() (res string, err error) {
	tableNamePerson := _getTableName("person")
	items := []Person_DANGEROUS{
		Person_DANGEROUS{
			Name:      "Lucas",
			Id:        _createId(),
			FaIcon:    "fas fa-chess-knight",
			Wins:      5,
			Losses:    12,
			Score:     0.3573,
			AccessKey: "123",
		},
		Person_DANGEROUS{
			Name:      "Conrad",
			Id:        _createId(),
			FaIcon:    "fas fa-water",
			Wins:      12,
			Losses:    5,
			Score:     0.6173,
			AccessKey: "123",
		},
		Person_DANGEROUS{
			Name:      "Christian",
			Id:        _createId(),
			FaIcon:    "fas fa-cat",
			Wins:      9,
			Losses:    3,
			Score:     0.4512,
			AccessKey: "123",
		},
	}

	ids := make([]string, 0)
	for _, item := range items {

		ids = append(ids, item.Id)
		av, err := dynamodbattribute.MarshalMap(item)
		if err != nil {
			return "", err
		}

		input := &dynamodb.PutItemInput{
			Item:      av,
			TableName: aws.String(tableNamePerson),
		}

		_, err = svc.PutItem(input)
		if err != nil {
			return "", err
		}
	}

	tableNameGame := _getTableName("game")
	created := time.Now().UnixNano()
	games := []Game{
		Game{
			PersonId:      ids[0],
			Created:       created,
			OtherPersonId: ids[1],
			Wins:          5,
			Losses:        6,
			Reporter:      false,
		},
		Game{
			PersonId:      ids[1],
			Created:       created,
			OtherPersonId: ids[0],
			Wins:          6,
			Losses:        5,
			Reporter:      true,
		},
		Game{
			PersonId:      ids[0],
			Created:       created + 1,
			OtherPersonId: ids[2],
			Wins:          11,
			Losses:        3,
			Reporter:      true,
		},
		Game{
			PersonId:      ids[2],
			Created:       created + 1,
			OtherPersonId: ids[0],
			Wins:          3,
			Losses:        11,
			Reporter:      true,
		},
	}
	for _, item := range games {

		av, err := dynamodbattribute.MarshalMap(item)
		if err != nil {
			return "", err
		}

		input := &dynamodb.PutItemInput{
			Item:      av,
			TableName: aws.String(tableNameGame),
		}

		_, err = svc.PutItem(input)
		if err != nil {
			return "", err
		}
	}

	// return "Successfully added item to table " + tableName + " with id " + id, nil
	scanInput := &dynamodb.ScanInput{
		TableName: aws.String(tableNameGame),
	}
	result, err := svc.Scan(scanInput)
	if err != nil {
		return "", err
	}
	for _, i := range result.Items {
		game := Game{}

		err = dynamodbattribute.UnmarshalMap(i, &game)

		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%s: %d", game.PersonId, game.Created), nil
	}

	return "", nil

}
