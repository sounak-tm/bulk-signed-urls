package services

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

var (
	client        *storage.Client
	bucketName    string
	googleAccessID string
	privateKey    string
)

func init() {
	var err error
	ctx := context.Background()
	client, err = storage.NewClient(ctx, option.WithCredentialsFile(os.Getenv("GOOGLE_CREDENTIALS_FILE_PATH")))
	if err != nil {
		panic(fmt.Sprintf("Failed to create client: %v", err))
	}

	bucketName = os.Getenv("BUCKET_NAME")
	googleAccessID = os.Getenv("GOOGLE_ACCESS_ID")
	privateKey = os.Getenv("PRIVATE_KEY")
}

func GenerateSignedURL(ctx context.Context, objectName string) (string, error) {
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
