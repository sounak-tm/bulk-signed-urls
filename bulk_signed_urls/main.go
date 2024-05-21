package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"bulk_signed_urls/handlers"
	"bulk_signed_urls/utils"
)

func main() {

	if err := utils.LoadEnv(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/generate-signed-urls", handlers.GenerateSignedURLsHandler).Methods("POST")

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
