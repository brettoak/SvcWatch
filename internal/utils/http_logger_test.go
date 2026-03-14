package utils

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestLoggingRoundTripper(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	// Create a mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if string(body) != "ping" {
			t.Errorf("expected body 'ping', got '%s'", string(body))
		}
		w.Write([]byte("pong"))
	}))
	defer ts.Close()

	// Use our RoundTripper
	client := ts.Client()
	client.Transport = &LoggingRoundTripper{Proxied: client.Transport}

	// Make request
	req, err := http.NewRequest("POST", ts.URL, strings.NewReader("ping"))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if string(respBody) != "pong" {
		t.Errorf("expected response 'pong', got '%s'", string(respBody))
	}

	out := buf.String()
	t.Logf("Log output: %s", out)
	if !strings.Contains(out, "[SERVICE_REQUEST]") {
		t.Errorf("Expected log containing [SERVICE_REQUEST], got: %s", out)
	}
	if !strings.Contains(out, "req: ping") {
		t.Errorf("Expected log containing req: ping, got: %s", out)
	}
	if !strings.Contains(out, "resp: pong") {
		t.Errorf("Expected log containing resp: pong, got: %s", out)
	}
	if !strings.Contains(out, "url: "+ts.URL) {
		t.Errorf("Expected log containing url: %s, got: %s", ts.URL, out)
	}
}
