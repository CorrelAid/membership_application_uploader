package validators

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/ssibrahimbas/turnstile"
)

func ValidateTurnstileToken(c *gin.Context, token string, ip string) error {
	if token == "" {
		log.Println("Token is required")
		return errors.New("token is required")
	}
	if os.Getenv("GIN_MODE") != "release" {
		if token == os.Getenv("TEST_TOKEN") {
			log.Println("Test token used")
			return nil
		}
	}

	secret := os.Getenv("TURNSTILE_SECRET_KEY")

	ctx := context.Background()
	srv := turnstile.New(turnstile.Config{
		Secret: secret,
	})
	ok, err := srv.Verify(ctx, token, ip)

	if err != nil {
		// print error
		log.Println("Verification error:", err)
		// return internal server error and status code 500
		return errors.New("internal_server_error")

	}
	if !ok {
		log.Println("Token not valid")
		return errors.New("token_not_valid")
	}
	return nil
}
