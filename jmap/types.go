package jmap

import "encoding/json"

// Session represents the JMAP session response
type Session struct {
	Accounts       map[string]Account `json:"accounts"`
	PrimaryAccount map[string]string  `json:"primaryAccounts"`
	APIURL         string             `json:"apiUrl"`
	DownloadURL    string             `json:"downloadUrl"`
	Username       string             `json:"username"`
}

// Account represents a JMAP account
type Account struct {
	Name string `json:"name"`
}

// Request represents a JMAP request
type Request struct {
	Using       []string     `json:"using"`
	MethodCalls []Invocation `json:"methodCalls"`
}

// Invocation represents a JMAP method call
type Invocation [3]interface{}

// NewInvocation creates a new JMAP method invocation
func NewInvocation(method string, args map[string]interface{}, callID string) Invocation {
	return Invocation{method, args, callID}
}

// Response represents a JMAP response
type Response struct {
	MethodResponses []json.RawMessage `json:"methodResponses"`
	SessionState    string            `json:"sessionState"`
}

// MethodResponse represents a parsed method response
type MethodResponse struct {
	Method string
	Args   json.RawMessage
	CallID string
}

// ParseMethodResponse parses a raw method response
func ParseMethodResponse(raw json.RawMessage) (*MethodResponse, error) {
	var arr []json.RawMessage
	if err := json.Unmarshal(raw, &arr); err != nil {
		return nil, err
	}

	var method, callID string
	if err := json.Unmarshal(arr[0], &method); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(arr[2], &callID); err != nil {
		return nil, err
	}

	return &MethodResponse{
		Method: method,
		Args:   arr[1],
		CallID: callID,
	}, nil
}

// Email represents a JMAP email object
type Email struct {
	ID            string               `json:"id"`
	ThreadID      string               `json:"threadId"`
	MailboxIDs    map[string]bool      `json:"mailboxIds"`
	From          []EmailAddress       `json:"from"`
	To            []EmailAddress       `json:"to"`
	CC            []EmailAddress       `json:"cc"`
	Subject       string               `json:"subject"`
	ReceivedAt    string               `json:"receivedAt"`
	Preview       string               `json:"preview"`
	TextBody      []BodyPart           `json:"textBody"`
	HTMLBody      []BodyPart           `json:"htmlBody"`
	BodyValues    map[string]BodyValue `json:"bodyValues"`
	Attachments   []Attachment         `json:"attachments"`
	HasAttachment bool                 `json:"hasAttachment"`
	MessageID     []string             `json:"messageId"`
	InReplyTo     []string             `json:"inReplyTo"`
	References    []string             `json:"references"`
}

// Attachment represents an email attachment
type Attachment struct {
	BlobID   string `json:"blobId"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Size     uint64 `json:"size"`
	CID      string `json:"cid"`
	IsInline bool   `json:"isInline"`
}

// EmailAddress represents an email address with optional name
type EmailAddress struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// String returns a formatted email address
func (e EmailAddress) String() string {
	if e.Name != "" {
		return e.Name + " <" + e.Email + ">"
	}
	return e.Email
}

// BodyPart represents a body part reference
type BodyPart struct {
	PartID string `json:"partId"`
	Type   string `json:"type"`
}

// BodyValue represents the actual body content
type BodyValue struct {
	Value string `json:"value"`
}

// Thread represents a JMAP thread object
type Thread struct {
	ID       string   `json:"id"`
	EmailIDs []string `json:"emailIds"`
}

// Mailbox represents a JMAP mailbox object
type Mailbox struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// EmailQueryResponse represents the response from Email/query
type EmailQueryResponse struct {
	AccountID string   `json:"accountId"`
	IDs       []string `json:"ids"`
	Position  int      `json:"position"`
	Total     int      `json:"total"`
}

// EmailGetResponse represents the response from Email/get
type EmailGetResponse struct {
	AccountID string   `json:"accountId"`
	State     string   `json:"state"`
	List      []Email  `json:"list"`
	NotFound  []string `json:"notFound"`
}

// ThreadGetResponse represents the response from Thread/get
type ThreadGetResponse struct {
	AccountID string   `json:"accountId"`
	State     string   `json:"state"`
	List      []Thread `json:"list"`
	NotFound  []string `json:"notFound"`
}

// MailboxGetResponse represents the response from Mailbox/get
type MailboxGetResponse struct {
	AccountID string    `json:"accountId"`
	State     string    `json:"state"`
	List      []Mailbox `json:"list"`
}
