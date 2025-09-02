package train

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	BaseURL            = "https://eticket.railway.uz/api/v3"
	BaseURLv1          = "https://eticket.railway.uz/api/v1"
	TrainsListEndpoint = "/handbook/trains/list"
	CSRFTokenEndpoint  = "/csrf-token"
	UserAgent          = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36"
)

// Supported languages for Railway.uz API
const (
	LanguageUzbek   = "uz"
	LanguageRussian = "ru"
	LanguageEnglish = "en"
)

// Client represents the train ticket API client
type Client struct {
	httpClient *http.Client
	baseURL    string
	headers    map[string]string
	language   string
}

// NewClient creates a new train API client
func NewClient(language string) *Client {
	// Default to Uzbek if no language specified
	if language == "" {
		language = LanguageUzbek
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:  BaseURL,
		language: language,
		headers: map[string]string{
			"Accept":          "application/json",
			"Accept-Language": language,
			"Content-Type":    "application/json",
			"User-Agent":      UserAgent,
		},
	}
}

// SetAuthHeaders sets authentication headers for the client
func (c *Client) SetAuthHeaders(xsrfToken, cookies string) {
	c.headers["X-XSRF-TOKEN"] = xsrfToken
	c.headers["Cookie"] = cookies
}

// SetLanguage changes the Accept-Language header for API requests
func (c *Client) SetLanguage(language string) {
	if language == "" {
		language = LanguageUzbek // Default to Uzbek
	}
	c.language = language
	c.headers["Accept-Language"] = language
}

// GetLanguage returns the current language setting
func (c *Client) GetLanguage() string {
	return c.language
}

// makeRequest makes an HTTP request to the API
func (c *Client) makeRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	return resp, nil
}

// InitializeCredentials automatically obtains fresh CSRF token and cookies
func (c *Client) InitializeCredentials(ctx context.Context) error {
	log.Printf("🔄 Initializing Railway.uz API credentials...")

	// Get fresh CSRF token
	token, err := c.RefreshCSRFToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize CSRF token: %w", err)
	}

	// Set the token in headers
	c.headers["X-XSRF-TOKEN"] = token

	// Create initial cookies with the new token
	cookies := fmt.Sprintf("XSRF-TOKEN=%s", token)
	c.headers["Cookie"] = cookies

	log.Printf("✅ Railway.uz API credentials initialized successfully")
	log.Printf("   XSRF-TOKEN: %s", token[:8]+"...")
	log.Printf("   Cookies: %s", cookies[:20]+"...")

	return nil
}

// RefreshCSRFToken refreshes the CSRF token using the /api/v1/csrf-token endpoint
func (c *Client) RefreshCSRFToken(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", BaseURLv1+CSRFTokenEndpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create CSRF request: %w", err)
	}

	// Set minimal headers for CSRF token request
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", c.language)
	req.Header.Set("User-Agent", UserAgent)

	// Add existing cookies if available
	if cookies, exists := c.headers["Cookie"]; exists {
		req.Header.Set("Cookie", cookies)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make CSRF request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("CSRF request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Extract XSRF-TOKEN from Set-Cookie header
	for _, cookie := range resp.Header["Set-Cookie"] {
		if strings.Contains(cookie, "XSRF-TOKEN=") {
			// Extract token value using regex
			re := regexp.MustCompile(`XSRF-TOKEN=([^;]+)`)
			matches := re.FindStringSubmatch(cookie)
			if len(matches) > 1 {
				return matches[1], nil
			}
		}
	}

	return "", fmt.Errorf("XSRF-TOKEN not found in response")
}

// SearchTrains searches for available trains with automatic token refresh
func (c *Client) SearchTrains(ctx context.Context, req *SearchTrainsRequest) (*SearchTrainsResponse, error) {
	resp, err := c.makeRequest(ctx, "POST", TrainsListEndpoint, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// If we get a 403 CSRF error, try to refresh the token and retry once
	if resp.StatusCode == 403 {
		body, _ := io.ReadAll(resp.Body)
		if strings.Contains(string(body), "CSRF") || strings.Contains(string(body), "Invalid CSRF Token") {
			// Try to refresh the CSRF token
			newToken, refreshErr := c.RefreshCSRFToken(ctx)
			if refreshErr != nil {
				return nil, fmt.Errorf("failed to refresh CSRF token: %w", refreshErr)
			}

			// Update the token in headers
			c.headers["X-XSRF-TOKEN"] = newToken

			// Update cookies to include new XSRF-TOKEN
			if cookies, exists := c.headers["Cookie"]; exists {
				// Replace or add XSRF-TOKEN in cookies
				re := regexp.MustCompile(`XSRF-TOKEN=[^;]*`)
				if re.MatchString(cookies) {
					cookies = re.ReplaceAllString(cookies, "XSRF-TOKEN="+newToken)
				} else {
					cookies = "XSRF-TOKEN=" + newToken + ";" + cookies
				}
				c.headers["Cookie"] = cookies
			} else {
				c.headers["Cookie"] = "XSRF-TOKEN=" + newToken
			}

			// Retry the request with new token
			resp, err = c.makeRequest(ctx, "POST", TrainsListEndpoint, req)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
		}
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result SearchTrainsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
