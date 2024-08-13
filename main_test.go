package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCreateItemHandler_Success(t *testing.T) {
	// Setup
	service := &DynamoDBService{
		svc: nil, // Mock DynamoDB service here if needed
	}
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/items", service.CreateItemHandler)

	item := Item{
		ID:    "123",
		Name:  "Test Item",
		Value: "Test Value",
	}
	jsonValue, _ := json.Marshal(item)
	req, _ := http.NewRequest("POST", "/items", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.Equal(t, "Item created successfully", response["message"])
}

func TestCreateItemHandler_BadRequest(t *testing.T) {
	// Setup
	service := &DynamoDBService{
		svc: nil, // Mock DynamoDB service here if needed
	}
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/items", service.CreateItemHandler)

	req, _ := http.NewRequest("POST", "/items", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.Equal(t, "Invalid request payload", response["error"])
}

func TestCreateItemHandler_InternalServerError(t *testing.T) {
	// Setup
	service := &DynamoDBService{
		svc: nil, // Mock DynamoDB service to return an error
	}
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/items", service.CreateItemHandler)

	item := Item{
		ID:    "123",
		Name:  "Test Item",
		Value: "Test Value",
	}
	jsonValue, _ := json.Marshal(item)
	req, _ := http.NewRequest("POST", "/items", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.Contains(t, response["error"], "Failed to put item")
}
