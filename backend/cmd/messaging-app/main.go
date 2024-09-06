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
	Content string `json:"content"`
}

const FILENAME = "messages.csv"

func enableCORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Set CORS headers for the main request
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        // Handle preflight requests
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent) // 204 No Content response for preflight
            return
        }

        // Pass to the next handler
        next.ServeHTTP(w, r)
    })
}


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
    header := []string{"from", "to", "content"}
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

	record := []string{msg.From, msg.To, msg.Content}
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
		if (record[0] == from && record[1] == to) || (record[0] == to && record[1] == from) {
			msg := Message{
				From:    record[0],
				To:      record[1],
				Content: record[2],
			}
			messages = append(messages, msg)
		}
		
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}


func main() {
    r := mux.NewRouter()

    // Apply CORS middleware globally
    r.Use(enableCORS)

	// Handle preflight requests for all routes
    r.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        w.WriteHeader(http.StatusOK)
    })

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


