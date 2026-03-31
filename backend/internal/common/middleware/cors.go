package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS configures cross-origin resource sharing with origin whitelist.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	originSet := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originSet[o] = true
	}
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if originSet[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
		} else if len(allowedOrigins) == 0 {
			c.Header("Access-Control-Allow-Origin", "*")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Trace-ID")
		c.Header("Access-Control-Expose-Headers", "X-Trace-ID")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// AdminAuth checks X-Admin-Token header for admin endpoints.
func AdminAuth(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if token == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access disabled"})
			return
		}
		if c.GetHeader("X-Admin-Token") != token {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid admin token"})
			return
		}
		c.Next()
	}
}
