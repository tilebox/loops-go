package loops

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

const defaultURL = "https://app.loops.so/api/v1/"

var ErrContactNotFound = errors.New("contact not found")

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type RequestInterceptor func(ctx context.Context, req *http.Request) error

type Client struct {
	apiURL              *url.URL
	httpClient          HTTPClient
	requestInterceptors []RequestInterceptor
}

// NewClient creates a new Loops client.
func NewClient(opts ...ClientOption) (*Client, error) {
	config := clientConfig{
		apiURL:     defaultURL,
		httpClient: http.DefaultClient,
	}
	for _, o := range opts {
		o(&config)
	}
	apiURL, err := url.Parse(config.apiURL)
	if err != nil {
		return nil, fmt.Errorf("invalid api url: %w", err)
	}

	requestInterceptors := config.requestInterceptors

	if config.apiKey != "" {
		requestInterceptors = append(requestInterceptors, func(ctx context.Context, req *http.Request) error {
			bearerToken := fmt.Sprintf("Bearer %s", config.apiKey)
			req.Header.Set("Authorization", bearerToken)
			return nil
		})
	}

	requestInterceptors = append(requestInterceptors, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Content-Type", "application/json")
		return nil
	})

	return &Client{
		apiURL:              apiURL,
		httpClient:          config.httpClient,
		requestInterceptors: requestInterceptors,
	}, nil
}

type clientConfig struct {
	apiURL              string
	apiKey              string
	httpClient          HTTPClient
	requestInterceptors []RequestInterceptor
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*clientConfig)

// WithURL allows overriding the default API URL (default: https://app.loops.so/api/v1/)
func WithURL(apiURL string) ClientOption {
	return func(c *clientConfig) {
		c.apiURL = apiURL
	}
}

// WithAPIKey sets the loops API key to use
func WithAPIKey(apiKey string) ClientOption {
	return func(c *clientConfig) {
		c.apiKey = apiKey
	}
}

// WithHTTPClient allows overriding the default http client, in case you want to use a custom one (e.g. retryablehttp)
// or for testing purposes (e.g. mocking)
func WithHTTPClient(httpClient HTTPClient) ClientOption {
	return func(c *clientConfig) {
		c.httpClient = httpClient
	}
}

// WithRequestInterceptors allows adding custom request interceptors, modifying API requests before they are sent
func WithRequestInterceptors(requestInterceptors ...RequestInterceptor) ClientOption {
	return func(c *clientConfig) {
		c.requestInterceptors = append(c.requestInterceptors, requestInterceptors...)
	}
}

// CreateContact creates a new contact with an email address and any other contact properties.
// See: https://loops.so/docs/api-reference/create-contact
func (c *Client) CreateContact(ctx context.Context, contact *Contact) (string, error) {
	req, err := newRequestWithBody(c, ctx, http.MethodPost, "/contacts/create", contact)
	if err != nil {
		return "", err
	}

	response, err := sendRequest[*IDResponse](c, req)
	if err != nil {
		return "", err
	}
	return response.ID, err
}

// UpdateContact updates or creates a contact.
// See: https://loops.so/docs/api-reference/update-contact
func (c *Client) UpdateContact(ctx context.Context, contact *Contact) (string, error) {
	req, err := newRequestWithBody(c, ctx, http.MethodPut, "/contacts/update", contact)
	if err != nil {
		return "", err
	}

	response, err := sendRequest[*IDResponse](c, req)
	if err != nil {
		return "", err
	}
	return response.ID, err
}

// FindContact finds a contact by email or userId.
// See: https://loops.so/docs/api-reference/find-contact
func (c *Client) FindContact(ctx context.Context, contact *ContactIdentifier) (*Contact, error) {
	if contact.Email == nil && contact.UserID == nil {
		return nil, errors.New("contact identifier must contain either an email or a userId")
	}
	if contact.Email != nil && contact.UserID != nil {
		return nil, errors.New("contact identifier must contain either an email or a userId, but not both")
	}

	params := url.Values{}
	if contact.Email != nil {
		params.Add("email", *contact.Email)
	}
	if contact.UserID != nil {
		params.Add("userId", *contact.UserID)
	}
	req, err := newGetRequestWithQueryParams(c, ctx, "/contacts/find", params)
	if err != nil {
		return nil, err
	}
	contacts, err := sendRequest[[]*Contact](c, req)
	if err != nil {
		return nil, err
	}
	if len(contacts) == 0 {
		return nil, ErrContactNotFound
	}
	return contacts[0], nil
}

