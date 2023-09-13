package operations

import (
	"log"
	"time"

	"github.com/CorrelAid/membership_application_uploader/inits"
	"github.com/CorrelAid/membership_application_uploader/models"
)

func InsertMember(processedFormData models.ProcessedFormData, currentTime string) error {
	newMember := &models.Member{
		Email:  processedFormData.Email,
		Name:   processedFormData.Name,
		Time:   currentTime,
		Expiry: time.Now().Add(24 * 14 * time.Hour).Format(time.RFC1123),
	}

	txn := inits.DB.Txn(true)
	defer txn.Abort()

	if err := txn.Insert("member", newMember); err != nil {
		return err
	}

	txn.Commit()

	log.Printf("Inserted member: email=%s", newMember.Email)

	return nil
}
