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
	for _, p := range people {
		if p.Wins+p.Losses > 0 {
			validPeople = append(validPeople, p)
		} else {
			invalidPeople = append(invalidPeople, p)
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
