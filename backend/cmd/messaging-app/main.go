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
    FromName  string `json:"from_name"`
    ToName    string `json:"to_name"`
    Content   string `json:"content"`
    Timestamp string `json:"timestamp"`
}

type Conversation struct {
    ConversationID string   `json:"conversation_id"`
    Participants   []string `json:"participants"` // This can now be names
    LastMessage    Message  `json:"last_message"`
}

type User struct {
    Name     string `json:"name"`
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
	header := []string{"id", "from_name", "to_name", "content", "timestamp"}
	if err := writer.Write(header); err != nil {
		return err
	}

	return nil
}

func loginUser(w http.ResponseWriter, r *http.Request) {
    var user User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    // Logic to check if user exists or create a new one
    // For simplicity, this could be an in-memory store or a CSV file

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
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

	// Record now uses names instead of IDs
	record := []string{msg.ID, msg.FromName, msg.ToName, msg.Content, msg.Timestamp}
	if err := writer.Write(record); err != nil {
		http.Error(w, "Failed to write to CSV file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "Message created successfully")
}

func createConversation(w http.ResponseWriter, r *http.Request) {
	var conversation Conversation
	if err := json.NewDecoder(r.Body).Decode(&conversation); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Generate a unique conversation ID based on names
	conversation.ConversationID = conversation.Participants[0] + "-" + conversation.Participants[1]

	// Initialize an empty message as the last message since no message has been sent yet
	conversation.LastMessage = Message{}

	// Save the conversation to storage (append to file)
	file, err := os.OpenFile(FILENAME, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		http.Error(w, "Unable to open CSV file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write conversation with no message yet
	record := []string{conversation.ConversationID, conversation.Participants[0], conversation.Participants[1], "", ""}
	if err := writer.Write(record); err != nil {
		http.Error(w, "Failed to write to CSV file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "Conversation created successfully")
}

func getConversations(w http.ResponseWriter, r *http.Request) {
	userName := r.URL.Query().Get("name")

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
		if record[1] == userName || record[2] == userName {
			convID := record[1] + "-" + record[2]
			var lastMessage Message
			if record[3] == "" && record[4] == "" {
				lastMessage = Message{}
			} else {
				lastMessage = Message{
					ID:        record[0],
					FromName:  record[1],
					ToName:    record[2],
					Content:   record[3],
					Timestamp: record[4],
				}
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
				FromName:  record[1],
				ToName:    record[2],
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
	// Assuming `userID` should be used to identify the user; for now, just demonstrate the use of existing fields.
	user := User{
		Name:   "Sample User", // Replace with dynamic data based on `userID` if needed
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
	
	// Route to create a new conversation
	r.HandleFunc("/api/conversations", createConversation).Methods("POST")

	// Route to get user details
	r.HandleFunc("/api/users/{user_id}", getUserDetails).Methods("GET")
	
	r.HandleFunc("/api/login", loginUser).Methods("POST")

	log.Println("Server starting on :8080")
	if err := initializeCSV(); err != nil {
		log.Fatalf("Failed to initialize CSV file: %v", err)
	}
	log.Println("CSV file initialized successfully")

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("could not start server: %s", err)
	}
}
