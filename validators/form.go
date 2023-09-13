package validators

import (
	"errors"
	"io"
	"mime/multipart"

	"github.com/CorrelAid/membership_application_uploader/models"
)

func ValidateProcessFormData(formData models.FormData, max_size int) (models.ProcessedFormData, error) {
	if formData.File == nil {
		return models.ProcessedFormData{}, errors.New("file field is required")
	}

	if formData.Name == "" || formData.Email == "" {
		return models.ProcessedFormData{}, errors.New("name and Email fields are required")
	}
	data, err := validateProcessFile(formData.File, max_size)

	if err != nil {
		return models.ProcessedFormData{}, err
	}
	processedFormData := models.ProcessedFormData{
		Name:        formData.Name,
		FileContent: data,
		Email:       formData.Email,
	}
	return processedFormData, nil
}
func validateProcessFile(file *multipart.FileHeader, max_size int) ([]byte, error) {
	// Check if file size exceeds the maximum limit

	if file.Size > int64(max_size) {
		return nil, errors.New("file size exceeds the maximum limit")
	}

	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}

	return data, nil
}
