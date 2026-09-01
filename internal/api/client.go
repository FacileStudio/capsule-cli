package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client talks to a Capsule instance API.
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// Error is an error returned by the server API.
type Error struct {
	Status  int
	Code    string
	Message string
}

// Error returns the formatted error message.
func (e *Error) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("the server refused the request (%d)", e.Status)
	}
	return e.Message
}

// Unauthenticated reports whether the request failed due to authentication.
func (e *Error) Unauthenticated() bool {
	return e.Status == http.StatusUnauthorized || e.Status == http.StatusForbidden
}

// CreatePasteRequest holds fields for creating a new paste.
type CreatePasteRequest struct {
	Content       string `json:"content"`
	BurnAfterRead bool   `json:"burn_after_read"`
	ExpiresIn     string `json:"expires_in"`
	HasPassword   bool   `json:"has_password"`
	Syntax        string `json:"syntax,omitempty"`
}

// CreatePasteResponse holds the response after creating a paste.
type CreatePasteResponse struct {
	ID          string `json:"id"`
	DeleteToken string `json:"delete_token"`
	ExpiresAt   string `json:"expires_at"`
	CreatedAt   string `json:"created_at"`
}

// PasteMetadata holds metadata for a paste.
type PasteMetadata struct {
	ID          string `json:"id"`
	Exists      bool   `json:"exists"`
	Burned      bool   `json:"burned"`
	HasPassword bool   `json:"has_password"`
	Syntax      string `json:"syntax"`
	ExpiresAt   string `json:"expires_at"`
	CreatedAt   string `json:"created_at"`
}

// PasteContent holds the content of a paste.
type PasteContent struct {
	Content string `json:"content"`
}

// DeleteResponse holds the response after deleting a paste.
type DeleteResponse struct {
	Deleted bool `json:"deleted"`
}

// NewClient creates a new API client with the given base URL.
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// New creates a new API client with the given base URL and authentication token.
func New(baseURL, token string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(encoded)
	}

	request, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, reader)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/json")
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if c.Token != "" {
		request.Header.Set("Authorization", "Bearer "+c.Token)
	}

	response, err := c.HTTPClient.Do(request)
	if err != nil {
		return fmt.Errorf("cannot reach %s: %w", c.BaseURL, err)
	}
	defer response.Body.Close()

	data, err := io.ReadAll(io.LimitReader(response.Body, 8<<20))
	if err != nil {
		return err
	}
	if response.StatusCode >= 400 {
		return decodeError(response.StatusCode, data)
	}
	if out == nil || len(data) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}
	return nil
}

func decodeError(status int, data []byte) error {
	var envelope struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(data, &envelope); err == nil && envelope.Error.Message != "" {
		return &Error{Status: status, Code: envelope.Error.Code, Message: envelope.Error.Message}
	}
	return &Error{Status: status, Code: "unknown", Message: fmt.Sprintf("server returned %d: %s", status, string(data))}
}

// CreatePaste creates and uploads a new paste.
func (c *Client) CreatePaste(req *CreatePasteRequest) (*CreatePasteResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	resp, err := c.HTTPClient.Post(c.BaseURL+"/api/pastes", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating paste: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result CreatePasteResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// GetMetadata retrieves metadata for a paste.
func (c *Client) GetMetadata(id string) (*PasteMetadata, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/api/pastes/" + id)
	if err != nil {
		return nil, fmt.Errorf("getting metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result PasteMetadata
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// GetContent retrieves the content of a paste.
func (c *Client) GetContent(id string) (*PasteContent, error) {
	resp, err := c.HTTPClient.Post(c.BaseURL+"/api/pastes/"+id+"/content", "application/json", nil)
	if err != nil {
		return nil, fmt.Errorf("getting content: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result PasteContent
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// Delete removes a paste using its delete token.
func (c *Client) Delete(id string, token string) error {
	req, err := http.NewRequest(http.MethodDelete, c.BaseURL+"/api/pastes/"+id, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("X-Delete-Token", token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("deleting paste: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
