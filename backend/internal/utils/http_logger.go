package utils

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"time"
)

// LoggingRoundTripper logs HTTP requests and responses
type LoggingRoundTripper struct {
	Proxied http.RoundTripper
}

// RoundTrip implements the http.RoundTripper interface
func (lrt *LoggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	// Read process request body
	var reqBodyBytes []byte
	if req.Body != nil {
		reqBodyBytes, _ = io.ReadAll(req.Body)
		// Restore the body for the actual request
		req.Body = io.NopCloser(bytes.NewBuffer(reqBodyBytes))
	}

	reqURL := req.URL.String()

	// Perform the actual request
	resp, err := lrt.Proxied.RoundTrip(req)

	duration := time.Since(start)

	if err != nil {
		log.Printf("[SERVICE_REQUEST] url: %s | duration: %v | req: %s | error: %v", reqURL, duration, string(reqBodyBytes), err)
		return resp, err
	}

	// Read and process response body
	var respBodyBytes []byte
	if resp.Body != nil {
		respBodyBytes, _ = io.ReadAll(resp.Body)
		// Restore the body for the caller
		resp.Body = io.NopCloser(bytes.NewBuffer(respBodyBytes))
	}

	log.Printf("[SERVICE_REQUEST] url: %s | duration: %v | req: %s | resp: %s", reqURL, duration, string(reqBodyBytes), string(respBodyBytes))

	return resp, err
}
