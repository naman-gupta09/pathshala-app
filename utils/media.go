package utils

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

// SaveUploadedFile saves the uploaded file to the specified folder and returns the file path.
func SaveUploadedFile(file *multipart.FileHeader, folder string) (string, error) {
	// Create folder if it doesn't exist
	if err := os.MkdirAll(folder, os.ModePerm); err != nil {
		return "", err
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Generate a unique filename
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
	fullPath := filepath.Join(folder, filename)

	// Create destination file
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Copy file contents
	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	return fullPath, nil
}

// delete file if exists
func DeleteFileIfExists(path string) {
	if path == "" {
		return
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		log.Printf("Failed to delete file %s: %v", path, err)
	}
}
