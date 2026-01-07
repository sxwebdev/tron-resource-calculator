package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sxwebdev/tron-resource-calculator/internal/models"
)

const (
	defaultTimeout = 5 * time.Second
	maxRetries     = 3
	initialBackoff = 100 * time.Millisecond
)

// Client is an HTTP client for TRON API
type Client struct {
	nodeURL    string
	httpClient *http.Client
}

// New creates a new TRON API client
func New(nodeURL string) *Client {
	return &Client{
		nodeURL: strings.TrimSuffix(nodeURL, "/"),
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// GetAccountResource fetches account resources from TRON API
func (c *Client) GetAccountResource(address string) (*models.APIResponse, error) {
	url := c.nodeURL + "/wallet/getaccountresource"

	payload := map[string]interface{}{
		"address": address,
		"visible": true,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	var lastErr error
	backoff := initialBackoff

	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err := c.doRequest(url, body)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		if attempt < maxRetries {
			time.Sleep(backoff)
			backoff *= 2 // exponential backoff
		}
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

func (c *Client) doRequest(url string, body []byte) (*models.APIResponse, error) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result models.APIResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// ValidateAddress checks if the given string is a valid TRON address
func ValidateAddress(address string) error {
	if len(address) != 34 {
		return fmt.Errorf("invalid address length: expected 34, got %d", len(address))
	}
	if !strings.HasPrefix(address, "T") {
		return fmt.Errorf("invalid address format: must start with 'T'")
	}
	return nil
}
