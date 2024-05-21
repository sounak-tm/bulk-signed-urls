package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"bulk_signed_urls/models"
	"bulk_signed_urls/services"
)

func GenerateSignedURLsHandler(w http.ResponseWriter, r *http.Request) {
	var requestPayload models.RequestPayload
	
	if err := json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	files := make([]models.FileInfo, 0, len(requestPayload.Files))
	for id, name := range requestPayload.Files {
		files = append(files, models.FileInfo{ID: id, Name: name})
	}

	ctx := context.Background()
	results := make(chan models.FileInfo, len(files))
	var wg sync.WaitGroup

	for _, file := range files {
		wg.Add(1)
		go func(file models.FileInfo) {
			defer wg.Done()
			url, err := services.GenerateSignedURL(ctx, file.Name)
			if err != nil {
				results <- models.FileInfo{ID: file.ID, Name: "Error: " + err.Error()}
				return
			}
			results <- models.FileInfo{ID: file.ID, Name: url}
		}(file)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	responsePayload := models.ResponsePayload{URLs: make(map[string]string)}
	for result := range results {
		responsePayload.URLs[result.ID] = result.Name
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(responsePayload); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
