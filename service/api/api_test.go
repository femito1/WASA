// File: api_test.go
package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/femito1/WASA/service/database"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/sirupsen/logrus"
)

// setupTestAPI creates an in-memory SQLite DB, initializes the database layer, and the API router.
func setupTestAPI(t *testing.T) http.Handler {
	t.Helper()

	// Open an in-memory SQLite DB.
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory SQLite DB: %v", err)
	}

	// Initialize the database layer.
	dbInstance, err := database.New(sqlDB)
	if err != nil {
		t.Fatalf("failed to initialize AppDatabase: %v", err)
	}

	// Create a logger.
	logger := logrus.New()

	// Create the API router.
	apiRouter, err := New(Config{
		Logger:   logger,
		Database: dbInstance,
	})
	if err != nil {
		t.Fatalf("failed to create API router: %v", err)
	}

	return apiRouter.Handler()
}

// TestDoLogin tests the /session endpoint.
func TestDoLogin(t *testing.T) {
	handler := setupTestAPI(t)

	// Create a POST request to /session.
	reqBody := []byte(`{"name": "testuser"}`)
	req := httptest.NewRequest("POST", "/session", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call the handler.
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	// Parse response as a generic map.
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if _, ok := resp["identifier"]; !ok {
		t.Fatalf("expected identifier in response, got: %v", resp)
	}
	if _, ok := resp["userId"]; !ok {
		t.Fatalf("expected userId in response, got: %v", resp)
	}
}

// TestListUsers tests GET /users endpoint.
func TestListUsers(t *testing.T) {
	handler := setupTestAPI(t)

	// First, create a user by calling /session.
	reqBody := []byte(`{"name": "testuser"}`)
	req := httptest.NewRequest("POST", "/session", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	// Now, send a GET request to /users.
	req = httptest.NewRequest("GET", "/users", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	var users []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &users); err != nil {
		t.Fatal(err)
	}
	if len(users) == 0 {
		t.Fatal("expected at least one user")
	}
}

// TestCreateConversationAndSendMessage tests conversation creation and message sending.
func TestCreateConversationAndSendMessage(t *testing.T) {
	handler := setupTestAPI(t)

	// Create a user.
	reqBody := []byte(`{"name": "user1"}`)
	req := httptest.NewRequest("POST", "/session", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, w.Code)
	}
	var loginResp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &loginResp); err != nil {
		t.Fatal(err)
	}
	uidFloat, ok := loginResp["userId"].(float64)
	if !ok {
		t.Fatal("expected userId as a number")
	}
	userID := strconv.FormatUint(uint64(uidFloat), 10)
	token, ok := loginResp["identifier"].(string)
	if !ok {
		t.Fatal("expected identifier as token")
	}

	// Create a conversation as user1.
	convURL := "/users/" + userID + "/conversations"
	convReqPayload := []byte(`{"name": "Test Conversation", "members": []}`)
	req = httptest.NewRequest("POST", convURL, bytes.NewBuffer(convReqPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusCreated, w.Code, w.Body.String())
	}
	var convResp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &convResp); err != nil {
		t.Fatal(err)
	}
	convIDFloat, ok := convResp["id"].(float64)
	if !ok {
		t.Fatal("expected conversation id in response")
	}
	convID := strconv.FormatUint(uint64(convIDFloat), 10)

	// Send a message in the conversation.
	msgURL := "/users/" + userID + "/conversations/" + convID + "/messages"
	msgReqPayload := []byte(`{"content": "Hello World", "format": "string"}`)
	req = httptest.NewRequest("POST", msgURL, bytes.NewBuffer(msgReqPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusCreated, w.Code, w.Body.String())
	}
	var msgResp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &msgResp); err != nil {
		t.Fatal(err)
	}
	if _, ok := msgResp["id"]; !ok {
		t.Fatal("expected message id in response")
	}
}
