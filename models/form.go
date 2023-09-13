package models

import (
	"mime/multipart"
)

type FormData struct {
	File  *multipart.FileHeader
	Name  string
	Email string
}

type ProcessedFormData struct {
	Name        string
	FileContent []byte
	Email       string
}
