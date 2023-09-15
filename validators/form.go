package validators

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"

	emailverifier "github.com/AfterShip/email-verifier"
	"github.com/CorrelAid/membership_application_uploader/models"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

var (
	verifier = emailverifier.NewVerifier()
)

func ValidateProcessFormData(formData models.FormData, max_size int) (models.ProcessedFormData, error) {
	if formData.File == nil {
		return models.ProcessedFormData{}, errors.New("file field is required")
	}

	if formData.Name == "" || formData.Email == "" {
		return models.ProcessedFormData{}, errors.New("name and Email fields are required")
	}

	ret, err := verifier.Verify(formData.Email)
	if err != nil {
		return models.ProcessedFormData{}, err
	}
	if !ret.Syntax.Valid {
		return models.ProcessedFormData{}, errors.New("email address syntax is invalid")
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

	if file.Size > int64(max_size) {
		return nil, errors.New("file size exceeds the maximum limit")
	}

	if file.Header.Get("Content-Type") != "application/pdf" {
		return nil, errors.New("file is not a PDF")
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

	err = api.Validate(bytes.NewReader(data), nil)
	if err != nil {
		return nil, errors.New("PDF validation failed")
	}

	return data, nil
}
