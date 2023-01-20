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
	Score       float64
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

type Score struct {
	YearMonth string
	Created   int64
	ScoresMap map[string]float64
}

// Person with AccessKey
type Person_DANGEROUS struct {
	Id          string
	Score       float64
	Name        string
	FaIcon      string
	Wins        int32
	Losses      int32
	NumReported int32
	AccessKey   string
}

var ENVIRONMENT = os.Getenv("ENVIRONMENT")

func _getSession() (sess *session.Session) {
	if ENVIRONMENT == "local" {
		if os.Getenv("LOCAL_DDB_HOST") == "" {
			panic("LOCAL_DDB_HOST not set in local mode")
		}
		sess = session.Must(session.NewSessionWithOptions(session.Options{
			Config: *aws.NewConfig().
				WithRegion("us-east-1").
				WithCredentials(credentials.NewStaticCredentials("DEFAULT_ACCESS_KEY", "DEFAULT_SECRET", "")).
				WithEndpoint(os.Getenv("LOCAL_DDB_HOST")),
		}))
	} else {
		sess = session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
		}))

	}
	return sess
}

var sess = _getSession()

// Create DynamoDB client
var svc = dynamodb.New(sess)

func _createId() string {
	return xid.New().String()
}
func _getNow() (int64, time.Time) {
	timeObj := time.Now()
	return timeObj.UnixMicro(), timeObj
}

func _getTableName(table string) string {
	return "tischtennis_" + ENVIRONMENT + "_" + table
}

func _loopTransactions(transactions []*dynamodb.TransactWriteItem) (err error) {
	// Loop through transactions 25 at a time
	for i := 0; i < len(transactions)/25+1; i++ {
		startIndex := i * 25
		endIndex := (i + 1) * 25
		if endIndex > len(transactions) {
			endIndex = len(transactions)
		}
		twii := &dynamodb.TransactWriteItemsInput{
			TransactItems: transactions[startIndex:endIndex],
		}
		_, err = svc.TransactWriteItems(twii)
		if err != nil {
			return err
		}
	}
	return nil

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

func GetGames(people []Person, limit int) (gamesMap map[string][]Game, err error) {
	// TODO add limit and paginate
	// TODO ensure it's actually sorting by Created properly
	gamesMap = make(map[string][]Game)
	for _, person := range people {
		input := &dynamodb.QueryInput{
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":personId": {
					S: aws.String(person.Id),
				},
			},
			TableName:              aws.String(_getTableName("game")),
			KeyConditionExpression: aws.String("PersonId = :personId"),
			// Reverse order
			ScanIndexForward: aws.Bool(false),
		}

		result, err := svc.Query(input)
		if err != nil {
			return gamesMap, err
		}

		numGames := limit
		if numGames == -1 || len(result.Items) < numGames {
			numGames = len(result.Items)
		}
		games := make([]Game, numGames)
		counter := 0
		for i, item := range result.Items {
			game := Game{}

			err = dynamodbattribute.UnmarshalMap(item, &game)
			if err != nil {
				return gamesMap, err
			}
			games[i] = game
			counter = counter + 1
			if limit != -1 && counter >= limit {
				break

			}
		}
		gamesMap[person.Id] = games
	}
	return gamesMap, nil
}

func GetPersonAccessKey(personId string) (personAccessKey string, err error) {
	out, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(_getTableName("person")),
		Key: map[string]*dynamodb.AttributeValue{
			"Id": {S: aws.String(personId)},
		},
	})
	if out.Item == nil {
		return "", errors.New(fmt.Sprintf("No person found with id %d.", personId))
	}
	person := Person_DANGEROUS{}

	err = dynamodbattribute.UnmarshalMap(out.Item, &person)

	if err != nil {
		return "", err
	}
	return person.AccessKey, nil

}

func GetGame(personId string, created int64) (game Game, err error) {
	out, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(_getTableName("game")),
		Key: map[string]*dynamodb.AttributeValue{
			"PersonId": {S: aws.String(personId)},
			"Created":  {N: aws.String(fmt.Sprintf("%d", created))},
		},
	})
	if err != nil {
		return game, err
	}
	if out.Item == nil {
		return game, errors.New(fmt.Sprintf("No game found with personId '%s' and created '%d'.", personId, created))
	}
	err = dynamodbattribute.UnmarshalMap(out.Item, &game)

	if err != nil {
		return game, err
	}
	return game, nil

}

