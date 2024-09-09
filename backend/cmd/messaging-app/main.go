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
	ID        string `json:"id"`
	From      string `json:"from"`
	To        string `json:"to"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

type Conversation struct {
	ConversationID string   `json:"conversation_id"`
	Participants   []string `json:"participants"`
	LastMessage    Message  `json:"last_message"`
}

type User struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Status   string `json:"status"`
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
	header := []string{"id", "from", "to", "content", "timestamp"}
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

	record := []string{msg.ID, msg.From, msg.To, msg.Content, msg.Timestamp}
	if err := writer.Write(record); err != nil {
		http.Error(w, "Failed to write to CSV file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "Message created successfully")
}

func getConversations(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")

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

	conversations := make(map[string]Conversation)
	for _, record := range records[1:] { // Skip header
		if record[1] == userID || record[2] == userID {
			convID := record[1] + "-" + record[2]
			lastMessage := Message{
				ID:        record[0],
				From:      record[1],
				To:        record[2],
				Content:   record[3],
				Timestamp: record[4],
			}

			conversation, exists := conversations[convID]
			if !exists {
				conversation = Conversation{
					ConversationID: convID,
					Participants:   []string{record[1], record[2]},
					LastMessage:    lastMessage,
				}
			}
			conversations[convID] = conversation
		}
	}

	var convList []Conversation
	for _, conv := range conversations {
		convList = append(convList, conv)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(convList)
}

func getMessages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	conversationID := vars["conversation_id"]

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
		convID := record[1] + "-" + record[2]
		if convID == conversationID {
			msg := Message{
				ID:        record[0],
				From:      record[1],
				To:        record[2],
				Content:   record[3],
				Timestamp: record[4],
			}
			messages = append(messages, msg)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func getUserDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	// Sample static user data, replace this with actual database calls.
	user := User{
		UserID:   userID,
		Username: "Sample User",
		Avatar:   "/path/to/avatar.jpg",
		Status:   "Active 2h ago",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
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
	r.HandleFunc("/api/conversations/{conversation_id}/messages", createMessage).Methods("POST")

	// Route to retrieve messages in a specific conversation
	r.HandleFunc("/api/conversations/{conversation_id}/messages", getMessages).Methods("GET")

	// Route to retrieve all conversations for a user
	r.HandleFunc("/api/conversations", getConversations).Methods("GET")

	// Route to get user details
	r.HandleFunc("/api/users/{user_id}", getUserDetails).Methods("GET")

	log.Println("Server starting on :8080")
	if err := initializeCSV(); err != nil {
		log.Fatalf("Failed to initialize CSV file: %v", err)
	}
	log.Println("CSV file initialized successfully")

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("could not start server: %s", err)
	}
}
