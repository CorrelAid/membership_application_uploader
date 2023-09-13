package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func RateLimitMiddleware(maxRequests float32) gin.HandlerFunc {
	per_second := maxRequests / 60.0
	limiter := rate.NewLimiter(rate.Limit(per_second), 1)
	return func(c *gin.Context) {
		if !limiter.Allow() {
			message := Message{
				Status: "Request Failed",
				Body:   "The API is at capacity, try again later.",
			}

			c.JSON(http.StatusTooManyRequests, message)
			c.Abort()
			return
		}

		c.Next()
	}
}

type Message struct {
	Status string `json:"status"`
	Body   string `json:"body"`
}