// DeleteContact deletes a contact by email or userId.
// See: https://loops.so/docs/api-reference/delete-contact
func (c *Client) DeleteContact(ctx context.Context, contact *ContactIdentifier) error {
	if contact.Email == nil && contact.UserID == nil {
		return errors.New("contact identifier must contain either an email or a userId")
	}
	if contact.Email != nil && contact.UserID != nil {
		return errors.New("contact identifier must contain either an email or a userId, but not both")
	}

	req, err := newRequestWithBody(c, ctx, http.MethodPost, "/contacts/delete", &contact)
	if err != nil {
		return err
	}
	_, err = sendRequest[*MessageResponse](c, req)
	return err
}

// GetMailingLists retrieves a list of an account’s mailing lists.
// See: https://loops.so/docs/api-reference/get-mailing-lists
func (c *Client) GetMailingLists(ctx context.Context) ([]*MailingList, error) {
	req, err := newGetRequestWithQueryParams(c, ctx, "/lists", nil)
	if err != nil {
		return nil, err
	}

	return sendRequest[[]*MailingList](c, req)
}

// SendEvent sends an event to trigger emails in Loops.
// See: https://loops.so/docs/api-reference/send-event
func (c *Client) SendEvent(ctx context.Context, event *Event) error {
	if event.Email == nil && event.UserID == nil {
		return errors.New("event must contain either an email or a userId")
	}
	if event.Email != nil && event.UserID != nil {
		return errors.New("event must contain either an email or a userId, but not both")
	}
	req, err := newRequestWithBody(c, ctx, http.MethodPost, "/events/send", event)
	if err != nil {
		return err
	}
	_, err = sendRequest[*MessageResponse](c, req)
	return err
}

// SendTransactionalEmail sends a transactional email to a contact.
// See: https://loops.so/docs/api-reference/send-transactional-email
func (c *Client) SendTransactionalEmail(ctx context.Context, transactional *TransactionalEmail) error {
	req, err := newRequestWithBody(c, ctx, http.MethodPost, "/transactional", transactional)
	if err != nil {
		return err
	}
	_, err = sendRequest[*MessageResponse](c, req)
	return err
}

type ContactPropertyType int

const (
	ContactPropertyTypeAll ContactPropertyType = iota
	ContactPropertyTypeCustom
)

type ContactPropertyListOptions struct {
	// Which contact properties to return (all or custom to only list your team's custom properties)
	List ContactPropertyType
}

// GetContactProperties retrieves a list of an account's contact properties.
// Use listType "all" (default) or "custom" to filter properties.
// See: https://loops.so/docs/api-reference/list-contact-properties
func (c *Client) GetContactProperties(ctx context.Context, opts ContactPropertyListOptions) ([]*ContactProperty, error) {
	params := url.Values{}
	if opts.List == ContactPropertyTypeCustom {
		params.Add("list", "custom")
	} else if opts.List != ContactPropertyTypeAll {
		return nil, errors.New("invalid list type")
	}
	req, err := newGetRequestWithQueryParams(c, ctx, "/contacts/properties", params)
	if err != nil {
		return nil, err
	}
	return sendRequest[[]*ContactProperty](c, req)
}

// CreateContactProperty creates a new contact property.
// See: https://loops.so/docs/api-reference/create-contact-property
func (c *Client) CreateContactProperty(ctx context.Context, property *ContactPropertyCreate) error {
	req, err := newRequestWithBody(c, ctx, http.MethodPost, "/contacts/properties", property)
	if err != nil {
		return err
	}
	_, err = sendRequest[*SuccessResponse](c, req)
	return err
}

