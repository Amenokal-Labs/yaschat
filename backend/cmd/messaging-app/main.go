package main

import (
	"encoding/csv"
    "encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

type Message struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Content string `json:"content"`
}

const FILENAME = "messages.csv"


func initializeCSV() error {
    // Create or truncate the file
    file, err := os.Create(FILENAME)
    if err != nil {
        return err
    }
    defer file.Close()

    // Create a new CSV writer
    writer := csv.NewWriter(file)
    defer writer.Flush()

    // Write the header row
    header := []string{"from", "to", "subject", "content"}
    if err := writer.Write(header); err != nil {
        return err
    }

    return nil
}

func createMessage(w http.ResponseWriter, r *http.Request) {
	var msg Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	file, err := os.OpenFile(FILENAME, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		http.Error(w, "Unable to open CSV file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	record := []string{msg.From, msg.To, msg.Subject, msg.Content}
	if err := writer.Write(record); err != nil {
		http.Error(w, "Failed to write to CSV file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "Message created successfully")
}

func getMessages(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	file, err := os.Open(FILENAME)
	if err != nil {
		http.Error(w, "Unable to open CSV file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		http.Error(w, "Failed to read CSV file", http.StatusInternalServerError)
		return
	}

	var messages []Message
	for _, record := range records[1:] { // Skip header
		if record[0] == from && record[1] == to {
			msg := Message{
				From:    record[0],
				To:      record[1],
				Subject: record[2],
				Content: record[3],
			}
			messages = append(messages, msg)
		}
	}

	if len(messages) == 0 {
		http.Error(w, "No messages found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}


func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, Messaging App!")
	}).Methods("GET")

	// Route to create a message
	r.HandleFunc("/message", createMessage).Methods("POST")

	// Route to retrieve messages by "from" and "to"
	r.HandleFunc("/messages", getMessages).Methods("GET")

	log.Println("Server starting on :8080")
	if err := initializeCSV(); err != nil {
		log.Fatalf("Failed to initialize CSV file: %v", err)
	}
	log.Println("CSV file initialized successfully")

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("could not start server: %s", err)
	}
}
