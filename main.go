package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type FormData struct {
	File  []byte
	Name  string
	Email string
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	router := gin.Default()
	// Set a lower memory limit for multipart forms (default is 32 MiB)
	router.MaxMultipartMemory = 8 << 20 // 8 MiB
	router.POST("/upload", uploadPDF)
	router.Run(":8080")
}

const MaxFileSize = 3 * 1024 * 1024 // 3 MB

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
	formData := FormData{
		File:  data,
		Name:  c.PostForm("name"),
		Email: c.PostForm("email"),
	}

	if err := validateFormData(formData); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	currentTime := time.Now()
	if err := uploadFileToNextcloud(formData, currentTime); err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err.Error()))
		return
	}
	c.String(http.StatusOK, "File uploaded successfully")
}

func uploadFileToNextcloud(formData FormData, currentTime time.Time) error {
	client := &http.Client{}

	filename := fmt.Sprintf("%s_%s", processName(formData.Name), currentTime.Format("2006-01-02"))

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
	return nil
}

func validateFormData(formData FormData) error {
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

func processName(input string) string {
	// Convert the string to lowercase
	lowercase := strings.ToLower(input)

	// Replace spaces with underscores
	result := strings.ReplaceAll(lowercase, " ", "_")

	return result
}
