// middleware/sanitization.go
package middleware

import (
	// "net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// SanitizationMiddleware membersihkan input dari potential SQL injection
func SanitizationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sanitize query parameters
		query := c.Request.URL.Query()
		for key, values := range query {
			for i, value := range values {
				query[key][i] = sanitizeString(value)
			}
		}
		c.Request.URL.RawQuery = query.Encode()

		// Sanitize path parameters
		for _, param := range c.Params {
			param.Value = sanitizeString(param.Value)
		}

		c.Next()
	}
}

func sanitizeString(input string) string {
	// Remove potentially dangerous SQL characters
	reg := regexp.MustCompile(`[<>"'%;()&+*|=/]`)
	sanitized := reg.ReplaceAllString(input, "")
	
	// Trim whitespace and limit length
	sanitized = strings.TrimSpace(sanitized)
	if len(sanitized) > 255 {
		sanitized = sanitized[:255]
	}
	
	return sanitized
}