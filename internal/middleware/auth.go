package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// tokenRequest maps exactly to the expected JSON payload for the passport check.
type tokenRequest struct {
	Token string `json:"token"`
}

// TokenAuthMiddleware verifies a bearer token against the central passport service.
func TokenAuthMiddleware(passportURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			c.Abort()
			return
		}

		token := parts[1]

		// Prepare payload
		reqPayload := tokenRequest{Token: token}
		jsonData, err := json.Marshal(reqPayload)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode token request"})
			c.Abort()
			return
		}

		// Make request to external passport service
		req, err := http.NewRequest(http.MethodPost, passportURL, bytes.NewBuffer(jsonData))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create verification request"})
			c.Abort()
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			// Network error
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": fmt.Sprintf("Authentication service unavailable: %v", err)})
			c.Abort()
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			// Verification failed
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Success -> proceed to handler
		c.Next()
	}
}
