package confluence

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	authHeader string
	httpClient *http.Client
}

type Options struct {
	BaseURL    string
	Email      string
	Token      string
	HTTPClient *http.Client
	Timeout    time.Duration
}

func NewClient(opts Options) *Client {
	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: opts.Timeout}
	}

	baseURL := strings.TrimRight(opts.BaseURL, "/")
	auth := base64.StdEncoding.EncodeToString([]byte(opts.Email + ":" + opts.Token))

	return &Client{
		baseURL:    baseURL + "/wiki/api/v2",
		authHeader: "Basic " + auth,
		httpClient: httpClient,
	}
}

// APIError represents an error from the Confluence API
type APIError struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("confluence API error (status %d): %s", e.StatusCode, e.Message)
}

func (c *Client) do(method, path string, query url.Values) ([]byte, error) {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequest(method, u, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", c.authHeader)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		apiErr := &APIError{StatusCode: resp.StatusCode}
		// Try to parse error message from response
		var errResp struct {
			Message string `json:"message"`
			// v2 API error format
			Errors []struct {
				Message string `json:"message"`
			} `json:"errors"`
		}
		if json.Unmarshal(body, &errResp) == nil {
			if errResp.Message != "" {
				apiErr.Message = errResp.Message
			} else if len(errResp.Errors) > 0 {
				apiErr.Message = errResp.Errors[0].Message
			}
		}
		if apiErr.Message == "" {
			apiErr.Message = string(body)
		}
		return nil, apiErr
	}

	return body, nil
}

func (c *Client) doV1(method, path string, query url.Values) ([]byte, error) {
	// Replace /wiki/api/v2 base with /wiki/rest/api for v1 endpoints
	baseV1 := strings.Replace(c.baseURL, "/wiki/api/v2", "/wiki/rest/api", 1)
	u := baseV1 + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequest(method, u, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", c.authHeader)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		apiErr := &APIError{StatusCode: resp.StatusCode}
		var errResp struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(body, &errResp) == nil && errResp.Message != "" {
			apiErr.Message = errResp.Message
		} else {
			apiErr.Message = string(body)
		}
		return nil, apiErr
	}

	return body, nil
}

// paginatedResponse is the raw Confluence v2 paginated response
type paginatedResponse struct {
	Results json.RawMessage `json:"results"`
	Links   struct {
		Next string `json:"next"`
	} `json:"_links"`
}

// extractCursor extracts the cursor parameter from a next link URL
func extractCursor(nextLink string) string {
	if nextLink == "" {
		return ""
	}
	u, err := url.Parse(nextLink)
	if err != nil {
		return ""
	}
	return u.Query().Get("cursor")
}
