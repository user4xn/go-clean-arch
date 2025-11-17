package helper

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

// SaveUploadedFile saves an uploaded file to the server with a unique filename.
func SaveUploadedFile(file *multipart.FileHeader) (string, error) {
	// Create a unique file name with a timestamp
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))

	// Define the path to save the file in the assets/uploads directory
	filePath := filepath.Join("assets", "uploads", filename)

	// Ensure the uploads directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return "", err
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Create the destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Copy the file contents
	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	return filePath, nil
}
