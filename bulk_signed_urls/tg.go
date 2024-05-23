package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

// models

type FileInfo struct {
	ID   string
	Name string
}

type RequestPayload struct {
	Files map[string]string `json:"files"`
}

type ResponsePayload struct {
	URLs map[string]string `json:"urls"`
}

// variables

var (
	bucketName     string
	googleAccessID string
	privateKey     string
	client         *storage.Client
)

func init() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Read environment variables
	credentialsFilePath := os.Getenv("GOOGLE_CREDENTIALS_FILE_PATH")
	bucketName = os.Getenv("BUCKET_NAME")
	googleAccessID = os.Getenv("GOOGLE_ACCESS_ID")
	privateKey = os.Getenv("PRIVATE_KEY")

	// Initialize GCS Client
	ctx := context.Background()
	client, err = storage.NewClient(ctx, option.WithCredentialsFile(credentialsFilePath))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
}

func GenerateSignedURLs(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	files := make([]FileInfo, 0, len(requestPayload.Files))
	for id, name := range requestPayload.Files {
		files = append(files, FileInfo{ID: id, Name: name})
	}

	ctx := context.Background()
	results := make(chan FileInfo, len(files))
	done := make(chan struct{})

	for _, file := range files {
		go func(file FileInfo) {
			url, err := getSignedUrl(ctx, client, bucketName, googleAccessID, privateKey, file.Name)
			if err != nil {
				log.Printf("Failed to generate signed URL for %s: %v", file.Name, err)
				results <- FileInfo{ID: file.ID, Name: fmt.Sprintf("Error: %v", err)}
				return
			}
			results <- FileInfo{ID: file.ID, Name: url}
		}(file)
	}

	go func() {
		for range files {
			<-results
		}
		close(done)
	}()

	responsePayload := ResponsePayload{URLs: make(map[string]string)}
	for result := range results {
		responsePayload.URLs[result.ID] = result.Name
	}

	<-done

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(responsePayload); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func getSignedUrl(ctx context.Context, client *storage.Client, bucketName, googleAccessID, privateKey, objectName string) (string, error) {
	url, err := storage.SignedURL(bucketName, objectName, &storage.SignedURLOptions{
		GoogleAccessID: googleAccessID,
		PrivateKey:     []byte(strings.ReplaceAll(privateKey, "\\n", "\n")),
		Method:         "GET",
		Expires:        time.Now().Add(15 * time.Minute),
	})
	if err != nil {
		return "", fmt.Errorf("unable to generate signed URL: %v", err)
	}
	return url, nil
}
