package app

import (
	"errors"
	"fmt"
	"strings"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// Function that takes a file upload from a multipart form and a destination folder, can be used for either avatars or posts
func UploadImage(file multipart.File, header multipart.FileHeader,dest string) (string, error) {

	defer file.Close()

	// Validate file size (e.g., 5MB limit)
	const maxFileSize = 5 * 1024 * 1024
	if header.Size > maxFileSize {

		return "", errors.New("file is too large (max 5MB)")
	}

	// TODO: Add MIME type validation to ensure it's an image.
	buffer := make([]byte,512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "",err
	}
	contentType := http.DetectContentType(buffer[:n])

	if !strings.HasPrefix(contentType,"image/") {
		return "",errors.New("invalid mime/type")
	}

	// Rewind the file pointer back to the start
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}

	// Save the file
	uniqueFilename := fmt.Sprintf("%s%s", uuid.New().String(), filepath.Ext(header.Filename))
	dst, err := os.Create(filepath.Join("./ui/templates/static/" + dest, uniqueFilename))
	if err != nil {
		return "", err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		return "", err
	}

	return uniqueFilename, nil
}
