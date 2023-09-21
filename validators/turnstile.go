package validators

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/ssibrahimbas/turnstile"
)

func ValidateTurnstileToken(c *gin.Context, token string, ip string) error {
	if token == "" {
		return errors.New("token is required")
	}
	if os.Getenv("GIN_MODE") != "release" {
		if token == os.Getenv("TEST_TOKEN") {
			return nil
		}
	}

	// print token
	fmt.Println(token)

	secret := os.Getenv("TURNSTILE_SECRET_KEY")

	ctx := context.Background()
	srv := turnstile.New(turnstile.Config{
		Secret: secret,
	})
	ok, err := srv.Verify(ctx, token, ip)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("token not valid")
	}
	return nil
}