// Deprecated: Use GetContactProperties instead.
func (c *Client) GetCustomFields(ctx context.Context) ([]*ContactProperty, error) {
	req, err := newGetRequestWithQueryParams(c, ctx, "/contacts/customFields", nil)
	if err != nil {
		return nil, err
	}
	return sendRequest[[]*ContactProperty](c, req)
}

// GetDedicatedSendingIPs retrieves a list of Loops' dedicated sending IP addresses.
// See: https://loops.so/docs/api-reference/list-dedicated-sending-ips
func (c *Client) GetDedicatedSendingIPs(ctx context.Context) ([]string, error) {
	req, err := newGetRequestWithQueryParams(c, ctx, "/dedicated-sending-ips", nil)
	if err != nil {
		return nil, err
	}
	return sendRequest[[]string](c, req)
}

type ListTransactionalEmailsOptions struct {
	// Number of results per page (10-50, default 20)
	PerPage int
	// Pagination cursor from previous response
	Cursor string
}

// ListTransactionalEmails retrieves a list of published transactional emails.
// perPage: number of results per page (10-50, default 20)
// cursor: pagination cursor from previous response
// See: https://loops.so/docs/api-reference/list-transactional-emails
func (c *Client) ListTransactionalEmails(ctx context.Context, opts ListTransactionalEmailsOptions) (*TransactionalEmailList, error) {
	params := url.Values{}
	if opts.PerPage != 0 {
		if opts.PerPage < 10 || opts.PerPage > 50 {
			return nil, errors.New("perPage must be between 10 and 50 (inclusive)")
		}
		params.Add("perPage", strconv.Itoa(opts.PerPage))
	}
	if opts.Cursor != "" {
		params.Add("cursor", opts.Cursor)
	}
	req, err := newGetRequestWithQueryParams(c, ctx, "/transactional", params)
	if err != nil {
		return nil, err
	}
	return sendRequest[*TransactionalEmailList](c, req)
}

// TestAPIKey tests that an API key is valid.
// See: https://loops.so/docs/api-reference/api-key
func (c *Client) TestAPIKey(ctx context.Context) (*APIKeyInfo, error) {
	req, err := newGetRequestWithQueryParams(c, ctx, "/api-key", nil)
	if err != nil {
		return nil, err
	}

	return sendRequest[*APIKeyInfo](c, req)
}

func newGetRequestWithQueryParams(c *Client, ctx context.Context, path string, queryParams url.Values) (*http.Request, error) {
	req, err := newRequestWithBody[Contact](c, ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	if queryParams != nil {
		req.URL.RawQuery = queryParams.Encode()
	}

	return req, nil
}

func newRequestWithBody[T any](c *Client, ctx context.Context, method, path string, message *T) (*http.Request, error) {
	if path[0] == '/' {
		path = "." + path
	}

	queryURL, err := c.apiURL.Parse(path)
	if err != nil {
		return nil, err
	}

	var body io.Reader
	if message != nil {
		buf, err := json.Marshal(message)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal message: %w", err)
		}
		body = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	for _, interceptor := range c.requestInterceptors {
		if err := interceptor(ctx, req); err != nil {
			return nil, err
		}
	}
	return req, nil
}

func sendRequest[T any](c *Client, req *http.Request) (T, error) {
	var none T
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return none, fmt.Errorf("failed to send request %s: %w", req.URL.String(), err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return none, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 300 { // success response
		var response T
		err = json.Unmarshal(body, &response)
		if err != nil {
			return none, fmt.Errorf("failed to unmarshal response body: %w", err)
		}
		return response, nil
	}

	// sometimes loops returns an "error": message, so check if that's the case and if so, return the error
	errorMsg := &errorResponse{}
	err = json.Unmarshal(body, &errorMsg)
	if err == nil && errorMsg.Error != "" {
		return none, errors.New(errorMsg.Error)
	}

	// error, get the message and return it
	msg := &MessageResponse{}
	err = json.Unmarshal(body, &msg)
	if err != nil {
		return none, fmt.Errorf("failed to unmarshal error message: %w", err)
	}
	if msg.Message == "" {
		return none, errors.New(string(body))
	}
	return none, errors.New(msg.Message)
}
