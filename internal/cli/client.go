package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	config *Config
	client *http.Client
}

func NewClient(config *Config) *Client {
	return &Client{config: config, client: &http.Client{}}
}

func (c *Client) Request(method, path string, body interface{}) ([]byte, error) {
	url := c.config.ServerURL + path
	var bodyReader io.Reader
	var requestBody []byte
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		requestBody = data
		bodyReader = bytes.NewReader(requestBody)
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	// Try session token first, then fall back to Basic Auth
	sessionToken, err := LoadSessionToken()
	if err == nil && sessionToken != nil && sessionToken.Token != "" {
		// Use JWT bearer token
		req.Header.Set("Authorization", "Bearer "+sessionToken.Token)
	} else if c.config.Username != "" && c.config.Password != "" {
		// Fall back to Basic Auth
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.config.Verbose {
		printVerboseRequest(req, requestBody)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if c.config.Verbose {
		printVerboseResponse(resp, respBody)
	}
	if resp.StatusCode < 100 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
}

func printVerboseRequest(req *http.Request, body []byte) {
	fmt.Printf("\n[HTTP Request] %s %s\n", req.Method, req.URL.String())
	for key, values := range req.Header {
		display := strings.Join(values, ",")
		if strings.EqualFold(key, "Authorization") {
			display = "<redacted>"
		}
		fmt.Printf("> %s: %s\n", key, display)
	}
	if len(body) > 0 {
		var decoded interface{}
		if err := json.Unmarshal(body, &decoded); err == nil {
			pretty, _ := json.MarshalIndent(decoded, "", "  ")
			fmt.Printf("> Body: %s\n", string(pretty))
		} else {
			fmt.Printf("> Body: %s\n", string(body))
		}
	}
}

func printVerboseResponse(resp *http.Response, body []byte) {
	fmt.Printf("[HTTP Response] %s\n", resp.Status)
	for key, values := range resp.Header {
		fmt.Printf("< %s: %s\n", key, strings.Join(values, ","))
	}
	if len(body) > 0 {
		var decoded interface{}
		if err := json.Unmarshal(body, &decoded); err == nil {
			pretty, _ := json.MarshalIndent(decoded, "", "  ")
			fmt.Printf("< Body: %s\n\n", string(pretty))
		} else {
			fmt.Printf("< Body: %s\n\n", string(body))
		}
	}
}
