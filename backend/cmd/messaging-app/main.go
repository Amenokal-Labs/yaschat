package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"

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
	Name string `json:"name"`
}

const FILENAME = "messages.csv"

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func initializeCSV() error {
	file, err := os.Create(FILENAME)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"id", "from_name", "to_name", "content", "timestamp"}
	if err := writer.Write(header); err != nil {
		return err
	}

	return nil
}

func loginUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	log.Printf("User logged in: %+v", user)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func createMessage(w http.ResponseWriter, r *http.Request) {
	var msg Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	log.Printf("Received message: %+v", msg)

	file, err := os.OpenFile(FILENAME, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Printf("Error opening CSV file: %v", err)
		http.Error(w, `{"error": "Unable to open CSV file"}`, http.StatusInternalServerError)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	record := []string{msg.ID, msg.FromName, msg.ToName, msg.Content, msg.Timestamp}
	if err := writer.Write(record); err != nil {
		log.Printf("Error writing to CSV file: %v", err)
		http.Error(w, `{"error": "Failed to write to CSV file"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, `{"message": "Message created successfully"}`)
}

func createConversation(w http.ResponseWriter, r *http.Request) {
	var conversation Conversation
	if err := json.NewDecoder(r.Body).Decode(&conversation); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	sort.Strings(conversation.Participants)
	log.Printf("Sorted participants: %+v", conversation.Participants)

	conversation.ConversationID = conversation.Participants[0] + "-" + conversation.Participants[1]
	conversation.LastMessage = Message{}

	file, err := os.OpenFile(FILENAME, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Printf("Error opening CSV file: %v", err)
		http.Error(w, `{"error": "Unable to open CSV file"}`, http.StatusInternalServerError)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	record := []string{conversation.ConversationID, conversation.Participants[0], conversation.Participants[1], "", ""}
	if err := writer.Write(record); err != nil {
		log.Printf("Error writing to CSV file: %v", err)
		http.Error(w, `{"error": "Failed to write to CSV file"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, `{"message": "Conversation created successfully"}`)
}

func getConversations(w http.ResponseWriter, r *http.Request) {
	userName := r.URL.Query().Get("name")
	log.Printf("Fetching conversations for user: %s", userName)

	file, err := os.Open(FILENAME)
	if err != nil {
		log.Printf("Error opening CSV file: %v", err)
		http.Error(w, `{"error": "Unable to open CSV file"}`, http.StatusInternalServerError)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Printf("Error reading CSV file: %v", err)
		http.Error(w, `{"error": "Failed to read CSV file"}`, http.StatusInternalServerError)
		return
	}

	conversations := make(map[string]Conversation)
	for _, record := range records[1:] {
		if record[1] == userName || record[2] == userName {
			participants := []string{record[1], record[2]}
			sort.Strings(participants)
			convID := participants[0] + "-" + participants[1]
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

	log.Printf("Found conversations: %+v", convList)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(convList)
}

func getMessages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	conversationID := vars["conversation_id"]
	log.Printf("Fetching messages for conversation ID: %s", conversationID)

	file, err := os.Open(FILENAME)
	if err != nil {
		log.Printf("Error opening CSV file: %v", err)
		http.Error(w, `{"error": "Unable to open CSV file"}`, http.StatusInternalServerError)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Printf("Error reading CSV file: %v", err)
		http.Error(w, `{"error": "Failed to read CSV file"}`, http.StatusInternalServerError)
		return
	}

	var messages []Message
	for _, record := range records[1:] {
		participants := []string{record[1], record[2]}
		sort.Strings(participants)
		convID := participants[0] + "-" + participants[1]

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

	log.Printf("Found messages: %+v", messages)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func getUserDetails(w http.ResponseWriter, r *http.Request) {
	user := User{
		Name: "Sample User", // Replace with dynamic data based on `userID` if needed
	}

	log.Printf("Fetching user details: %+v", user)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func main() {
	r := mux.NewRouter()

	r.Use(enableCORS)

	r.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
	})

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, Messaging App!")
	}).Methods("GET")

	r.HandleFunc("/api/conversations/{conversation_id}/messages", createMessage).Methods("POST")
	r.HandleFunc("/api/conversations/{conversation_id}/messages", getMessages).Methods("GET")
	r.HandleFunc("/api/conversations", getConversations).Methods("GET")
	r.HandleFunc("/api/conversations", createConversation).Methods("POST")
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
