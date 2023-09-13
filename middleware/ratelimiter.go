package middleware

import (
	"net/http"
	"time"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gin-gonic/gin"
)

func RateLimitMiddleware(maxRequests float64) gin.HandlerFunc {
	per_second := maxRequests / 60.0
	limiter := tollbooth.NewLimiter(per_second, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Minute})

	limiter.SetIPLookups([]string{"RemoteAddr", "X-Forwarded-For", "X-Real-IP"})

	return func(c *gin.Context) {
		httpError := tollbooth.LimitByRequest(limiter, c.Writer, c.Request)
		if httpError != nil {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, Message{
				Status: "Request Failed",
				Body:   "The API is at capacity, try again later.",
			})
			return
		}
		c.Next()
	}
}

type Message struct {
	Status string `json:"status"`
	Body   string `json:"body"`
}