func CreatePerson(name string, faIcon string, accessKey string) (personId string, err error) {
	id := _createId()
	person := Person_DANGEROUS{
		Id:          id,
		Name:        name,
		FaIcon:      faIcon,
		Wins:        0,
		Losses:      0,
		NumReported: 0,
		Score:       -1.0,
		AccessKey:   accessKey,
	}
	av, err := dynamodbattribute.MarshalMap(person)
	if err != nil {
		return "", err
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(_getTableName("person")),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		return "", err
	}
	return id, nil

}

func CreateGame(reporterId string, otherPersonId string, wins int32, losses int32) (personId1 string, personId2 string, created int64, err error) {
	created, _ = _getNow()
	game1 := Game{
		PersonId:      reporterId,
		Created:       created,
		OtherPersonId: otherPersonId,
		Wins:          wins,
		Losses:        losses,
		Reporter:      true,
	}
	game2 := Game{
		PersonId:      otherPersonId,
		Created:       created,
		OtherPersonId: reporterId,
		Wins:          losses,
		Losses:        wins,
		Reporter:      false,
	}
	av1, err := dynamodbattribute.MarshalMap(game1)
	if err != nil {
		return "", "", -1, err
	}
	av2, err := dynamodbattribute.MarshalMap(game2)
	if err != nil {
		return "", "", -1, err
	}

	twii := &dynamodb.TransactWriteItemsInput{
		TransactItems: []*dynamodb.TransactWriteItem{
			// Add game for reporter
			&dynamodb.TransactWriteItem{
				Put: &dynamodb.Put{
					TableName: aws.String(_getTableName("game")),
					Item:      av1,
				},
			},
			// Add game for other person
			&dynamodb.TransactWriteItem{
				Put: &dynamodb.Put{
					TableName: aws.String(_getTableName("game")),
					Item:      av2,
				},
			},
			// Update reporter's wins and losses
			&dynamodb.TransactWriteItem{
				Update: &dynamodb.Update{
					ExpressionAttributeNames: map[string]*string{
						"#Wins":        aws.String("Wins"),
						"#Losses":      aws.String("Losses"),
						"#NumReported": aws.String("NumReported"),
					},
					ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
						":wins": {
							N: aws.String(fmt.Sprintf("%d", game1.Wins)),
						},
						":losses": {
							N: aws.String(fmt.Sprintf("%d", game1.Losses)),
						},
						":one": {
							N: aws.String(fmt.Sprintf("%d", 1)),
						},
					},
					TableName: aws.String(_getTableName("person")),
					Key: map[string]*dynamodb.AttributeValue{
						"Id": &dynamodb.AttributeValue{
							S: aws.String(game1.PersonId),
						},
					},
					UpdateExpression: aws.String(
						"SET #Wins = #Wins + :wins, #Losses = #Losses + :losses, #NumReported = #NumReported + :one",
					),
				},
			},
			// Update other person's wins and losses
			&dynamodb.TransactWriteItem{
				Update: &dynamodb.Update{
					ExpressionAttributeNames: map[string]*string{
						"#Wins":   aws.String("Wins"),
						"#Losses": aws.String("Losses"),
					},
					ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
						":wins": {
							N: aws.String(fmt.Sprintf("%d", game2.Wins)),
						},
						":losses": {
							N: aws.String(fmt.Sprintf("%d", game2.Losses)),
						},
					},
					TableName: aws.String(_getTableName("person")),
					Key: map[string]*dynamodb.AttributeValue{
						"Id": &dynamodb.AttributeValue{
							S: aws.String(game2.PersonId),
						},
					},
					UpdateExpression: aws.String(
						"SET #Wins = #Wins + :wins, #Losses = #Losses + :losses",
					),
				},
			},
		},
	}

	_, err = svc.TransactWriteItems(twii)
	if err != nil {
		return "", "", -1, err
	}

	_, err = ComputeScores()
	if err != nil {
		return "", "", -1, err
	}

	return reporterId, otherPersonId, created, nil
}

