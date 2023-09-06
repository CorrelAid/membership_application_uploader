package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/CorrelAid/membership_application_uploader/inits"
	"github.com/CorrelAid/membership_application_uploader/models"
	"github.com/CorrelAid/membership_application_uploader/validators"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-memdb"
	"github.com/joho/godotenv"
)

var DB *memdb.MemDB

func main() {
	err := godotenv.Load(".env")
	DB = inits.DBInit()

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	router := gin.Default()
	// Set a lower memory limit for multipart forms (default is 32 MiB)
	router.MaxMultipartMemory = 8 << 20 // 8 MiB
	router.ForwardedByClientIP = true
	router.SetTrustedProxies([]string{"127.0.0.1", "correlaid.org"})
	router.POST("/upload", uploadPDF)
	router.Run(":8080")

}

func uploadPDF(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, "Error getting file: "+err.Error())
		return
	}

	src, err := file.Open()
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err.Error()))
		return
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err.Error()))
		return
	}

	formData := models.FormData{
		File:  data,
		Name:  c.PostForm("name"),
		Email: c.PostForm("email"),
	}

	if err := validators.ValidateFormData(formData); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	currentTime := time.Now().Format(time.RFC1123)

	// Lookup by email
	txn := DB.Txn(false)
	defer txn.Abort()

	raw, err := txn.First("member", "id", formData.Email)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err.Error()))
		return
	}

	if raw != nil {
		c.String(http.StatusBadRequest, "Email already exists")
		return
	}

	if err := uploadFileToNextcloud(formData, currentTime); err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err.Error()))
		return
	}

	c.String(http.StatusOK, "File uploaded successfully")
}

func uploadFileToNextcloud(formData models.FormData, currentTime string) error {
	client := &http.Client{}

	filename := fmt.Sprintf("%s_%s", processName(formData.Name), currentTime)

	req, err := http.NewRequest(http.MethodPut, "https://cloud.correlaid.org/remote.php/dav/files/bot@correlaid.org/Mitgliedsanträge/"+filename+".pdf", bytes.NewReader(formData.File))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/octet-stream")

	pw := os.Getenv("NEXTCLOUD_PW")
	req.SetBasicAuth("bot@correlaid.org", pw)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to upload file: %s", resp.Status)
	}

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	insertMember(formData, currentTime)
	return nil
}
func insertMember(formData models.FormData, currentTime string) error {
	newMember := &models.Member{
		Email:  formData.Email,
		Name:   formData.Name,
		Time:   currentTime,
		Expiry: time.Now().Add(24 * 14 * time.Hour).Format(time.RFC1123),
	}

	txn := DB.Txn(true)
	defer txn.Abort()

	if err := txn.Insert("member", newMember); err != nil {
		return err
	}

	txn.Commit()

	log.Printf("Inserted member: email=%s", newMember.Email)

	return nil
}

func processName(input string) string {
	// Convert the string to lowercase
	lowercase := strings.ToLower(input)

	// Replace spaces with underscores
	result := strings.ReplaceAll(lowercase, " ", "_")

	return result
}
