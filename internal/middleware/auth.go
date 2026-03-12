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
			// HTTP layer check failed
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to connect to passport service or token rejected"})
			c.Abort()
			return
		}

		var passportResp struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    struct {
				Active bool `json:"active"`
				Valid  bool `json:"valid"`
			} `json:"data"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&passportResp); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse authentication response"})
			c.Abort()
			return
		}

		if passportResp.Code != 200 || !passportResp.Data.Active || !passportResp.Data.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token validation failed or token is inactive"})
			c.Abort()
			return
		}

		// Success -> proceed to handler
		c.Next()
	}
}

// permissionRequest maps exactly to the expected JSON payload for checking action-specific permissions.
type permissionRequest struct {
	Token              string `json:"token"`
	SysCode            string `json:"sysCode"`
	RequiredPermission string `json:"requiredPermission"`
	Path               string `json:"path"`
	Method             string `json:"method"`
}

// PermissionMiddleware verifies if the token bearer has the required permission via the passport service.
// This should be chained AFTER TokenAuthMiddleware so we can assume the token format is somewhat valid.
func PermissionMiddleware(permissionURL, sysCode, requiredPermission string) gin.HandlerFunc {
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

		reqPayload := permissionRequest{
			Token:              token,
			SysCode:            sysCode,
			RequiredPermission: requiredPermission,
			Path:               c.Request.URL.Path,
			Method:             c.Request.Method,
		}

		jsonData, err := json.Marshal(reqPayload)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode permission request"})
			c.Abort()
			return
		}

		req, err := http.NewRequest(http.MethodPost, permissionURL, bytes.NewBuffer(jsonData))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create permission request"})
			c.Abort()
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		// Optional: Propagate the original Authorization header to the passport service as well
		req.Header.Set("Authorization", authHeader)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": fmt.Sprintf("Permission service unavailable: %v", err)})
			c.Abort()
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Permission denied by passport service"})
			c.Abort()
			return
		}

		var passportResp struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			// Assuming the structure requires code 200 for permission success.
		}

		if err := json.NewDecoder(resp.Body).Decode(&passportResp); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse permission response"})
			c.Abort()
			return
		}

		if passportResp.Code != 200 {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}
