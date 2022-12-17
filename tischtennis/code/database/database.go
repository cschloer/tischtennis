package database

import (
	"github.com/aws/aws-sdk-go/aws"
	"gonum.org/v1/gonum/mat"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/rs/xid"
	"math"
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
	MonthYear string
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
		games := make([]Game, len(result.Items))
		for i, item := range result.Items {
			game := Game{}

			err = dynamodbattribute.UnmarshalMap(item, &game)
			if err != nil {
				return gamesMap, err
			}
			games[i] = game
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
	created = time.Now().UnixNano()
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
func ComputeScores() (insertedId string, err error) {
	// TODO we can compute the eigenvector using the power method
	// https://www.jstor.org/stable/2132526
	people, err := GetPeople()
	if err != nil {
		return insertedId, err
	}

	invalidPeople := make([]Person, 0)
	// People with at least 1 win
	validPeople := make([]Person, 0)
	for _, p := range people {
		if p.Wins+p.Losses > 0 {
			validPeople = append(validPeople, p)
		} else {
			invalidPeople = append(invalidPeople, p)
		}
	}

	scoresMaps := make(map[string]float64)
	transactions := make([]*dynamodb.TransactWriteItem, len(validPeople)+len(invalidPeople))
	if len(validPeople) > 0 {

		recordMatrix := make([][]float64, len(validPeople))
		// A map mapping person ID to index
		personIdMap := map[string]int{}

		for i, p := range validPeople {
			col := make([]float64, len(validPeople))
			for j := range col {
				if j == i {
					// No games played against yourself
					col[j] = 0
				} else {
					// We artifically give everyone 0.1 win against eachother
					// to prevent issues with undefeated players forcing everyone to 0
					col[j] = 0.1
				}
			}
			recordMatrix[i] = col
			personIdMap[p.Id] = i
		}

		if err != nil {
			return insertedId, err
		}

		for _, person := range validPeople {
			gamesMap, err := GetGames([]Person{person}, -1)
			if err != nil {
				return insertedId, err
			}
			for _, game := range gamesMap[person.Id] {
				p1Index, ok1 := personIdMap[game.PersonId]
				p2Index, ok2 := personIdMap[game.OtherPersonId]
				if ok1 && ok2 {
					// Fill up the record matrix with wins only for the current person
					recordMatrix[p1Index][p2Index] = recordMatrix[p1Index][p2Index] + float64(game.Wins)
				}
			}
		}

		// Create a record matrix that is weighted by number of games played
		data := make([]float64, len(validPeople)*len(validPeople))
		for i, p := range validPeople {
			totalGames := p.Wins + p.Losses
			for j, _ := range validPeople {
				index := i*len(validPeople) + j
				// We divide by total games + len(validPeople) because we've artifically added 1 win
				// against everyone in order to prevent undefeated players from setting everyone to 0
				data[index] = recordMatrix[i][j] / float64(totalGames+int32(len(validPeople)))
			}
		}

		a := mat.NewDense(len(validPeople), len(validPeople), data)
		var eig mat.Eigen
		ok := eig.Factorize(a, mat.EigenRight)
		if !ok {
			return insertedId, errors.New("Eigendecomposition failed")
		}
		eigenvalues := eig.Values(nil)
		eigenvectors := mat.NewCDense(len(validPeople), len(validPeople), nil)
		eig.VectorsTo(eigenvectors)

		largestEigenvalue := real(eigenvalues[0])
		largestEigenvalueIndex := 0
		for i, _ := range validPeople {
			realValue := real(eigenvalues[1])
			if realValue > largestEigenvalue {
				largestEigenvalue = realValue
				largestEigenvalueIndex = i
			}
		}
		scores := make([]float64, len(validPeople))
		for j, _ := range validPeople {
			scores[j] = math.Abs(real(eigenvectors.At(j, largestEigenvalueIndex)))
		}

		for i, p := range validPeople {
			scoresMap[p.Id] = scores[i]
			transactions[i] = &dynamodb.TransactWriteItem{
				Update: &dynamodb.Update{
					ExpressionAttributeNames: map[string]*string{
						"#Score": aws.String("Score"),
					},
					ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
						":score": {
							N: aws.String(fmt.Sprintf("%f", scores[i])),
						},
					},
					TableName: aws.String(_getTableName("person")),
					Key: map[string]*dynamodb.AttributeValue{
						"Id": &dynamodb.AttributeValue{
							S: aws.String(p.Id),
						},
					},
					UpdateExpression: aws.String(
						"SET #Score = :score",
					),
				},
			}
		}
		/*
			// PRINT states for future debugging
				fmt.Println("Valid people: ", validPeople)
				fmt.Println("Record matrix: ", recordMatrix)
				fmt.Println("Weighted data: ", data)
				fmt.Printf("Eigenvalues:\n%v\n", eigenvalues)
				fmt.Println("Eigenvectors: ")
				for i, _ := range validPeople {
					for j, _ := range validPeople {
						floatValue := real(eigenvectors.At(i, j))
						fmt.Printf("%f ", floatValue)
					}
					fmt.Printf("\n")
				}
				fmt.Println("Largest eigenvalue index: ", largestEigenvalueIndex)
				fmt.Println("Final scores: ", scores)
		*/

	}

	// Set invalid people scores to -1
	for i, p := range invalidPeople {
		scoresMap[p.Id] = -1.0
		transactions[1 + len(validPeople)+i] = &dynamodb.TransactWriteItem{
			Update: &dynamodb.Update{
				ExpressionAttributeNames: map[string]*string{
					"#Score": aws.String("Score"),
				},
				ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
					":score": {
						N: aws.String(fmt.Sprintf("%f", -1.0)),
					},
				},
				TableName: aws.String(_getTableName("person")),
				Key: map[string]*dynamodb.AttributeValue{
					"Id": &dynamodb.AttributeValue{
						S: aws.String(p.Id),
					},
				},
				UpdateExpression: aws.String(
					"SET #Score = :score",
				),
			},
		}
	}
		transactions[] = &dynamodb.TransactWriteItem{

		}
	err = _loopTransactions(transactions)
	if err != nil {
		return insertedId, err
	}

	return insertedId, nil
}

func AdminDatabase() (res string, err error) {
	person1, _ := CreatePerson("dev_Lucas", "fas fa-chess-knight", "123")
	person2, _ := CreatePerson("dev_Conrad", "fas fa-water", "123")
	person3, _ := CreatePerson("dev_Christian", "fas fa-cat", "123")
	_, _ = CreatePerson("dev_Ron", "", "123")

	CreateGame(person1, person3, 1, 1)
	CreateGame(person1, person3, 2, 2)
	CreateGame(person1, person3, 3, 3)
	CreateGame(person1, person3, 4, 4)
	CreateGame(person1, person3, 5, 5)
	CreateGame(person1, person3, 6, 6)
	CreateGame(person2, person3, 7, 7)
	CreateGame(person2, person3, 8, 8)
	CreateGame(person2, person3, 9, 9)
	CreateGame(person1, person3, 10, 10)
	CreateGame(person2, person3, 11, 11)

	_, err = ComputeScores()

	return person3, err

}
