package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func setup() {
	// Ensure the CSV file is initialized before each test
	err := initializeCSV()
	if err != nil {
		panic(err)
	}
}

func teardown() {
	// Cleanup after tests
	os.Remove(FILENAME)
}

func TestMessagingApp(t *testing.T) {
	setup()
	defer teardown()

	t.Run("List messages initially", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/messages?from=Alice&to=Bob", nil)
		rr := httptest.NewRecorder()
		r := mux.NewRouter()
		r.HandleFunc("/messages", getMessages).Methods("GET")
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body, _ := ioutil.ReadAll(rr.Body)
		var messages []Message
		json.Unmarshal(body, &messages)
		assert.Equal(t, 0, len(messages))
	})

	t.Run("Create messages", func(t *testing.T) {
		msg1 := Message{From: "Alice", To: "Bob", Subject: "Hello", Content: "Hi Bob!"}
		msg2 := Message{From: "Bob", To: "Alice", Subject: "Re: Hello", Content: "Hi Alice!"}

		// Create first message
		body, _ := json.Marshal(msg1)
		req, _ := http.NewRequest("POST", "/message", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		r := mux.NewRouter()
		r.HandleFunc("/message", createMessage).Methods("POST")
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusCreated, rr.Code)

		// Create second message
		body, _ = json.Marshal(msg2)
		req, _ = http.NewRequest("POST", "/message", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr = httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("List messages after creation", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/messages?from=Alice&to=Bob", nil)
		rr := httptest.NewRecorder()
		r := mux.NewRouter()
		r.HandleFunc("/messages", getMessages).Methods("GET")
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		body, _ := ioutil.ReadAll(rr.Body)
		var messages []Message
		json.Unmarshal(body, &messages)

		assert.Equal(t, 1, len(messages))
		assert.Equal(t, "Alice", messages[0].From)
		assert.Equal(t, "Bob", messages[0].To)
		assert.Equal(t, "Hello", messages[0].Subject)
		assert.Equal(t, "Hi Bob!", messages[0].Content)

		req, _ = http.NewRequest("GET", "/messages?from=Bob&to=Alice", nil)
		rr = httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		body, _ = ioutil.ReadAll(rr.Body)
		json.Unmarshal(body, &messages)

		assert.Equal(t, 1, len(messages))
		assert.Equal(t, "Bob", messages[0].From)
		assert.Equal(t, "Alice", messages[0].To)
		assert.Equal(t, "Re: Hello", messages[0].Subject)
		assert.Equal(t, "Hi Alice!", messages[0].Content)
	})
}
