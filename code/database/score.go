package database

import (
	"github.com/aws/aws-sdk-go/aws"
	"gonum.org/v1/gonum/mat"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"math"

	"errors"
	"fmt"
)

func ComputeScores() (insertedId string, err error) {
	// TODO if you care, you should get a lock here to prevent race conditions on
	// person score and the score table in case there are more than 25 people
	// TODO we can compute the eigenvector using the power method
	// https://www.jstor.org/stable/2132526
	people, err := GetPeople()
	if err != nil {
		return insertedId, err
	}

	invalidPeople := make([]Person, 0)
	// People with at least 1 win
	validPeople := make([]Person, 0)
	peopleGamesMap := map[string]map[string][]Game{}
	for _, person := range people {
		gamesMap, err := GetGames([]Person{person}, -1)
		if err != nil {
			return insertedId, err
		}
		peopleGamesMap[person.Id] = gamesMap

		// Ensure that valid players have at least 9 games played against 3 unique opponents
		opponents := map[string]bool{}
		numGames := person.Wins + person.Losses
		valid := false
		if numGames >= 5 {
			for _, game := range gamesMap[person.Id] {
				if _, ok := opponents[game.OtherPersonId]; !ok {
					opponents[game.OtherPersonId] = true
				}
				if len(opponents) >= 3 {
					valid = true
					break
				}
			}
		}
		if valid {
			validPeople = append(validPeople, person)
		} else {
			invalidPeople = append(invalidPeople, person)
		}
	}

	scoresMap := make(map[string]float64)
	transactions := make([]*dynamodb.TransactWriteItem, 1+len(validPeople)+len(invalidPeople))
	if len(validPeople) > 0 {

		recordMatrix := make([][]float64, len(validPeople))
		// A map mapping person ID to index
		personIdMap := map[string]int{}

		for i, p := range validPeople {
			col := make([]float64, len(validPeople))
			for j := range col {
				col[j] = 0
			}
			recordMatrix[i] = col
			personIdMap[p.Id] = i
		}
		for _, p := range invalidPeople {
			personIdMap[p.Id] = -1
		}

		if err != nil {
			return insertedId, err
		}

		for _, person := range validPeople {
			p1Index, ok := personIdMap[person.Id]
			if !ok {
				continue
			}
			gamesMap := peopleGamesMap[person.Id]
			for _, game := range gamesMap[person.Id] {
				if game.PersonId != person.Id {
					return insertedId, errors.New("Looking at a game that we shouldn't be")

				}
				p2Index, ok := personIdMap[game.OtherPersonId]
				if ok && p2Index != -1 {
					// Fill up the record matrix with wins only for the current person
					recordMatrix[p1Index][p2Index] = recordMatrix[p1Index][p2Index] + float64(game.Wins) - float64(game.Losses)/2
				}
			}
			for _, otherPerson := range validPeople {
				p2Index, ok := personIdMap[otherPerson.Id]
				// Square root the totals to ensure # games isn't taken too significantly into account
				if ok {
					curValue := recordMatrix[p1Index][p2Index]
					if curValue <= 0 {
						recordMatrix[p1Index][p2Index] = 0.1
					} else {
						recordMatrix[p1Index][p2Index] = math.Sqrt(curValue)
					}
				}
			}
		}

		// Create a record matrix
		data := make([]float64, len(validPeople)*len(validPeople))
		for i, _ := range validPeople {
			for j, _ := range validPeople {
				index := i*len(validPeople) + j
				data[index] = recordMatrix[i][j]
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
			realValue := real(eigenvalues[i])
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
			transactions[1+i] = &dynamodb.TransactWriteItem{
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
		transactions[1+len(validPeople)+i] = &dynamodb.TransactWriteItem{
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

	// Create the object to pass into the score table
	created, createdObj := _getNow()
	yearMonth := fmt.Sprintf("%d-%d", createdObj.Year(), createdObj.Month())
	av, err := dynamodbattribute.MarshalMap(Score{
		Created:   created,
		ScoresMap: scoresMap,
		YearMonth: yearMonth,
	})
	if err != nil {
		return insertedId, err
	}
	transactions[0] = &dynamodb.TransactWriteItem{
		Put: &dynamodb.Put{
			TableName: aws.String(_getTableName("score")),
			Item:      av,
		},
	}

	err = _loopTransactions(transactions)
	if err != nil {
		return insertedId, err
	}

	return insertedId, nil
}