func DeletePerson(personId string) (deletedId string, err error) {
	gamesMap, err := GetGames([]Person{Person{Id: personId}}, -1)
	if err != nil {
		return "", err
	}
	// The deletion of the requested person
	transactions := make([]*dynamodb.TransactWriteItem, 1+len(gamesMap[personId])*2)
	transactions[0] = &dynamodb.TransactWriteItem{
		Delete: &dynamodb.Delete{
			TableName: aws.String(_getTableName("person")),
			Key: map[string]*dynamodb.AttributeValue{
				"Id": &dynamodb.AttributeValue{
					S: aws.String(personId),
				},
			},
		},
	}
	type OtherPersonChanges struct {
		Wins        int32
		Losses      int32
		NumReported int32
	}
	otherPeopleChanges := make(map[string]*OtherPersonChanges)
	for i, game := range gamesMap[personId] {
		if _, ok := otherPeopleChanges[game.OtherPersonId]; !ok {
			otherPersonChanges := OtherPersonChanges{
				Wins:        0,
				Losses:      0,
				NumReported: 0,
			}
			otherPeopleChanges[game.OtherPersonId] = &otherPersonChanges
		}
		// Delete the game of the requested person
		transactions[1+(i*2)] = &dynamodb.TransactWriteItem{
			Delete: &dynamodb.Delete{
				TableName: aws.String(_getTableName("game")),
				Key: map[string]*dynamodb.AttributeValue{
					"PersonId": &dynamodb.AttributeValue{
						S: aws.String(game.PersonId),
					},
					"Created": &dynamodb.AttributeValue{
						N: aws.String(fmt.Sprintf("%d", game.Created)),
					},
				},
			},
		}
		// Delete the associated game entry of the other person
		transactions[1+(i*2)+1] = &dynamodb.TransactWriteItem{
			Delete: &dynamodb.Delete{
				TableName: aws.String(_getTableName("game")),
				Key: map[string]*dynamodb.AttributeValue{
					"PersonId": &dynamodb.AttributeValue{
						S: aws.String(game.OtherPersonId),
					},
					"Created": &dynamodb.AttributeValue{
						N: aws.String(fmt.Sprintf("%d", game.Created)),
					},
				},
			},
		}

		otherPersonChanges := otherPeopleChanges[game.OtherPersonId]
		// Update other persons values accordingly
		otherPersonChanges.Wins = otherPersonChanges.Wins + game.Losses
		otherPersonChanges.Losses = otherPersonChanges.Losses + game.Wins
		if !game.Reporter {
			otherPersonChanges.NumReported = otherPersonChanges.NumReported + 1
		}
	}
	// Now update all of the other person objects
	for otherPersonId, otherPersonChanges := range otherPeopleChanges {
		transactions = append(
			transactions,
			&dynamodb.TransactWriteItem{
				Update: &dynamodb.Update{
					ExpressionAttributeNames: map[string]*string{
						"#Wins":        aws.String("Wins"),
						"#Losses":      aws.String("Losses"),
						"#NumReported": aws.String("NumReported"),
					},
					ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
						":wins": {
							N: aws.String(fmt.Sprintf("%d", otherPersonChanges.Wins)),
						},
						":losses": {
							N: aws.String(fmt.Sprintf("%d", otherPersonChanges.Losses)),
						},
						":one": {
							N: aws.String(fmt.Sprintf("%d", otherPersonChanges.NumReported)),
						},
					},
					TableName: aws.String(_getTableName("person")),
					Key: map[string]*dynamodb.AttributeValue{
						"Id": &dynamodb.AttributeValue{
							S: aws.String(otherPersonId),
						},
					},
					UpdateExpression: aws.String(
						"SET #Wins = #Wins - :wins, #Losses = #Losses - :losses, #NumReported = #NumReported - :one",
					),
				},
			},
		)

	}
	err = _loopTransactions(transactions)
	if err != nil {
		return "", err
	}

	_, err = ComputeScores()
	if err != nil {
		return "", err
	}

	return personId, nil
}

