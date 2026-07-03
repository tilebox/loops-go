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
	"path"
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

// CreateContact adds a contact to your audience.
//
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

// UpdateContact updates a contact by `email` or `userId`. You must provide one of these parameters.
//
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

// FindContact searches for a contact by `email` or `userId`. Only one parameter is allowed.
//
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

// DeleteContact deletes a contact by `email` or `userId`.
//
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

// GetMailingLists retrieves a list of your account's mailing lists.
//
// See: https://loops.so/docs/api-reference/list-mailing-lists
func (c *Client) GetMailingLists(ctx context.Context) ([]*MailingList, error) {
	req, err := newGetRequestWithQueryParams(c, ctx, "/lists", nil)
	if err != nil {
		return nil, err
	}

	return sendRequest[[]*MailingList](c, req)
}

// SendEvent sends events to trigger emails in Loops.
//
// See: https://loops.so/docs/api-reference/send-event
func (c *Client) SendEvent(ctx context.Context, event *Event) error {
	return c.SendEventWithIdempotencyKey(ctx, event, "")
}

// SendEventWithIdempotencyKey sends events to trigger emails in Loops with an optional idempotency key.
//
// See: https://loops.so/docs/api-reference/send-event
func (c *Client) SendEventWithIdempotencyKey(ctx context.Context, event *Event, idempotencyKey string) error {
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
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}
	_, err = sendRequest[*MessageResponse](c, req)
	return err
}

// SendTransactionalEmail sends a transactional email to a contact.
//
// See: https://loops.so/docs/api-reference/send-transactional-email
func (c *Client) SendTransactionalEmail(ctx context.Context, transactional *TransactionalRequest) error {
	return c.SendTransactionalEmailWithIdempotencyKey(ctx, transactional, "")
}

