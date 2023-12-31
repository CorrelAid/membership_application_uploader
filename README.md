# Membership Application Uploader
This Go application handles the upload of CorrelAid membership applications to a Nextcloud instance. 

## Description of Functionality
When the application starts, it initializes configurations and environment variables, and establishes a connection to an in-memory database. The application creates a Gin router and applies middleware for rate limiting per IP and CORS. When a POST request is made to the server, the application's route handler is invoked. In the route handler, it 
- first validates a token and client IP address using the **Turnstile** verify endpoint. 
- Next, the application validates and processes the form data, including checking the file size and syntax checking the email. It performs a database lookup to check if the email already exists.
- Afterwards, the application uploads the file to a Nextcloud server using WebDAV. It inserts the email in the database to be able to check for duplicate emails in the future.

## Development
1. Clone the repository.
2. Install the required packages using go get.
3. Set the necessary environment variables. Create a .env file that looks like this:
  ```bash
  NEXTCLOUD_PW=pw
  NEXTCLOUD_USER=user
  WEBDAV_URL=https://nextloud_instance/remote.php/dav/files/user/applications
  TURNSTILE_SECRET_KEY=secret
  TEST_TOKEN=tst
  MAX_FILE_SIZE=3145728 #3mb
  MAX_REQUESTS_PER_MINUTE=5
  ```
4. Run the application using go run main.go.


## Test 
1. Existing email without token
```bash
curl -X POST http://localhost:8080 \
  -F "file=@foo.pdf" \
  -H "Content-Type: multipart/form-data" -F "name=Test Name" -F "email=test@example.com" -F "token=<your_test_token>"

```
2. New email without token 
```bash
curl -X POST http://localhost:8080 \
  -F "file=@foo.pdf" \
  -H "Content-Type: multipart/form-data" -F "name=Test Name" -F "email=test2@example.com" -F "token=<your_test_token>"
```