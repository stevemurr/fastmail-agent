package jmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	FastmailSessionURL = "https://api.fastmail.com/jmap/session"
)

// Client is a JMAP client for Fastmail
type Client struct {
	token      string
	httpClient *http.Client
	session    *Session
	accountID  string
}

// NewClient creates a new JMAP client
func NewClient(token string) *Client {
	return &Client{
		token:      token,
		httpClient: &http.Client{},
	}
}

// Connect establishes a session with Fastmail
func (c *Client) Connect() error {
	session, err := c.getSession()
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	c.session = session

	// Get the primary account for mail
	accountID, ok := session.PrimaryAccount["urn:ietf:params:jmap:mail"]
	if !ok {
		return fmt.Errorf("no mail account found")
	}
	c.accountID = accountID

	return nil
}

// getSession fetches the JMAP session
func (c *Client) getSession() (*Session, error) {
	req, err := http.NewRequest("GET", FastmailSessionURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("session request failed: %s - %s", resp.Status, string(body))
	}

	var session Session
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, err
	}

	return &session, nil
}

// Call makes a JMAP API call
func (c *Client) Call(methodCalls []Invocation) (*Response, error) {
	if c.session == nil {
		return nil, fmt.Errorf("not connected")
	}

	request := Request{
		Using: []string{
			"urn:ietf:params:jmap:core",
			"urn:ietf:params:jmap:mail",
		},
		MethodCalls: methodCalls,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.session.APIURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API call failed: %s - %s", resp.Status, string(respBody))
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// AccountID returns the primary account ID
func (c *Client) AccountID() string {
	return c.accountID
}

// Session returns the current session
func (c *Client) Session() *Session {
	return c.session
}

// DownloadBlob downloads an attachment blob by ID
func (c *Client) DownloadBlob(blobID, name, mimeType string) ([]byte, error) {
	if c.session == nil {
		return nil, fmt.Errorf("not connected")
	}

	// Build download URL from session template
	// Template format: https://www.fastmailusercontent.com/jmap/download/{accountId}/{blobId}/{name}?type={type}
	downloadURL := c.session.DownloadURL
	downloadURL = strings.ReplaceAll(downloadURL, "{accountId}", c.accountID)
	downloadURL = strings.ReplaceAll(downloadURL, "{blobId}", blobID)
	downloadURL = strings.ReplaceAll(downloadURL, "{name}", url.PathEscape(name))
	downloadURL = strings.ReplaceAll(downloadURL, "{type}", url.QueryEscape(mimeType))

	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("download failed: %s - %s", resp.Status, string(body))
	}

	return io.ReadAll(resp.Body)
}