// SendTransactionalEmailWithIdempotencyKey sends a transactional email to a contact with an optional idempotency key.
//
// See: https://loops.so/docs/api-reference/send-transactional-email
func (c *Client) SendTransactionalEmailWithIdempotencyKey(ctx context.Context, transactional *TransactionalRequest, idempotencyKey string) error {
	req, err := newRequestWithBody(c, ctx, http.MethodPost, "/transactional", transactional)
	if err != nil {
		return err
	}
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
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

// GetContactProperties retrieves a list of your account's contact properties.
// Use the `list` parameter to query "all" or "custom" properties.
//
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

// CreateContactProperty adds a contact property to your team.
//
// See: https://loops.so/docs/api-reference/create-contact-property
func (c *Client) CreateContactProperty(ctx context.Context, property *ContactPropertyCreate) error {
	req, err := newRequestWithBody(c, ctx, http.MethodPost, "/contacts/properties", property)
	if err != nil {
		return err
	}
	_, err = sendRequest[*SuccessResponse](c, req)
	return err
}

// GetCustomFields retrieves your team's custom contact properties.
//
// Deprecated: Use GetContactProperties instead.
// See: https://loops.so/docs/api-reference/list-contact-properties
func (c *Client) GetCustomFields(ctx context.Context) ([]*ContactProperty, error) {
	req, err := newGetRequestWithQueryParams(c, ctx, "/contacts/customFields", nil)
	if err != nil {
		return nil, err
	}
	return sendRequest[[]*ContactProperty](c, req)
}

// GetDedicatedSendingIPs retrieves a list of Loops' dedicated sending IP addresses.
//
// See: https://loops.so/docs/api-reference/dedicated-sending-ips
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

type ListOptions = ListTransactionalEmailsOptions

// ListTransactionalEmails gets a list of published transactional emails.
//
// See: https://loops.so/docs/api-reference/list-transactional-emails-v1
func (c *Client) ListTransactionalEmails(ctx context.Context, opts ListTransactionalEmailsOptions) (*TransactionalEmailList, error) {
	params := url.Values{}
	if err := addListOptions(params, opts); err != nil {
		return nil, err
	}
	req, err := newGetRequestWithQueryParams(c, ctx, "/transactional", params)
	if err != nil {
		return nil, err
	}
	return sendRequest[*TransactionalEmailList](c, req)
}

// GetContactSuppression retrieves suppression status and removal quota for a contact by `email` or `userId`.
// Include only one query parameter.
//
// See: https://loops.so/docs/api-reference/check-contact-suppression
func (c *Client) GetContactSuppression(ctx context.Context, contact *ContactIdentifier) (*ContactSuppressionStatusResponse, error) {
	params, err := contactIdentifierQueryParams(contact)
	if err != nil {
		return nil, err
	}
	req, err := newGetRequestWithQueryParams(c, ctx, "/contacts/suppression", params)
	if err != nil {
		return nil, err
	}
	return sendRequest[*ContactSuppressionStatusResponse](c, req)
}

// RemoveContactSuppression removes a suppressed contact from the suppression list by `email` or `userId`.
// Include only one query parameter.
//
// See: https://loops.so/docs/api-reference/remove-contact-suppression
func (c *Client) RemoveContactSuppression(ctx context.Context, contact *ContactIdentifier) (*ContactSuppressionRemoveResponse, error) {
	params, err := contactIdentifierQueryParams(contact)
	if err != nil {
		return nil, err
	}
	req, err := newRequestWithBody[Contact](c, ctx, http.MethodDelete, "/contacts/suppression", nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = params.Encode()
	return sendRequest[*ContactSuppressionRemoveResponse](c, req)
}

// ListTransactionalEmailResources retrieves a paginated list of transactional emails, most recently created first.
//
// See: https://loops.so/docs/api-reference/list-transactional-emails
func (c *Client) ListTransactionalEmailResources(ctx context.Context, opts ListOptions) (*ListTransactionalsResourceResponse, error) {
	return getList[ListTransactionalsResourceResponse](c, ctx, "/transactional-emails", opts)
}

// CreateTransactionalEmail creates a new transactional email.
// An empty draft email message is created automatically and its `draftEmailMessageId` is returned.
//
// See: https://loops.so/docs/api-reference/create-transactional-email
func (c *Client) CreateTransactionalEmail(ctx context.Context, request *CreateTransactionalRequest) (*TransactionalDraftResponse, error) {
	return post[CreateTransactionalRequest, TransactionalDraftResponse](c, ctx, "/transactional-emails", request)
}

// GetTransactionalEmail retrieves a single transactional email by ID.
//
// See: https://loops.so/docs/api-reference/get-transactional-email
func (c *Client) GetTransactionalEmail(ctx context.Context, transactionalID string) (*TransactionalResponse, error) {
	return get[TransactionalResponse](c, ctx, path.Join("/transactional-emails", transactionalID))
}

// UpdateTransactionalEmail updates a transactional email by ID.
//
// See: https://loops.so/docs/api-reference/update-transactional-email
func (c *Client) UpdateTransactionalEmail(ctx context.Context, transactionalID string, request *UpdateTransactionalRequest) (*TransactionalResponse, error) {
	return post[UpdateTransactionalRequest, TransactionalResponse](c, ctx, path.Join("/transactional-emails", transactionalID), request)
}

// EnsureTransactionalEmailDraft ensures the transactional email has a draft email message.
// If a draft already exists it is returned unchanged; otherwise a new empty draft is created.
//
// See: https://loops.so/docs/api-reference/ensure-transactional-draft
func (c *Client) EnsureTransactionalEmailDraft(ctx context.Context, transactionalID string) (*TransactionalDraftResponse, error) {
	return post[Contact, TransactionalDraftResponse](c, ctx, path.Join("/transactional-emails", transactionalID, "draft"), nil)
}

// PublishTransactionalEmailDraft publishes the transactional email's current draft email message.
// The draft becomes the published version and the draft is cleared.
//
// See: https://loops.so/docs/api-reference/publish-transactional-email
func (c *Client) PublishTransactionalEmailDraft(ctx context.Context, transactionalID string) (*TransactionalResponse, error) {
	return post[Contact, TransactionalResponse](c, ctx, path.Join("/transactional-emails", transactionalID, "publish"), nil)
}

// ListThemes retrieves a paginated list of email themes, most recently created first.
//
// See: https://loops.so/docs/api-reference/list-themes
func (c *Client) ListThemes(ctx context.Context, opts ListOptions) (*ListThemesResponse, error) {
	return getList[ListThemesResponse](c, ctx, "/themes", opts)
}

// CreateTheme creates a new email theme.
//
// See: https://loops.so/docs/api-reference/list-themes
func (c *Client) CreateTheme(ctx context.Context, request *CreateThemeBody) (*ThemeResponse, error) {
	return post[CreateThemeBody, ThemeResponse](c, ctx, "/themes", request)
}

// GetTheme retrieves a single theme by ID.
//
// See: https://loops.so/docs/api-reference/get-theme
func (c *Client) GetTheme(ctx context.Context, themeID string) (*ThemeResponse, error) {
	return get[ThemeResponse](c, ctx, path.Join("/themes", themeID))
}

// UpdateTheme updates a theme's name and/or styles.
// When `styles` change, the update cascades to every email using this theme.
//
// See: https://loops.so/docs/api-reference/get-theme
func (c *Client) UpdateTheme(ctx context.Context, themeID string, request *UpdateThemeBody) (*UpdateThemeResponse, error) {
	return post[UpdateThemeBody, UpdateThemeResponse](c, ctx, path.Join("/themes", themeID), request)
}

// ListComponents retrieves a paginated list of email components.
//
// See: https://loops.so/docs/api-reference/list-components
func (c *Client) ListComponents(ctx context.Context, opts ListOptions) (*ListComponentsResponse, error) {
	return getList[ListComponentsResponse](c, ctx, "/components", opts)
}

// CreateComponent creates a new email component from an LMX body.
//
// See: https://loops.so/docs/api-reference/list-components
func (c *Client) CreateComponent(ctx context.Context, request *CreateComponentBody) (*ComponentResponse, error) {
	return post[CreateComponentBody, ComponentResponse](c, ctx, "/components", request)
}

// GetComponent retrieves a single component by ID.
//
// See: https://loops.so/docs/api-reference/get-component
func (c *Client) GetComponent(ctx context.Context, componentID string) (*ComponentResponse, error) {
	return get[ComponentResponse](c, ctx, path.Join("/components", componentID))
}

// UpdateComponent updates a component's name and/or body.
// When the `lmx` body changes, the update cascades to every email using this component.
//
// See: https://loops.so/docs/api-reference/get-component
func (c *Client) UpdateComponent(ctx context.Context, componentID string, request *UpdateComponentBody) (*UpdateComponentResponse, error) {
	return post[UpdateComponentBody, UpdateComponentResponse](c, ctx, path.Join("/components", componentID), request)
}

// ListAudienceSegments retrieves a paginated list of audience segments, most recently created first.
//
// See: https://loops.so/docs/api-reference/list-audience-segments
func (c *Client) ListAudienceSegments(ctx context.Context, opts ListOptions) (*ListAudienceSegmentsResponse, error) {
	return getList[ListAudienceSegmentsResponse](c, ctx, "/audience-segments", opts)
}

// GetAudienceSegment retrieves a single audience segment by ID.
//
// See: https://loops.so/docs/api-reference/get-audience-segment
func (c *Client) GetAudienceSegment(ctx context.Context, audienceSegmentID string) (*AudienceSegmentResponse, error) {
	return get[AudienceSegmentResponse](c, ctx, path.Join("/audience-segments", audienceSegmentID))
}

// ListCampaigns retrieves a paginated list of campaigns.
//
// See: https://loops.so/docs/api-reference/list-campaigns
func (c *Client) ListCampaigns(ctx context.Context, opts ListOptions) (*ListCampaignsResponse, error) {
	return getList[ListCampaignsResponse](c, ctx, "/campaigns", opts)
}

// CreateCampaign creates a new draft campaign.
// An empty email message is created automatically and its `emailMessageId` is returned.
//
// See: https://loops.so/docs/api-reference/create-campaign
func (c *Client) CreateCampaign(ctx context.Context, request *CreateCampaignRequest) (*CreateCampaignResponse, error) {
	return post[CreateCampaignRequest, CreateCampaignResponse](c, ctx, "/campaigns", request)
}

// GetCampaign retrieves a single campaign by ID.
//
// See: https://loops.so/docs/api-reference/get-campaign
func (c *Client) GetCampaign(ctx context.Context, campaignID string) (*CampaignResponse, error) {
	return get[CampaignResponse](c, ctx, path.Join("/campaigns", campaignID))
}

// UpdateCampaign updates a draft campaign's name, group, audience, or scheduling.
// Campaigns can only be updated while in draft status.
//
// See: https://loops.so/docs/api-reference/update-campaign
func (c *Client) UpdateCampaign(ctx context.Context, campaignID string, request *UpdateCampaignRequest) (*CampaignResponse, error) {
	return post[UpdateCampaignRequest, CampaignResponse](c, ctx, path.Join("/campaigns", campaignID), request)
}

// GetEmailMessage retrieves an email message, including its compiled LMX content.
//
// See: https://loops.so/docs/api-reference/get-email-message
func (c *Client) GetEmailMessage(ctx context.Context, emailMessageID string) (*EmailMessageResponse, error) {
	return get[EmailMessageResponse](c, ctx, path.Join("/email-messages", emailMessageID))
}

// UpdateEmailMessage updates fields on an email message.
// Supply `expectedRevisionId` matching the current `contentRevisionId`.
//
// See: https://loops.so/docs/api-reference/update-email-message
func (c *Client) UpdateEmailMessage(ctx context.Context, emailMessageID string, request *UpdateEmailMessageRequest) (*EmailMessageResponse, error) {
	return post[UpdateEmailMessageRequest, EmailMessageResponse](c, ctx, path.Join("/email-messages", emailMessageID), request)
}

// SendEmailMessagePreview sends a test preview of an email message to one or more addresses.
// The accepted variable fields depend on the parent type.
//
// See: https://loops.so/docs/api-reference/preview-email-message
func (c *Client) SendEmailMessagePreview(ctx context.Context, emailMessageID string, request *EmailMessagePreviewRequest) (*EmailMessagePreviewResponse, error) {
	return post[EmailMessagePreviewRequest, EmailMessagePreviewResponse](c, ctx, path.Join("/email-messages", emailMessageID, "preview"), request)
}

// GetEmailMessageGuardian validates an email message's content against Guardian rules.
// Errors must be resolved before the email can be published; warnings are advisory.
//
// See: https://loops.so/docs/api-reference/run-guardian-checks
func (c *Client) GetEmailMessageGuardian(ctx context.Context, emailMessageID string) (*EmailMessageGuardianResponse, error) {
	return get[EmailMessageGuardianResponse](c, ctx, path.Join("/email-messages", emailMessageID, "guardian"))
}

// ListWorkflows retrieves a paginated list of workflows.
//
// See: https://loops.so/docs/api-reference/list-workflows
func (c *Client) ListWorkflows(ctx context.Context, opts ListOptions) (*ListWorkflowsResponse, error) {
	return getList[ListWorkflowsResponse](c, ctx, "/workflows", opts)
}

// GetWorkflow retrieves a workflow graph with node type names, connections, and selected display fields.
//
// See: https://loops.so/docs/api-reference/get-workflow
func (c *Client) GetWorkflow(ctx context.Context, workflowID string) (*SimplifiedWorkflow, error) {
	return get[SimplifiedWorkflow](c, ctx, path.Join("/workflows", workflowID))
}

// GetWorkflowNode retrieves detailed data for a single workflow node.
//
// See: https://loops.so/docs/api-reference/get-workflow-node
func (c *Client) GetWorkflowNode(ctx context.Context, workflowID, nodeID string) (*WorkflowNode, error) {
	return get[WorkflowNode](c, ctx, path.Join("/workflows", workflowID, "nodes", nodeID))
}

// ListCampaignGroups retrieves a paginated list of campaign groups, most recently created first.
//
// See: https://loops.so/docs/api-reference/list-campaign-groups
func (c *Client) ListCampaignGroups(ctx context.Context, opts ListOptions) (*ListGroupsResponse, error) {
	return getList[ListGroupsResponse](c, ctx, "/campaign-groups", opts)
}

// CreateCampaignGroup creates a new campaign group.
//
// See: https://loops.so/docs/api-reference/create-campaign-group
func (c *Client) CreateCampaignGroup(ctx context.Context, request *CreateGroupRequest) (*GroupResponse, error) {
	return post[CreateGroupRequest, GroupResponse](c, ctx, "/campaign-groups", request)
}

// GetCampaignGroup retrieves a single campaign group by ID.
//
// See: https://loops.so/docs/api-reference/get-campaign-group
func (c *Client) GetCampaignGroup(ctx context.Context, campaignGroupID string) (*GroupResponse, error) {
	return get[GroupResponse](c, ctx, path.Join("/campaign-groups", campaignGroupID))
}

// UpdateCampaignGroup updates a campaign group's name or description.
// At least one field must be provided.
//
// See: https://loops.so/docs/api-reference/update-campaign-group
func (c *Client) UpdateCampaignGroup(ctx context.Context, campaignGroupID string, request *UpdateGroupRequest) (*GroupResponse, error) {
	return post[UpdateGroupRequest, GroupResponse](c, ctx, path.Join("/campaign-groups", campaignGroupID), request)
}

// ListTransactionalGroups retrieves a paginated list of transactional groups, most recently created first.
//
// See: https://loops.so/docs/api-reference/list-transactional-groups
func (c *Client) ListTransactionalGroups(ctx context.Context, opts ListOptions) (*ListGroupsResponse, error) {
	return getList[ListGroupsResponse](c, ctx, "/transactional-groups", opts)
}

// CreateTransactionalGroup creates a new transactional group.
//
// See: https://loops.so/docs/api-reference/create-transactional-group
func (c *Client) CreateTransactionalGroup(ctx context.Context, request *CreateGroupRequest) (*GroupResponse, error) {
	return post[CreateGroupRequest, GroupResponse](c, ctx, "/transactional-groups", request)
}

// GetTransactionalGroup retrieves a single transactional group by ID.
//
// See: https://loops.so/docs/api-reference/get-transactional-group
func (c *Client) GetTransactionalGroup(ctx context.Context, transactionalGroupID string) (*GroupResponse, error) {
	return get[GroupResponse](c, ctx, path.Join("/transactional-groups", transactionalGroupID))
}

// UpdateTransactionalGroup updates a transactional group's name or description.
// At least one field must be provided.
//
// See: https://loops.so/docs/api-reference/update-transactional-group
func (c *Client) UpdateTransactionalGroup(ctx context.Context, transactionalGroupID string, request *UpdateGroupRequest) (*GroupResponse, error) {
	return post[UpdateGroupRequest, GroupResponse](c, ctx, path.Join("/transactional-groups", transactionalGroupID), request)
}

// CreateUpload requests a pre-signed URL to upload an image asset.
// Upload the file with an HTTP PUT to the returned `presignedUrl`, then call CompleteUpload.
//
// See: https://loops.so/docs/api-reference/create-upload
func (c *Client) CreateUpload(ctx context.Context, request *CreateUploadRequest) (*CreateUploadResponse, error) {
	return post[CreateUploadRequest, CreateUploadResponse](c, ctx, "/uploads", request)
}

// CompleteUpload finalizes an asset after the file has been uploaded to the pre-signed URL.
// It returns the public URL of the uploaded asset.
//
// See: https://loops.so/docs/api-reference/complete-upload
func (c *Client) CompleteUpload(ctx context.Context, id string) (*CompleteUploadResponse, error) {
	return post[Contact, CompleteUploadResponse](c, ctx, path.Join("/uploads", id, "complete"), nil)
}

// TestAPIKey tests that an API key is valid.
//
// See: https://loops.so/docs/api-reference/api-key
func (c *Client) TestAPIKey(ctx context.Context) (*APIKeyInfo, error) {
	req, err := newGetRequestWithQueryParams(c, ctx, "/api-key", nil)
	if err != nil {
		return nil, err
	}

	return sendRequest[*APIKeyInfo](c, req)
}

func contactIdentifierQueryParams(contact *ContactIdentifier) (url.Values, error) {
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
	return params, nil
}

func addListOptions(params url.Values, opts ListOptions) error {
	if opts.PerPage != 0 {
		if opts.PerPage < 10 || opts.PerPage > 50 {
			return errors.New("perPage must be between 10 and 50 (inclusive)")
		}
		params.Add("perPage", strconv.Itoa(opts.PerPage))
	}
	if opts.Cursor != "" {
		params.Add("cursor", opts.Cursor)
	}
	return nil
}

func getList[T any](c *Client, ctx context.Context, requestPath string, opts ListOptions) (*T, error) {
	params := url.Values{}
	if err := addListOptions(params, opts); err != nil {
		return nil, err
	}
	req, err := newGetRequestWithQueryParams(c, ctx, requestPath, params)
	if err != nil {
		return nil, err
	}
	return sendRequest[*T](c, req)
}

func get[T any](c *Client, ctx context.Context, requestPath string) (*T, error) {
	req, err := newGetRequestWithQueryParams(c, ctx, requestPath, nil)
	if err != nil {
		return nil, err
	}
	return sendRequest[*T](c, req)
}

func post[Request any, Response any](c *Client, ctx context.Context, requestPath string, request *Request) (*Response, error) {
	req, err := newRequestWithBody(c, ctx, http.MethodPost, requestPath, request)
	if err != nil {
		return nil, err
	}
	return sendRequest[*Response](c, req)
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
	if err == nil && errorMsg.Message != "" {
		return none, errors.New(errorMsg.Message)
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
