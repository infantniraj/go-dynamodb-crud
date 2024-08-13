package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"

	//"github.com/aws/aws-sdk-go-v2/aws/credentials"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type Item struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var svc *dynamodb.Client
var tableName string

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	env := os.Getenv("APP_ENV")
	tableName = os.Getenv("DYNAMODB_TABLE")

	var cfg aws.Config

	if env == "local" {
		// Local DynamoDB configuration
		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(os.Getenv("AWS_REGION")),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				os.Getenv("AWS_ACCESS_KEY_ID"),
				os.Getenv("AWS_SECRET_ACCESS_KEY"),
				"",
			)),
			config.WithEndpointResolver(aws.EndpointResolverFunc(
				func(service, region string) (aws.Endpoint, error) {
					if service == dynamodb.ServiceID && region == os.Getenv("AWS_REGION") {
						return aws.Endpoint{
							PartitionID:       "aws",
							URL:               os.Getenv("DYNAMODB_LOCAL_ENDPOINT"),
							SigningRegion:     os.Getenv("AWS_REGION"),
							HostnameImmutable: true,
						}, nil
					}
					return aws.Endpoint{}, &aws.EndpointNotFoundError{}
				},
			)),
		)
	} else {
		// Production DynamoDB configuration
		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(os.Getenv("AWS_REGION")),
		)
	}

	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	svc = dynamodb.NewFromConfig(cfg)

	// Create the table if it doesn't exist
	err = createTableIfNotExists()
	if err != nil {
		log.Fatalf("Error ensuring table exists: %v", err)
	}
}

func createItem(w http.ResponseWriter, r *http.Request) {
	var item Item
	_ = json.NewDecoder(r.Body).Decode(&item)

	input := &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item: map[string]types.AttributeValue{
			"ID":    &types.AttributeValueMemberS{Value: item.ID},
			"Name":  &types.AttributeValueMemberS{Value: item.Name},
			"Email": &types.AttributeValueMemberS{Value: item.Email},
		},
	}

	_, err := svc.PutItem(context.TODO(), input)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to add item: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode("Successfully added item")
}

func getItem(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	input := &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"ID": &types.AttributeValueMemberS{Value: id},
		},
	}

	result, err := svc.GetItem(context.TODO(), input)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get item: %v", err), http.StatusInternalServerError)
		return
	}

	if result.Item == nil {
		http.Error(w, "Could not find item", http.StatusNotFound)
		return
	}

	item := Item{
		ID:    result.Item["ID"].(*types.AttributeValueMemberS).Value,
		Name:  result.Item["Name"].(*types.AttributeValueMemberS).Value,
		Email: result.Item["Email"].(*types.AttributeValueMemberS).Value,
	}

	json.NewEncoder(w).Encode(item)
}

func updateItem(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	var item Item
	_ = json.NewDecoder(r.Body).Decode(&item)

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"ID": &types.AttributeValueMemberS{Value: id},
		},
		UpdateExpression: aws.String("set Name = :n, Email = :e"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":n": &types.AttributeValueMemberS{Value: item.Name},
			":e": &types.AttributeValueMemberS{Value: item.Email},
		},
		ReturnValues: types.ReturnValueUpdatedNew,
	}

	_, err := svc.UpdateItem(context.TODO(), input)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update item: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode("Successfully updated item")
}

func deleteItem(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"ID": &types.AttributeValueMemberS{Value: id},
		},
	}

	_, err := svc.DeleteItem(context.TODO(), input)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete item: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode("Successfully deleted item")
}

func createTableIfNotExists() error {
	// Describe the table to check if it exists
	_, err := svc.DescribeTable(context.TODO(), &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if ok := errors.As(err, &nfe); ok {
			// Table does not exist, create it
			_, err = svc.CreateTable(context.TODO(), &dynamodb.CreateTableInput{
				TableName: aws.String(tableName),
				AttributeDefinitions: []types.AttributeDefinition{
					{
						AttributeName: aws.String("ID"),
						AttributeType: types.ScalarAttributeTypeS,
					},
				},
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("ID"),
						KeyType:       types.KeyTypeHash, // Partition key
					},
				},
				ProvisionedThroughput: &types.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(5),
					WriteCapacityUnits: aws.Int64(5),
				},
			})
			if err != nil {
				return fmt.Errorf("failed to create table: %w", err)
			}
			log.Println("Table created successfully.")
		} else {
			// Some other error occurred
			return fmt.Errorf("failed to describe table: %w", err)
		}
	} else {
		log.Println("Table already exists.")
	}
	return nil
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/item", createItem).Methods("POST")
	r.HandleFunc("/item/{id}", getItem).Methods("GET")
	r.HandleFunc("/item/{id}", updateItem).Methods("PUT")
	r.HandleFunc("/item/{id}", deleteItem).Methods("DELETE")

	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
