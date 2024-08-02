package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)


func initializeCSV(filename string) error {
    // Create or truncate the file
    file, err := os.Create(filename)
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

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, Messaging App!")
	}).Methods("GET")

	// Additional routes can be added here

	log.Println("Server starting on :8080")
	filename := "messages.csv"
    if err := initializeCSV(filename); err != nil {
        log.Fatalf("Failed to initialize CSV file: %v", err)
    }
    log.Println("CSV file initialized successfully")

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("could not start server: %s", err)
	}
}
