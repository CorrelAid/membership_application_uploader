//go:build linux
// +build linux

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
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)

func main() {

	ginMode := os.Getenv("GIN_MODE")
	if ginMode != "release" {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatalf("Error loading .env file: %s", err.Error())
		}
	}

	inits.DBInit()

	router := gin.Default()

	rateLimiter := rate.NewLimiter(rate.Every(time.Minute), 10)

	// Configure IP-based rate limiting middleware
	router.Use(func(c *gin.Context) {
		if !rateLimiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"message": "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}
		c.Next()
	})

	// Configure CORS middleware
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://correlaid.org", "http://localhost"}
	config.AllowMethods = []string{"POST"}
	router.Use(cors.New(config))

	// Set a lower memory limit for multipart forms (default is 32 MiB)
	router.MaxMultipartMemory = 8 << 20 // 8 MiB
	router.ForwardedByClientIP = true
	router.SetTrustedProxies(nil)
	router.POST("/", handle)
	// Listen on the correct port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default to port 8080 if PORT environment variable is not set
	}

	err := router.Run(":" + port)
	if err != nil {
		log.Fatal(err)
	}

}

func handle(c *gin.Context) {

	ip := c.ClientIP()

	err := validators.ValidateTurnstileToken(c, c.PostForm("token"), ip)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error validating token: "+err.Error())
		return
	}

	formFile, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, "Error getting file: "+err.Error())
		return
	}

	formData := models.FormData{
		File:  formFile,
		Name:  c.PostForm("name"),
		Email: c.PostForm("email"),
	}

	processedFormData, err := validators.ValidateProcessFormData(formData)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	currentTime := time.Now().Format(time.RFC1123)

	// Lookup by email
	txn := inits.DB.Txn(false)
	defer txn.Abort()

	raw, err := txn.First("member", "id", processedFormData.Email)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err.Error()))
		return
	}

	if raw != nil {
		c.String(http.StatusBadRequest, "Email already exists")
		return
	}

	if err := uploadFileToNextcloud(processedFormData, currentTime); err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err.Error()))
		return
	}

	c.String(http.StatusOK, "File uploaded successfully")
}

func uploadFileToNextcloud(processedFormData models.ProcessedFormData, currentTime string) error {
	client := &http.Client{}

	filename := fmt.Sprintf("%s_%s", processName(processedFormData.Name), currentTime)

	req, err := http.NewRequest(http.MethodPut, "https://cloud.correlaid.org/remote.php/dav/files/bot@correlaid.org/Mitgliedsanträge/"+filename+".pdf", bytes.NewReader(processedFormData.FileContent))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/octet-stream")

	pw := os.Getenv("NEXTCLOUD_PW")
	user := os.Getenv("NEXTCLOUD_USER")
	req.SetBasicAuth(user, pw)

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
	insertMember(processedFormData, currentTime)
	return nil
}
func insertMember(processedFormData models.ProcessedFormData, currentTime string) error {
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

func processName(input string) string {
	// Convert the string to lowercase
	lowercase := strings.ToLower(input)

	// Replace spaces with underscores
	result := strings.ReplaceAll(lowercase, " ", "_")

	return result
}
