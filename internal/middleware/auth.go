package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"SvcWatch/internal/utils"

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
			utils.Error(c, http.StatusUnauthorized, "Authorization header required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utils.Error(c, http.StatusUnauthorized, "Authorization header format must be Bearer {token}")
			c.Abort()
			return
		}

		token := parts[1]

		// Prepare payload
		reqPayload := tokenRequest{Token: token}
		jsonData, err := json.Marshal(reqPayload)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, "Failed to encode token request")
			c.Abort()
			return
		}

		// Make request to external passport service
		req, err := http.NewRequest(http.MethodPost, passportURL, bytes.NewBuffer(jsonData))
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, "Failed to create verification request")
			c.Abort()
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		client := &http.Client{
			Transport: &utils.LoggingRoundTripper{Proxied: http.DefaultTransport},
		}
		resp, err := client.Do(req)
		if err != nil {
			// Network error
			utils.Error(c, http.StatusServiceUnavailable, fmt.Sprintf("Authentication service unavailable: %v", err))
			c.Abort()
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			// HTTP layer check failed
			utils.Error(c, http.StatusUnauthorized, "Failed to connect to passport service or token rejected")
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
			utils.Error(c, http.StatusInternalServerError, "Failed to parse authentication response")
			c.Abort()
			return
		}

		if passportResp.Code != 200 || !passportResp.Data.Active || !passportResp.Data.Valid {
			utils.Error(c, http.StatusUnauthorized, "Token validation failed or token is inactive")
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
			utils.Error(c, http.StatusUnauthorized, "Authorization header required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utils.Error(c, http.StatusUnauthorized, "Authorization header format must be Bearer {token}")
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
			utils.Error(c, http.StatusInternalServerError, "Failed to encode permission request")
			c.Abort()
			return
		}

		req, err := http.NewRequest(http.MethodPost, permissionURL, bytes.NewBuffer(jsonData))
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, "Failed to create permission request")
			c.Abort()
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		// Optional: Propagate the original Authorization header to the passport service as well
		req.Header.Set("Authorization", authHeader)

		client := &http.Client{
			Transport: &utils.LoggingRoundTripper{Proxied: http.DefaultTransport},
		}
		resp, err := client.Do(req)
		if err != nil {
			utils.Error(c, http.StatusServiceUnavailable, fmt.Sprintf("Permission service unavailable: %v", err))
			c.Abort()
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			utils.Error(c, http.StatusForbidden, "Forbidden: Permission denied by passport service")
			c.Abort()
			return
		}

		var passportResp struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    struct {
				Valid         bool `json:"valid"`
				HasPermission bool `json:"hasPermission"`
			} `json:"data"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&passportResp); err != nil {
			utils.Error(c, http.StatusInternalServerError, "Failed to parse permission response")
			c.Abort()
			return
		}

		if passportResp.Code != 200 || !passportResp.Data.Valid || !passportResp.Data.HasPermission {
			utils.Error(c, http.StatusForbidden, "Forbidden: Insufficient permissions or token invalid")
			c.Abort()
			return
		}

		c.Next()
	}
}