func DeleteGame(personId string, created int64) (personId1 string, personId2 string, createdReturn int64, err error) {
	game, err := GetGame(personId, created)
	if err != nil {
		return "", "", -1, err
	}
	requesterReported := "1"
	otherPersonReported := "0"
	if !game.Reporter {
		requesterReported = "0"
		otherPersonReported = "1"
	}

	twii := &dynamodb.TransactWriteItemsInput{
		TransactItems: []*dynamodb.TransactWriteItem{
			// Delete requester game
			&dynamodb.TransactWriteItem{
				Delete: &dynamodb.Delete{
					TableName: aws.String(_getTableName("game")),
					Key: map[string]*dynamodb.AttributeValue{
						"PersonId": &dynamodb.AttributeValue{
							S: aws.String(personId),
						},
						"Created": &dynamodb.AttributeValue{
							N: aws.String(fmt.Sprintf("%d", created)),
						},
					},
				},
			},
			// Delete other person game
			&dynamodb.TransactWriteItem{
				Delete: &dynamodb.Delete{
					TableName: aws.String(_getTableName("game")),
					Key: map[string]*dynamodb.AttributeValue{
						"PersonId": &dynamodb.AttributeValue{
							S: aws.String(game.OtherPersonId),
						},
						"Created": &dynamodb.AttributeValue{
							N: aws.String(fmt.Sprintf("%d", created)),
						},
					},
				},
			},
			// Update requester's wins and losses
			&dynamodb.TransactWriteItem{
				Update: &dynamodb.Update{
					ExpressionAttributeNames: map[string]*string{
						"#Wins":        aws.String("Wins"),
						"#Losses":      aws.String("Losses"),
						"#NumReported": aws.String("NumReported"),
					},
					ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
						":wins": {
							N: aws.String(fmt.Sprintf("%d", game.Wins)),
						},
						":losses": {
							N: aws.String(fmt.Sprintf("%d", game.Losses)),
						},
						":one": {
							N: aws.String(requesterReported),
						},
					},
					TableName: aws.String(_getTableName("person")),
					Key: map[string]*dynamodb.AttributeValue{
						"Id": &dynamodb.AttributeValue{
							S: aws.String(personId),
						},
					},
					UpdateExpression: aws.String(
						"SET #Wins = #Wins - :wins, #Losses = #Losses - :losses, #NumReported = #NumReported - :one",
					),
				},
			},
			// Update other person's wins and losses
			&dynamodb.TransactWriteItem{
				Update: &dynamodb.Update{
					ExpressionAttributeNames: map[string]*string{
						"#Wins":        aws.String("Wins"),
						"#Losses":      aws.String("Losses"),
						"#NumReported": aws.String("NumReported"),
					},
					ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
						":wins": {
							N: aws.String(fmt.Sprintf("%d", game.Losses)),
						},
						":losses": {
							N: aws.String(fmt.Sprintf("%d", game.Wins)),
						},
						":one": {
							N: aws.String(otherPersonReported),
						},
					},
					TableName: aws.String(_getTableName("person")),
					Key: map[string]*dynamodb.AttributeValue{
						"Id": &dynamodb.AttributeValue{
							S: aws.String(game.OtherPersonId),
						},
					},
					UpdateExpression: aws.String(
						"SET #Wins = #Wins - :wins, #Losses = #Losses - :losses, #NumReported = #NumReported - :one",
					),
				},
			},
		},
	}

	_, err = svc.TransactWriteItems(twii)
	if err != nil {
		return "", "", -1, err
	}

	_, err = ComputeScores()
	if err != nil {
		return "", "", -1, err
	}

	return personId, game.OtherPersonId, created, nil
}
func AdminDatabase() (res string, err error) {
	person1, _ := CreatePerson("David", "fas fa-chess-knight", "123")
	person2, _ := CreatePerson("Conrad", "fas fa-water", "123")
	person3, _ := CreatePerson("Soenke", "fas fa-cat", "123")
	person4, _ := CreatePerson("Leo", "", "123")
	person5, _ := CreatePerson("Ron", "", "123")
	person6, _ := CreatePerson("Phil", "", "123")
	person7, _ := CreatePerson("Vlad", "", "123")
	_, _ = CreatePerson("Lucas", "", "123")

	CreateGame(person1, person2, 4, 6)
	CreateGame(person2, person5, 0, 3)
	CreateGame(person2, person3, 2, 0)
	CreateGame(person5, person3, 2, 0)
	CreateGame(person2, person6, 2, 0)
	CreateGame(person2, person6, 3, 0)
	CreateGame(person5, person4, 3, 0)
	CreateGame(person5, person3, 2, 0)
	CreateGame(person5, person6, 2, 0)
	CreateGame(person5, person6, 1, 0)
	CreateGame(person5, person3, 2, 0)
	CreateGame(person5, person7, 2, 0)
	CreateGame(person6, person3, 5, 0)
	CreateGame(person7, person6, 1, 1)
	CreateGame(person7, person2, 1, 0)
	CreateGame(person7, person3, 0, 13)

	return person3, err

}
