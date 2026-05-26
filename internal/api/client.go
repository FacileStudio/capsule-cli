package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

type CreatePasteRequest struct {
	Content       string `json:"content"`
	BurnAfterRead bool   `json:"burn_after_read"`
	ExpiresIn     string `json:"expires_in"`
	HasPassword   bool   `json:"has_password"`
	Syntax        string `json:"syntax,omitempty"`
}

type CreatePasteResponse struct {
	ID          string `json:"id"`
	DeleteToken string `json:"delete_token"`
	ExpiresAt   string `json:"expires_at"`
	CreatedAt   string `json:"created_at"`
}

type PasteMetadata struct {
	ID          string `json:"id"`
	Exists      bool   `json:"exists"`
	Burned      bool   `json:"burned"`
	HasPassword bool   `json:"has_password"`
	Syntax      string `json:"syntax"`
	ExpiresAt   string `json:"expires_at"`
	CreatedAt   string `json:"created_at"`
}

type PasteContent struct {
	Content string `json:"content"`
}

type DeleteResponse struct {
	Deleted bool `json:"deleted"`
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

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

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
