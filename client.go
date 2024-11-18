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
)

const defaultApiURL = "https://app.loops.so/api/v1/"

var ErrContactNotFound = errors.New("contact not found")

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type RequestInterceptor func(ctx context.Context, req *http.Request) error

type Client struct {
	apiURL              *url.URL
	httpClient          HttpClient
	requestInterceptors []RequestInterceptor
}

// NewClient creates a new Loops client.
func NewClient(opts ...ClientOption) (*Client, error) {
	config := clientConfig{
		apiURL:     defaultApiURL,
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
	httpClient          HttpClient
	requestInterceptors []RequestInterceptor
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*clientConfig)

// WithApiURL allows overriding the default API URL
func WithApiURL(apiURL string) ClientOption {
	return func(c *clientConfig) {
		c.apiURL = apiURL
	}
}

// WithApiKey sets the loops API key to use
func WithApiKey(apiKey string) ClientOption {
	return func(c *clientConfig) {
		c.apiKey = apiKey
	}
}

// WithHttpClient allows overriding the default http client, in case you want to use a custom one (e.g. retryablehttp)
// or for testing purposes (e.g. mocking)
func WithHttpClient(httpClient HttpClient) ClientOption {
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
	if contact.Email == nil && contact.UserId == nil {
		return nil, errors.New("contact identifier must contain either an email or a userId")
	}
	if contact.Email != nil && contact.UserId != nil {
		return nil, errors.New("contact identifier must contain either an email or a userId, but not both")
	}

	params := url.Values{}
	if contact.Email != nil {
		params.Add("email", *contact.Email)
	}
	if contact.UserId != nil {
		params.Add("userId", *contact.UserId)
	}
	req, err := newRequestWithQueryParams(c, ctx, http.MethodGet, "/contacts/find", params)
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
	if contact.Email == nil && contact.UserId == nil {
		return errors.New("contact identifier must contain either an email or a userId")
	}
	if contact.Email != nil && contact.UserId != nil {
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
	req, err := newRequestWithQueryParams(c, ctx, http.MethodGet, "/lists", nil)
	if err != nil {
		return nil, err
	}

	return sendRequest[[]*MailingList](c, req)
}

// SendEvent sends an event to trigger emails in Loops.
// See: https://loops.so/docs/api-reference/send-event
func (c *Client) SendEvent(ctx context.Context, event *Event) error {
	if event.Email == nil && event.UserId == nil {
		return errors.New("event must contain either an email or a userId")
	}
	if event.Email != nil && event.UserId != nil {
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

// GetCustomFields retrieves a list of an account's custom contact properties.
func (c *Client) GetCustomFields(ctx context.Context) ([]*CustomField, error) {
	req, err := newRequestWithQueryParams(c, ctx, http.MethodGet, "/contacts/customFields", nil)
	if err != nil {
		return nil, err
	}
	customFields, err := sendRequest[[]*CustomField](c, req)
	if err != nil {
		return nil, err
	}
	return customFields, nil
}

// TestApiKey tests that an API key is valid.
// See: https://loops.so/docs/api-reference/api-key
func (c *Client) TestApiKey(ctx context.Context) (*ApiKeyInfo, error) {
	req, err := newRequestWithQueryParams(c, ctx, http.MethodGet, "/api-key", nil)
	if err != nil {
		return nil, err
	}

	return sendRequest[*ApiKeyInfo](c, req)
}

func newRequestWithQueryParams(c *Client, ctx context.Context, method, path string, queryParams url.Values) (*http.Request, error) {
	req, err := newRequestWithBody[Contact](c, ctx, method, path, nil)
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
	if err == nil {
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
