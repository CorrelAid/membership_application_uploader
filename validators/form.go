package validators

import (
	"errors"

	"github.com/CorrelAid/membership_application_uploader/models"
)

const MaxFileSize = 3 * 1024 * 1024 // 3 MB

func ValidateFormData(formData models.FormData) error {
	if formData.File == nil {
		return errors.New("file field is required")
	}

	if formData.Name == "" || formData.Email == "" {
		return errors.New("name and Email fields are required")
	}

	if err := validateFile(formData.File); err != nil {
		return err
	}

	return nil
}

func validateFile(file []byte) error {
	if len(file) > MaxFileSize {
		return errors.New("file size exceeds the maximum limit")
	}

	return nil
}
