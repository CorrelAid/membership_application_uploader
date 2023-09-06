# Membership Application Uploader
This Go application handles the upload of PDF files for membership applications. It provides an HTTP endpoint for uploading files and storing the application details in a database. The uploaded files are also sent to a Nextcloud server for storage.

## Prerequisites
Go (version 1.16 or higher)
github.com/joho/godotenv package (go get github.com/joho/godotenv)
github.com/gin-gonic/gin package (go get github.com/gin-gonic/gin)
github.com/hashicorp/go-memdb package (go get github.com/hashicorp/go-memdb)

## Development
1. Clone the repository.
2. Install the required packages using go get.
3. Set the necessary environment variable NEXTCLOUD_PW
4. Run the application using go run main.go.

## API Endpoint
POST /upload: Handles the upload of a PDF file for a membership application.

## Functionality
- Retrieves the uploaded file from the request.
- Validates the form data.
- Checks if the email already exists in-memory database.
- Uploads the file to the Nextcloud server.
- Inserts the member details into the in-memory database.
- Runs a background routine that cleans up expired data entries


## Test 
1. Existing email
```bash
curl -X POST http://localhost:8080/upload \
  -F "file=@foo.pdf" \
  -H "Content-Type: multipart/form-data" -F "name=Test Name" -F "email=test@example.com"

```
2. New email
```bash
curl -X POST http://localhost:8080/upload \
  -F "file=@foo.pdf" \
  -H "Content-Type: multipart/form-data" -F "name=Test Name" -F "email=test2@example.com"

```