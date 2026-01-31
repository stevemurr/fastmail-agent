package jmap

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// SearchEmails searches for emails matching the query
func (c *Client) SearchEmails(query string, limit int) ([]Email, error) {
	if limit <= 0 {
		limit = 50
	}

	// Build filter - search all mail with text query
	filter := map[string]interface{}{}
	if query != "" {
		filter["text"] = query
	}

	// Query for email IDs
	calls := []Invocation{
		NewInvocation("Email/query", map[string]interface{}{
			"accountId": c.accountID,
			"filter":    filter,
			"sort": []map[string]interface{}{
				{"property": "receivedAt", "isAscending": false},
			},
			"limit": limit,
		}, "0"),
		NewInvocation("Email/get", map[string]interface{}{
			"accountId": c.accountID,
			"#ids": map[string]interface{}{
				"resultOf": "0",
				"name":     "Email/query",
				"path":     "/ids",
			},
			"properties": []string{
				"id", "threadId", "mailboxIds", "from", "to", "cc",
				"subject", "receivedAt", "preview",
			},
		}, "1"),
	}

	resp, err := c.Call(calls)
	if err != nil {
		return nil, err
	}

	if len(resp.MethodResponses) < 2 {
		return nil, fmt.Errorf("unexpected response")
	}

	// Parse Email/get response
	mr, err := ParseMethodResponse(resp.MethodResponses[1])
	if err != nil {
		return nil, err
	}

	if mr.Method == "error" {
		return nil, fmt.Errorf("JMAP error: %s", string(mr.Args))
	}

	var emailResp EmailGetResponse
	if err := json.Unmarshal(mr.Args, &emailResp); err != nil {
		return nil, err
	}

	return emailResp.List, nil
}

// GetThread fetches a thread and all its emails
func (c *Client) GetThread(threadID string) ([]Email, error) {
	// First get the thread to get email IDs
	calls := []Invocation{
		NewInvocation("Thread/get", map[string]interface{}{
			"accountId": c.accountID,
			"ids":       []string{threadID},
		}, "0"),
	}

	resp, err := c.Call(calls)
	if err != nil {
		return nil, err
	}

	mr, err := ParseMethodResponse(resp.MethodResponses[0])
	if err != nil {
		return nil, err
	}

	var threadResp ThreadGetResponse
	if err := json.Unmarshal(mr.Args, &threadResp); err != nil {
		return nil, err
	}

	if len(threadResp.List) == 0 {
		return nil, fmt.Errorf("thread not found")
	}

	thread := threadResp.List[0]
	if f, err := os.OpenFile("debug.log", os.O_APPEND|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintf(f, "Thread %s has %d emailIds: %v\n", thread.ID, len(thread.EmailIDs), thread.EmailIDs)
		f.Close()
	}

	// Now get all emails in the thread with full body
	calls = []Invocation{
		NewInvocation("Email/get", map[string]interface{}{
			"accountId": c.accountID,
			"ids":       thread.EmailIDs,
			"properties": []string{
				"id", "threadId", "from", "to", "cc",
				"subject", "receivedAt", "preview",
				"textBody", "htmlBody", "bodyValues",
				"attachments", "hasAttachment",
				"messageId", "inReplyTo", "references",
			},
			"fetchTextBodyValues": true,
			"fetchHTMLBodyValues": true,
		}, "0"),
	}

	resp, err = c.Call(calls)
	if err != nil {
		return nil, err
	}

	mr, err = ParseMethodResponse(resp.MethodResponses[0])
	if err != nil {
		return nil, err
	}

	var emailResp EmailGetResponse
	if err := json.Unmarshal(mr.Args, &emailResp); err != nil {
		return nil, err
	}

	return emailResp.List, nil
}

// getInboxID returns the inbox mailbox ID
func (c *Client) getInboxID() (string, error) {
	calls := []Invocation{
		NewInvocation("Mailbox/get", map[string]interface{}{
			"accountId": c.accountID,
		}, "0"),
	}

	resp, err := c.Call(calls)
	if err != nil {
		return "", err
	}

	mr, err := ParseMethodResponse(resp.MethodResponses[0])
	if err != nil {
		return "", err
	}

	var mailboxResp MailboxGetResponse
	if err := json.Unmarshal(mr.Args, &mailboxResp); err != nil {
		return "", err
	}

	for _, mb := range mailboxResp.List {
		if mb.Role == "inbox" {
			return mb.ID, nil
		}
	}

	return "", fmt.Errorf("inbox not found")
}

// GetEmails fetches emails by ID with full body content
func (c *Client) GetEmails(ids []string) ([]Email, error) {
	calls := []Invocation{
		NewInvocation("Email/get", map[string]interface{}{
			"accountId": c.accountID,
			"ids":       ids,
			"properties": []string{
				"id", "threadId", "from", "to", "cc",
				"subject", "receivedAt", "preview",
				"textBody", "htmlBody", "bodyValues",
				"attachments", "hasAttachment",
				"messageId", "inReplyTo", "references",
			},
			"fetchTextBodyValues": true,
			"fetchHTMLBodyValues": true,
		}, "0"),
	}

	resp, err := c.Call(calls)
	if err != nil {
		return nil, err
	}

	mr, err := ParseMethodResponse(resp.MethodResponses[0])
	if err != nil {
		return nil, err
	}

	var emailResp EmailGetResponse
	if err := json.Unmarshal(mr.Args, &emailResp); err != nil {
		return nil, err
	}

	// Sort by received date (oldest first)
	sort.Slice(emailResp.List, func(i, j int) bool {
		return emailResp.List[i].ReceivedAt < emailResp.List[j].ReceivedAt
	})

	return emailResp.List, nil
}

// GetEmailBody returns the body text for an email
func (e *Email) GetBodyText() string {
	// Prefer text body
	for _, part := range e.TextBody {
		if val, ok := e.BodyValues[part.PartID]; ok {
			return val.Value
		}
	}

	// Fall back to HTML body
	for _, part := range e.HTMLBody {
		if val, ok := e.BodyValues[part.PartID]; ok {
			return val.Value
		}
	}

	return e.Preview
}
