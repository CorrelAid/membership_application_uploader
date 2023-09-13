package validators

import (
	"errors"
	"io"
	"mime/multipart"

	"github.com/CorrelAid/membership_application_uploader/models"
)

const MaxFileSize = 3 * 1024 * 1024 // 3 MB

func ValidateProcessFormData(formData models.FormData) (models.ProcessedFormData, error) {
	if formData.File == nil {
		return models.ProcessedFormData{}, errors.New("file field is required")
	}

	if formData.Name == "" || formData.Email == "" {
		return models.ProcessedFormData{}, errors.New("name and Email fields are required")
	}
	data, err := validateProcessFile(formData.File)

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

func validateProcessFile(file *multipart.FileHeader) ([]byte, error) {

	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}

	if len(data) > MaxFileSize {
		return nil, errors.New("file size exceeds the maximum limit")
	}

	return data, nil
}
