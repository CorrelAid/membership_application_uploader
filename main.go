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
	"strconv"
	"strings"
	"time"

	"github.com/CorrelAid/membership_application_uploader/inits"
	"github.com/CorrelAid/membership_application_uploader/middleware"
	"github.com/CorrelAid/membership_application_uploader/models"
	"github.com/CorrelAid/membership_application_uploader/operations"
	"github.com/CorrelAid/membership_application_uploader/validators"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var MaxSize int
var WebDavURL string

func main() {
	ginMode := os.Getenv("GIN_MODE")
	if ginMode != "release" {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatalf("error loading .env file: %s", err.Error())
		}
	}

	maxSize, err := parseEnvInt("MAX_FILE_SIZE")
	if err != nil {
		panic(err)
	}

	MaxSize = maxSize

	webDavURL := os.Getenv("WEBDAV_URL")
	if webDavURL == "" {
		panic("Missing WEBDAV_URL")
	}
	WebDavURL = webDavURL

	MaxRequests, err := parseEnvFloat64("MAX_REQUESTS_PER_MINUTE")
	if err != nil {
		panic(err)
	}

	inits.DBInit()

	router := gin.Default()

	// Rate Limiting
	router.Use(middleware.RateLimitMiddleware(MaxRequests))

	// CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://correlaid.org", "http://localhost"}
	config.AllowMethods = []string{"POST"}
	router.Use(cors.New(config))

	// Set a lower memory limit for multipart forms (default is 32 MiB)
	router.MaxMultipartMemory = 3 << 20 // 3 MiB
	router.ForwardedByClientIP = true
	router.SetTrustedProxies(nil)
	router.POST("/", handle)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default to port 8080 if PORT environment variable is not set
	}

	err = router.Run(":" + port)
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

	processedFormData, err := validators.ValidateProcessFormData(formData, MaxSize)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	currentTime := time.Now().Format("2006-01-02_15-04-05")

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

	lowercase := strings.ToLower(processedFormData.Name)
	result := strings.ReplaceAll(lowercase, " ", "_")

	filename := fmt.Sprintf("%s_%s", result, currentTime)

	fmt.Println(filename)

	req, err := http.NewRequest(http.MethodPut, WebDavURL+"/"+filename+".pdf", bytes.NewReader(processedFormData.FileContent))
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
	operations.InsertMember(processedFormData, currentTime)
	return nil
}

func parseEnvInt(key string) (int, error) {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return 0, fmt.Errorf("missing %s", key)
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("error converting %s to int", key)
	}
	return value, nil
}

func parseEnvFloat64(key string) (float64, error) {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return 0, fmt.Errorf("missing %s", key)
	}
	value64, err := strconv.ParseFloat(valueStr, 32)
	if err != nil {
		return 0, fmt.Errorf("error converting %s to float32", key)
	}
	return value64, nil
}
