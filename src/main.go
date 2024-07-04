package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

type User struct {
	PK        string  `json:"pk" dynamodbav:"pk"`
	FirstName string  `json:"firstName" dynamodbav:"firstName"`
	LastName  string  `json:"lastName" dynamodbav:"lastName"`
	Age       int     `json:"age" dynamodbav:"age"`
	Weight    float64 `json:"weight" dynamodbav:"weight"`
	Smoker    bool    `json:"smoker" dynamodbav:"smoker"`
}

func handleRequest(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	cfg, _ := config.LoadDefaultConfig(ctx)
	ddbClient := dynamodb.NewFromConfig(cfg)

	ddbTableName := os.Getenv("DDB_TABLE_NAME")

	fmt.Printf("DDB Table Name: %s", ddbTableName)

	fmt.Println(event.RouteKey)

	switch event.RouteKey {
	case "GET /users":
		input := &dynamodb.ScanInput{TableName: &ddbTableName}

		result, _ := ddbClient.Scan(context.TODO(), input)

		users := []User{}

		attributevalue.UnmarshalListOfMaps(result.Items, &users)

		response, _ := json.Marshal(users)

		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusOK,
			Body:       string(response),
		}, nil

	case "POST /users":
		var user User

		fmt.Println("Body:")
		fmt.Println(event.Body)

		json.Unmarshal([]byte(event.Body), &user)

		user.PK = uuid.NewString()

		fmt.Println("User:")
		fmt.Println(user)

		av, _ := attributevalue.MarshalMap(user)

		fmt.Println("AV:")
		fmt.Println(av)

		input := &dynamodb.PutItemInput{TableName: &ddbTableName, Item: av}

		fmt.Println("Input:")
		fmt.Println(input)

		result, dbErr := ddbClient.PutItem(context.TODO(), input)

		fmt.Println("Result:")
		fmt.Println(result)

		fmt.Println("Result Error:")
		fmt.Println(dbErr)

		response, _ := json.Marshal(user)

		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusCreated,
			Body:       string(response),
		}, nil

	case "GET /users/{id}":
		key, _ := attributevalue.Marshal(event.PathParameters["id"])

		input := &dynamodb.GetItemInput{
			TableName: &ddbTableName,
			Key:       map[string]types.AttributeValue{"pk": key},
		}

		result, _ := ddbClient.GetItem(context.TODO(), input)

		user := User{}

		attributevalue.UnmarshalMap(result.Item, &user)

		response, _ := json.Marshal(user)

		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusOK,
			Body:       string(response),
		}, nil

	case "PUT /users/{id}":
		var user User

		fmt.Println("Body:")
		fmt.Println(event.Body)

		json.Unmarshal([]byte(event.Body), &user)

		user.PK = event.PathParameters["id"]

		av, _ := attributevalue.MarshalMap(user)

		input := &dynamodb.PutItemInput{TableName: &ddbTableName, Item: av}

		ddbClient.PutItem(context.TODO(), input)

		response, _ := json.Marshal(user)

		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusOK,
			Body:       string(response),
		}, nil

	case "DELETE /users/{id}":
		key, _ := attributevalue.Marshal(event.PathParameters["id"])

		input := &dynamodb.DeleteItemInput{
			TableName: &ddbTableName,
			Key:       map[string]types.AttributeValue{"pk": key},
		}

		ddbClient.DeleteItem(context.TODO(), input)

		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusOK,
			Body:       `{"message":"ID deleted successfully"}`,
		}, nil
	}

	return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusMethodNotAllowed}, nil
}

func main() {
	lambda.Start(handleRequest)
}
