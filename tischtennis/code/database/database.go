package database

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
	"os"

	"errors"
	"fmt"
)

type Person struct {
	Id          string
	Score       float64
	Name        string
	FaIcon      string
	Wins        int64
	Losses      int64
	NumReported int64
	// TODO make sure person access key isn't being sent to the front end
}
type Game struct {
	PersonId      string
	Created       int64
	OtherPersonId string
	Reporter      bool
	Wins          int64
	Losses        int64
}

// Person with AccessKey
type Person_DANGEROUS struct {
	Id        string
	Score     float64
	Name      string
	FaIcon    string
	Wins      int64
	Losses    int64
	AccessKey string
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

func GetPeople() (people []Person, err error) {
	people = []Person{
		Person{
			Name:   "Lucas",
			Id:     "abc",
			FaIcon: "fas fa-chess-knight",
			Wins:   5,
			Losses: 12,
			Score:  0.3573,
		},
		Person{
			Name:   "Conrad",
			Id:     "efg",
			FaIcon: "fas fa-water",
			Wins:   12,
			Losses: 5,
			Score:  0.6173,
		},
		Person{
			Name:   "Christian",
			Id:     "hij",
			FaIcon: "fas fa-cat",
			Wins:   9,
			Losses: 3,
			Score:  0.4512,
		},
	}
	return people, nil
}

func GetGames(people []Person, onlyReporter bool, limit int) (games map[string][]Game, err error) {
	games = make(map[string][]Game)
	// A map to add people names to the games
	for _, person := range people {
		// TODO make query to to get games
		personGames := []Game{
			Game{
				PersonId:      person.Id,
				OtherPersonId: "abc",
				Wins:          4,
				Losses:        1,
				Reporter:      false,
			},
			Game{
				PersonId:      person.Id,
				OtherPersonId: "efg",
				Wins:          8,
				Losses:        4,
				Reporter:      false,
			},
			Game{
				PersonId:      person.Id,
				OtherPersonId: "hij",
				Wins:          2,
				Losses:        6,
				Reporter:      true,
			},
		}
		games[person.Id] = personGames
	}
	return games, nil
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

func _createId() string {
	return uuid.New().String()

}

func AdminDatabase() (res string, err error) {
	fmt.Println("INSIDE ADMINE DATABASE")
	fmt.Println(svc)
	tableName := ENVIRONMENT + "_person"
	id := _createId()
	fmt.Println("a1")
	item := Person_DANGEROUS{
		Id:     id,
		Score:  .5321,
		Name:   "Conrad",
		FaIcon: "fas fa-wave",
		Wins:   53,
		Losses: 14,
	}
	fmt.Println("a1")

	av, err := dynamodbattribute.MarshalMap(item)
	fmt.Println("a2")
	if err != nil {
		return "", err
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}
	fmt.Println("a3")

	_, err = svc.PutItem(input)
	fmt.Println("a4")
	if err != nil {
		return "", err
	}

	// return "Successfully added item to table " + tableName + " with id " + id, nil
	scanInput := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}
	fmt.Println("a5")
	result, err := svc.Scan(scanInput)
	fmt.Println("a6")
	if err != nil {
		return "", err
	}
	for _, i := range result.Items {
		fmt.Println("a7", i)
		person := Person_DANGEROUS{}

		err = dynamodbattribute.UnmarshalMap(i, &person)

		if err != nil {
			return "", err
		}

		fmt.Println("Name: ", person.Name)
		fmt.Println("Id:", person.Id)
		fmt.Println("Score:", person.Score)
		fmt.Println("AccessKey:", person.AccessKey)
		fmt.Println()
	}

	return "", nil

}
