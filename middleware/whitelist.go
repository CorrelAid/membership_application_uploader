package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func DomainWhitelistMiddleware(allowedDomains []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		host := c.Request.Host
		// print host
		allowed := false
		fmt.Println(host)
		for _, domain := range allowedDomains {
			if strings.EqualFold(domain, host) {
				allowed = true
				break
			}
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"status":  http.StatusForbidden,
				"message": "Permission denied",
			})
			return
		}

		c.Next()
	}
}
