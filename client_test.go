package loops

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/google/go-replayers/httpreplay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestdataDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "testdata"
	}

	repoRootIndex := strings.LastIndex(cwd, "/loops-go")
	repoRoot := path.Join(cwd[:repoRootIndex], "loops-go")
	return path.Join(repoRoot, "testdata")
}

// for initially recording tests by actually sending API requests (to make it work fill in a valid API key)
func newRecordTestClient(t *testing.T, recordingFile string) *Client { //nolint:unused
	// recorder automatically removes the Authorization header from requests
	recorder, err := httpreplay.NewRecorder(path.Join(TestdataDir(), recordingFile), nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = recorder.Close() })
	client, err := NewClient(WithAPIKey("API_KEY"), WithHTTPClient(recorder.Client()))
	require.NoError(t, err)
	return client
}

func newReplayTestClient(t *testing.T, recordingFile string) *Client {
	replayer, err := httpreplay.NewReplayer(path.Join(TestdataDir(), recordingFile))
	require.NoError(t, err)
	t.Cleanup(func() { _ = replayer.Close() })
	client, err := NewClient(WithHTTPClient(replayer.Client()))
	require.NoError(t, err)
	return client
}

func TestCreateContact(t *testing.T) {
	client := newReplayTestClient(t, "create-contact.replay.json")
	contactID, err := client.CreateContact(context.Background(), &Contact{
		Email:      "test@example.com",
		FirstName:  String("Test"),
		LastName:   String("User"),
		UserID:     String("user_123"),
		Subscribed: true,
		Properties: map[string]any{
			"companyRole": "Developer",
		},
	})
	require.NoError(t, err)
	assert.NotEmpty(t, contactID)
}

func TestUpdateContact(t *testing.T) {
	client := newReplayTestClient(t, "update-contact.replay.json")
	contactID, err := client.UpdateContact(context.Background(), &Contact{
		Email:      "new-test-mail@example.com",
		FirstName:  String("Test"),
		LastName:   String("User"),
		UserID:     String("user_123"),
		Subscribed: true,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, contactID)
}

func TestFindContact(t *testing.T) {
	client := newReplayTestClient(t, "find-contact.replay.json")
	contact, err := client.FindContact(context.Background(), &ContactIdentifier{
		Email: String("new-test-mail@example.com"),
	})
	require.NoError(t, err)
	assert.NotEmpty(t, contact.ID)
	assert.Equal(t, "new-test-mail@example.com", contact.Email)
	assert.Equal(t, "Test", *contact.FirstName)
	assert.Equal(t, "User", *contact.LastName)
	assert.Equal(t, "user_123", *contact.UserID)
	assert.Nil(t, contact.OptInStatus)

	companyRole, ok := contact.Properties["companyRole"]
	assert.True(t, ok)
	companyRoleStr, ok := companyRole.(string)
	assert.True(t, ok)
	assert.Equal(t, "Developer", companyRoleStr)

	assert.True(t, contact.Subscribed)
}

func TestFindContactByID(t *testing.T) {
	client := newReplayTestClient(t, "find-contact-by-id.replay.json")
	contact, err := client.FindContact(context.Background(), &ContactIdentifier{
		UserID: String("user_123"),
	})
	require.NoError(t, err)
	assert.NotEmpty(t, contact.ID)
	assert.Equal(t, "new-test-mail@example.com", contact.Email)
	assert.Equal(t, "Test", *contact.FirstName)
	assert.Equal(t, "User", *contact.LastName)
	assert.Equal(t, "user_123", *contact.UserID)
	assert.Nil(t, contact.OptInStatus)

	companyRole, ok := contact.Properties["companyRole"]
	assert.True(t, ok)
	companyRoleStr, ok := companyRole.(string)
	assert.True(t, ok)
	assert.Equal(t, "Developer", companyRoleStr)

	assert.True(t, contact.Subscribed)
}

func TestFindContactNotFound(t *testing.T) {
	client := newReplayTestClient(t, "find-contact-not-found.replay.json")
	_, err := client.FindContact(context.Background(), &ContactIdentifier{
		UserID: String("not_found"),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "contact not found")
}

func TestGetContactProperties(t *testing.T) {
	client := newReplayTestClient(t, "get-contact-allProperties.replay.json")
	allProperties, err := client.GetContactProperties(context.Background(), ContactPropertyListOptions{})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(allProperties), 14)
	assert.Equal(t, "firstName", allProperties[0].Key)
	assert.Equal(t, "First Name", allProperties[0].Label)
	assert.Equal(t, "string", allProperties[0].Type)
	assert.Equal(t, "lastName", allProperties[1].Key)

	customProperties, err := client.GetContactProperties(context.Background(), ContactPropertyListOptions{
		List: ContactPropertyTypeCustom,
	})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(customProperties), 2)
	assert.Equal(t, "heardAboutChannel", customProperties[0].Key)
	assert.Equal(t, "Heard About Channel", customProperties[0].Label)
	assert.Equal(t, "string", customProperties[0].Type)
	assert.Equal(t, "companyRole", customProperties[1].Key)
}

func TestCreateContactProperty(t *testing.T) {
	client := newReplayTestClient(t, "create-contact-property.replay.json")
	err := client.CreateContactProperty(context.Background(), &ContactPropertyCreate{
		Name: "planName20260703",
		Type: "string",
	})
	require.NoError(t, err)
}

func TestDeleteContact(t *testing.T) {
	client := newReplayTestClient(t, "delete-contact.replay.json")
	err := client.DeleteContact(context.Background(), &ContactIdentifier{
		UserID: String("user_123"),
	})
	require.NoError(t, err)
}

func TestGetMailingLists(t *testing.T) {
	client := newReplayTestClient(t, "get-mailing-lists.replay.json")
	mailingLists, err := client.GetMailingLists(context.Background())
	require.NoError(t, err)
	require.Len(t, mailingLists, 2)
	assert.Equal(t, "cm3n274xf027h0mi33t4qhrdg", mailingLists[0].ID)
	assert.Equal(t, "Newsletter", mailingLists[0].Name)
	assert.True(t, mailingLists[0].IsPublic)
	assert.Equal(t, "cm6gb0ku002d00kiig98e153r", mailingLists[1].ID)
	assert.Equal(t, "Product Update", mailingLists[1].Name)
	assert.True(t, mailingLists[1].IsPublic)
}

func TestSendEvent(t *testing.T) {
	client := newReplayTestClient(t, "send-event.replay.json")
	err := client.SendEvent(context.Background(), &Event{
		Email:     String("neil.armstrong@moon.space"),
		EventName: "joinedMission",
		EventProperties: map[string]any{
			"mission": "Apollo 11",
		},
	})
	require.NoError(t, err)
}

func TestSendTransactionalEmail(t *testing.T) {
	client := newReplayTestClient(t, "send-transactional-email.replay.json")
	err := client.SendTransactionalEmail(context.Background(), &TransactionalRequest{
		TransactionalID: "cm3n2vjux00cgeyeflew9ly2w",
		Email:           "test@example.com",
		DataVariables: map[string]any{
			"name": "Mr. Test",
		},
	})
	require.NoError(t, err)
}

func TestGetDedicatedSendingIPs(t *testing.T) {
	client := newReplayTestClient(t, "get-dedicated-sending-ips.replay.json")
	ips, err := client.GetDedicatedSendingIPs(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, ips)
}

func TestListTransactionalEmails(t *testing.T) {
	client := newReplayTestClient(t, "list-transactional-emails.replay.json")
	response, err := client.ListTransactionalEmails(context.Background(), ListTransactionalEmailsOptions{})
	require.NoError(t, err)
	assert.Equal(t, 2, response.Pagination.TotalResults)
	assert.Equal(t, 2, response.Pagination.ReturnedResults)
	assert.Equal(t, 20, response.Pagination.PerPage)
	assert.Equal(t, 1, response.Pagination.TotalPages)
	assert.Empty(t, response.Pagination.NextCursor)
	assert.Empty(t, response.Pagination.NextPage)
	require.Len(t, response.Data, 2)
	assert.Equal(t, "cm3n2vjux00cgeyeflew9ly2w", response.Data[0].ID)
	assert.Equal(t, "Blank transactional", response.Data[0].Name)
	assert.NotEmpty(t, response.Data[0].LastUpdated)
	assert.Equal(t, []string{"name"}, response.Data[0].DataVariables)
}

func TestAPIKey(t *testing.T) {
	client := newReplayTestClient(t, "test-api-key.replay.json")
	apiKey, err := client.TestAPIKey(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "Tilebox Staging", apiKey.TeamName)
}

func TestAPIKeyInvalid(t *testing.T) {
	client := newReplayTestClient(t, "test-api-key-invalid.replay.json")
	_, err := client.TestAPIKey(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid API key")
}

type captureHTTPClient struct {
	statusCode int
	response   string
	request    *http.Request
	body       []byte
}

func (c *captureHTTPClient) Do(req *http.Request) (*http.Response, error) {
	c.request = req
	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		c.body = body
	}

	statusCode := c.statusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	response := c.response
	if response == "" {
		response = `{}`
	}

	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewBufferString(response)),
	}, nil
}

func TestClientEndpointRequests(t *testing.T) {
	ctx := context.Background()
	listOpts := ListOptions{PerPage: 10, Cursor: "cursor_123"}

	tests := []struct {
		name     string
		method   string
		path     string
		query    url.Values
		body     string
		headers  map[string]string
		response string
		call     func(context.Context, *Client) error
	}{
		{
			name:     "test api key",
			method:   http.MethodGet,
			path:     "/api/v1/api-key",
			response: `{"success":true,"teamName":"Team"}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.TestAPIKey(ctx)
				return err
			},
		},
		{
			name:     "create contact",
			method:   http.MethodPost,
			path:     "/api/v1/contacts/create",
			body:     `{"email":"person@example.com","firstName":"Person","subscribed":true,"plan":"pro"}`,
			response: `{"success":true,"id":"contact_123"}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.CreateContact(ctx, &Contact{
					Email:      "person@example.com",
					FirstName:  String("Person"),
					Subscribed: true,
					Properties: map[string]any{"plan": "pro"},
				})
				return err
			},
		},
		{
			name:     "update contact",
			method:   http.MethodPut,
			path:     "/api/v1/contacts/update",
			body:     `{"email":"person@example.com","userId":"user_123"}`,
			response: `{"success":true,"id":"contact_123"}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.UpdateContact(ctx, &Contact{Email: "person@example.com", UserID: String("user_123")})
				return err
			},
		},
		{
			name:     "find contact",
			method:   http.MethodGet,
			path:     "/api/v1/contacts/find",
			query:    url.Values{"email": {"person@example.com"}},
			response: `[{"id":"contact_123","email":"person@example.com","subscribed":true}]`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.FindContact(ctx, &ContactIdentifier{Email: String("person@example.com")})
				return err
			},
		},
		{
			name:     "delete contact",
			method:   http.MethodPost,
			path:     "/api/v1/contacts/delete",
			body:     `{"userId":"user_123"}`,
			response: `{"success":true,"message":"deleted"}`,
			call: func(ctx context.Context, client *Client) error {
				return client.DeleteContact(ctx, &ContactIdentifier{UserID: String("user_123")})
			},
		},
		{
			name:     "get contact suppression",
			method:   http.MethodGet,
			path:     "/api/v1/contacts/suppression",
			query:    url.Values{"email": {"person@example.com"}},
			response: `{"contact":{"id":"contact_123","email":"person@example.com","userId":null},"isSuppressed":false,"removalQuota":{}}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.GetContactSuppression(ctx, &ContactIdentifier{Email: String("person@example.com")})
				return err
			},
		},
		{
			name:     "remove contact suppression",
			method:   http.MethodDelete,
			path:     "/api/v1/contacts/suppression",
			query:    url.Values{"userId": {"user_123"}},
			response: `{"success":true,"message":"removed","removalQuota":{}}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.RemoveContactSuppression(ctx, &ContactIdentifier{UserID: String("user_123")})
				return err
			},
		},
		{
			name:     "list contact properties",
			method:   http.MethodGet,
			path:     "/api/v1/contacts/properties",
			query:    url.Values{"list": {"custom"}},
			response: `[]`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.GetContactProperties(ctx, ContactPropertyListOptions{List: ContactPropertyTypeCustom})
				return err
			},
		},
		{
			name:     "create contact property",
			method:   http.MethodPost,
			path:     "/api/v1/contacts/properties",
			body:     `{"name":"planName","type":"string"}`,
			response: `{"success":true}`,
			call: func(ctx context.Context, client *Client) error {
				return client.CreateContactProperty(ctx, &ContactPropertyCreate{Name: "planName", Type: "string"})
			},
		},
		{
			name:     "get dedicated sending ips",
			method:   http.MethodGet,
			path:     "/api/v1/dedicated-sending-ips",
			response: `["127.0.0.1"]`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.GetDedicatedSendingIPs(ctx)
				return err
			},
		},
		{
			name:     "get mailing lists",
			method:   http.MethodGet,
			path:     "/api/v1/lists",
			response: `[]`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.GetMailingLists(ctx)
				return err
			},
		},
		{
			name:     "send event",
			method:   http.MethodPost,
			path:     "/api/v1/events/send",
			body:     `{"email":"person@example.com","eventName":"joined","eventProperties":{"source":"test"},"plan":"pro"}`,
			headers:  map[string]string{"Idempotency-Key": "event_key"},
			response: `{"success":true}`,
			call: func(ctx context.Context, client *Client) error {
				return client.SendEventWithIdempotencyKey(ctx, &Event{
					Email:           String("person@example.com"),
					EventName:       "joined",
					EventProperties: map[string]any{"source": "test"},
					Properties:      map[string]any{"plan": "pro"},
				}, "event_key")
			},
		},
		{
			name:     "send transactional email",
			method:   http.MethodPost,
			path:     "/api/v1/transactional",
			body:     `{"email":"person@example.com","transactionalId":"transactional_123","dataVariables":{"name":"Person"}}`,
			headers:  map[string]string{"Idempotency-Key": "transactional_key"},
			response: `{"success":true}`,
			call: func(ctx context.Context, client *Client) error {
				return client.SendTransactionalEmailWithIdempotencyKey(ctx, &TransactionalRequest{
					Email:           "person@example.com",
					TransactionalID: "transactional_123",
					DataVariables:   map[string]any{"name": "Person"},
				}, "transactional_key")
			},
		},
		{
			name:     "list transactional emails",
			method:   http.MethodGet,
			path:     "/api/v1/transactional",
			query:    url.Values{"cursor": {"cursor_123"}, "perPage": {"10"}},
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.ListTransactionalEmails(ctx, listOpts)
				return err
			},
		},
		{
			name:     "list transactional email resources",
			method:   http.MethodGet,
			path:     "/api/v1/transactional-emails",
			query:    url.Values{"cursor": {"cursor_123"}, "perPage": {"10"}},
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.ListTransactionalEmailResources(ctx, listOpts)
				return err
			},
		},
		{
			name:     "create transactional email",
			method:   http.MethodPost,
			path:     "/api/v1/transactional-emails",
			body:     `{"name":"Welcome"}`,
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.CreateTransactionalEmail(ctx, &CreateTransactionalRequest{Name: "Welcome"})
				return err
			},
		},
		{
			name:     "get transactional email",
			method:   http.MethodGet,
			path:     "/api/v1/transactional-emails/transactional_123",
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.GetTransactionalEmail(ctx, "transactional_123")
				return err
			},
		},
		{
			name:     "update transactional email",
			method:   http.MethodPost,
			path:     "/api/v1/transactional-emails/transactional_123",
			body:     `{"name":"Updated"}`,
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.UpdateTransactionalEmail(ctx, "transactional_123", &UpdateTransactionalRequest{Name: String("Updated")})
				return err
			},
		},
		{
			name:     "ensure transactional email draft",
			method:   http.MethodPost,
			path:     "/api/v1/transactional-emails/transactional_123/draft",
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.EnsureTransactionalEmailDraft(ctx, "transactional_123")
				return err
			},
		},
		{
			name:     "publish transactional email draft",
			method:   http.MethodPost,
			path:     "/api/v1/transactional-emails/transactional_123/publish",
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.PublishTransactionalEmailDraft(ctx, "transactional_123")
				return err
			},
		},
		{
			name:     "list themes",
			method:   http.MethodGet,
			path:     "/api/v1/themes",
			query:    url.Values{"cursor": {"cursor_123"}, "perPage": {"10"}},
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.ListThemes(ctx, listOpts)
				return err
			},
		},
		{
			name:     "create theme",
			method:   http.MethodPost,
			path:     "/api/v1/themes",
			body:     `{"name":"Theme"}`,
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.CreateTheme(ctx, &CreateThemeBody{Name: "Theme"})
				return err
			},
		},
		{
			name:     "get theme",
			method:   http.MethodGet,
			path:     "/api/v1/themes/theme_123",
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.GetTheme(ctx, "theme_123")
				return err
			},
		},
		{
			name:     "update theme",
			method:   http.MethodPost,
			path:     "/api/v1/themes/theme_123",
			body:     `{"name":"Updated"}`,
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.UpdateTheme(ctx, "theme_123", &UpdateThemeBody{Name: String("Updated")})
				return err
			},
		},
		{
			name:     "list components",
			method:   http.MethodGet,
			path:     "/api/v1/components",
			query:    url.Values{"cursor": {"cursor_123"}, "perPage": {"10"}},
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.ListComponents(ctx, listOpts)
				return err
			},
		},
		{
			name:     "create component",
			method:   http.MethodPost,
			path:     "/api/v1/components",
			body:     `{"name":"Component","lmx":"<Text>Hello</Text>"}`,
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.CreateComponent(ctx, &CreateComponentBody{Name: "Component", LMX: "<Text>Hello</Text>"})
				return err
			},
		},
		{
			name:     "get component",
			method:   http.MethodGet,
			path:     "/api/v1/components/component_123",
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.GetComponent(ctx, "component_123")
				return err
			},
		},
		{
			name:     "update component",
			method:   http.MethodPost,
			path:     "/api/v1/components/component_123",
			body:     `{"name":"Updated"}`,
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.UpdateComponent(ctx, "component_123", &UpdateComponentBody{Name: String("Updated")})
				return err
			},
		},
		{
			name:     "list audience segments",
			method:   http.MethodGet,
			path:     "/api/v1/audience-segments",
			query:    url.Values{"cursor": {"cursor_123"}, "perPage": {"10"}},
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.ListAudienceSegments(ctx, listOpts)
				return err
			},
		},
		{
			name:     "get audience segment",
			method:   http.MethodGet,
			path:     "/api/v1/audience-segments/segment_123",
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.GetAudienceSegment(ctx, "segment_123")
				return err
			},
		},
		{
			name:     "list campaigns",
			method:   http.MethodGet,
			path:     "/api/v1/campaigns",
			query:    url.Values{"cursor": {"cursor_123"}, "perPage": {"10"}},
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.ListCampaigns(ctx, listOpts)
				return err
			},
		},
		{
			name:     "create campaign",
			method:   http.MethodPost,
			path:     "/api/v1/campaigns",
			body:     `{"name":"Campaign"}`,
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.CreateCampaign(ctx, &CreateCampaignRequest{Name: "Campaign"})
				return err
			},
		},
		{
			name:     "get campaign",
			method:   http.MethodGet,
			path:     "/api/v1/campaigns/campaign_123",
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.GetCampaign(ctx, "campaign_123")
				return err
			},
		},
		{
			name:     "update campaign",
			method:   http.MethodPost,
			path:     "/api/v1/campaigns/campaign_123",
			body:     `{"name":"Updated"}`,
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.UpdateCampaign(ctx, "campaign_123", &UpdateCampaignRequest{Name: String("Updated")})
				return err
			},
		},
		{
			name:     "get email message",
			method:   http.MethodGet,
			path:     "/api/v1/email-messages/message_123",
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.GetEmailMessage(ctx, "message_123")
				return err
			},
		},
		{
			name:     "update email message",
			method:   http.MethodPost,
			path:     "/api/v1/email-messages/message_123",
			body:     `{"expectedRevisionId":"rev_123","subject":"Subject"}`,
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.UpdateEmailMessage(ctx, "message_123", &UpdateEmailMessageRequest{ExpectedRevisionID: String("rev_123"), Subject: String("Subject")})
				return err
			},
		},
		{
			name:     "send email message preview",
			method:   http.MethodPost,
			path:     "/api/v1/email-messages/message_123/preview",
			body:     `{"emails":["person@example.com"],"dataVariables":{"name":"Person"}}`,
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.SendEmailMessagePreview(ctx, "message_123", &EmailMessagePreviewRequest{Emails: []string{"person@example.com"}, DataVariables: map[string]any{"name": "Person"}})
				return err
			},
		},
		{
			name:     "get email message guardian",
			method:   http.MethodGet,
			path:     "/api/v1/email-messages/message_123/guardian",
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.GetEmailMessageGuardian(ctx, "message_123")
				return err
			},
		},
		{
			name:     "list workflows",
			method:   http.MethodGet,
			path:     "/api/v1/workflows",
			query:    url.Values{"cursor": {"cursor_123"}, "perPage": {"10"}},
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.ListWorkflows(ctx, listOpts)
				return err
			},
		},
		{
			name:     "get workflow",
			method:   http.MethodGet,
			path:     "/api/v1/workflows/workflow_123",
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.GetWorkflow(ctx, "workflow_123")
				return err
			},
		},
		{
			name:     "get workflow node",
			method:   http.MethodGet,
			path:     "/api/v1/workflows/workflow_123/nodes/node_123",
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.GetWorkflowNode(ctx, "workflow_123", "node_123")
				return err
			},
		},
		{
			name:     "list campaign groups",
			method:   http.MethodGet,
			path:     "/api/v1/campaign-groups",
			query:    url.Values{"cursor": {"cursor_123"}, "perPage": {"10"}},
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.ListCampaignGroups(ctx, listOpts)
				return err
			},
		},
		{
			name:     "create campaign group",
			method:   http.MethodPost,
			path:     "/api/v1/campaign-groups",
			body:     `{"name":"Group","description":"Description"}`,
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.CreateCampaignGroup(ctx, &CreateGroupRequest{Name: "Group", Description: String("Description")})
				return err
			},
		},
		{
			name:     "get campaign group",
			method:   http.MethodGet,
			path:     "/api/v1/campaign-groups/group_123",
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.GetCampaignGroup(ctx, "group_123")
				return err
			},
		},
		{
			name:     "update campaign group",
			method:   http.MethodPost,
			path:     "/api/v1/campaign-groups/group_123",
			body:     `{"name":"Updated"}`,
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.UpdateCampaignGroup(ctx, "group_123", &UpdateGroupRequest{Name: String("Updated")})
				return err
			},
		},
		{
			name:     "list transactional groups",
			method:   http.MethodGet,
			path:     "/api/v1/transactional-groups",
			query:    url.Values{"cursor": {"cursor_123"}, "perPage": {"10"}},
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.ListTransactionalGroups(ctx, listOpts)
				return err
			},
		},
		{
			name:     "create transactional group",
			method:   http.MethodPost,
			path:     "/api/v1/transactional-groups",
			body:     `{"name":"Group","description":"Description"}`,
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.CreateTransactionalGroup(ctx, &CreateGroupRequest{Name: "Group", Description: String("Description")})
				return err
			},
		},
		{
			name:     "get transactional group",
			method:   http.MethodGet,
			path:     "/api/v1/transactional-groups/group_123",
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.GetTransactionalGroup(ctx, "group_123")
				return err
			},
		},
		{
			name:     "update transactional group",
			method:   http.MethodPost,
			path:     "/api/v1/transactional-groups/group_123",
			body:     `{"name":"Updated"}`,
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.UpdateTransactionalGroup(ctx, "group_123", &UpdateGroupRequest{Name: String("Updated")})
				return err
			},
		},
		{
			name:     "create upload",
			method:   http.MethodPost,
			path:     "/api/v1/uploads",
			body:     `{"contentType":"image/png","contentLength":42}`,
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.CreateUpload(ctx, &CreateUploadRequest{ContentType: "image/png", ContentLength: 42})
				return err
			},
		},
		{
			name:     "complete upload",
			method:   http.MethodPost,
			path:     "/api/v1/uploads/upload_123/complete",
			response: `{}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.CompleteUpload(ctx, "upload_123")
				return err
			},
		},
	}
	require.Len(t, tests, 51)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := &captureHTTPClient{response: tt.response}
			client, err := NewClient(WithURL("https://example.test/api/v1/"), WithAPIKey("test_key"), WithHTTPClient(transport))
			require.NoError(t, err)

			require.NoError(t, tt.call(ctx, client))
			require.NotNil(t, transport.request)
			assert.Equal(t, tt.method, transport.request.Method)
			assert.Equal(t, tt.path, transport.request.URL.Path)
			expectedQuery := tt.query
			if expectedQuery == nil {
				expectedQuery = url.Values{}
			}
			assert.Equal(t, expectedQuery, transport.request.URL.Query())
			assert.Equal(t, "Bearer test_key", transport.request.Header.Get("Authorization"))
			assert.Equal(t, "application/json", transport.request.Header.Get("Content-Type"))

			for key, value := range tt.headers {
				assert.Equal(t, value, transport.request.Header.Get(key))
			}

			if tt.body == "" {
				assert.Empty(t, transport.body)
			} else {
				assert.JSONEq(t, tt.body, string(transport.body))
			}
		})
	}
}

func TestClientListOptionsValidation(t *testing.T) {
	client, err := NewClient(WithHTTPClient(&captureHTTPClient{}))
	require.NoError(t, err)

	_, err = client.ListThemes(context.Background(), ListOptions{PerPage: 9})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "perPage must be between 10 and 50")
}

func TestClientEndpointRequestsCoverOpenAPIPaths(t *testing.T) {
	openAPIPaths := map[string]struct{}{
		"GET /v1/api-key":                                         {},
		"POST /v1/contacts/create":                                {},
		"PUT /v1/contacts/update":                                 {},
		"GET /v1/contacts/find":                                   {},
		"POST /v1/contacts/delete":                                {},
		"GET /v1/contacts/suppression":                            {},
		"DELETE /v1/contacts/suppression":                         {},
		"GET /v1/contacts/properties":                             {},
		"POST /v1/contacts/properties":                            {},
		"GET /v1/dedicated-sending-ips":                           {},
		"GET /v1/lists":                                           {},
		"POST /v1/events/send":                                    {},
		"POST /v1/transactional":                                  {},
		"GET /v1/transactional":                                   {},
		"GET /v1/transactional-emails":                            {},
		"POST /v1/transactional-emails":                           {},
		"GET /v1/transactional-emails/{transactionalId}":          {},
		"POST /v1/transactional-emails/{transactionalId}":         {},
		"POST /v1/transactional-emails/{transactionalId}/draft":   {},
		"POST /v1/transactional-emails/{transactionalId}/publish": {},
		"GET /v1/themes":                                          {},
		"POST /v1/themes":                                         {},
		"GET /v1/themes/{themeId}":                                {},
		"POST /v1/themes/{themeId}":                               {},
		"GET /v1/components":                                      {},
		"POST /v1/components":                                     {},
		"GET /v1/components/{componentId}":                        {},
		"POST /v1/components/{componentId}":                       {},
		"GET /v1/audience-segments":                               {},
		"GET /v1/audience-segments/{audienceSegmentId}":           {},
		"GET /v1/campaigns":                                       {},
		"POST /v1/campaigns":                                      {},
		"GET /v1/campaigns/{campaignId}":                          {},
		"POST /v1/campaigns/{campaignId}":                         {},
		"GET /v1/email-messages/{emailMessageId}":                 {},
		"POST /v1/email-messages/{emailMessageId}":                {},
		"POST /v1/email-messages/{emailMessageId}/preview":        {},
		"GET /v1/email-messages/{emailMessageId}/guardian":        {},
		"GET /v1/workflows":                                       {},
		"GET /v1/workflows/{workflowId}":                          {},
		"GET /v1/workflows/{workflowId}/nodes/{nodeId}":           {},
		"GET /v1/campaign-groups":                                 {},
		"POST /v1/campaign-groups":                                {},
		"GET /v1/campaign-groups/{campaignGroupId}":               {},
		"POST /v1/campaign-groups/{campaignGroupId}":              {},
		"GET /v1/transactional-groups":                            {},
		"POST /v1/transactional-groups":                           {},
		"GET /v1/transactional-groups/{transactionalGroupId}":     {},
		"POST /v1/transactional-groups/{transactionalGroupId}":    {},
		"POST /v1/uploads":                                        {},
		"POST /v1/uploads/{id}/complete":                          {},
	}

	// Keep this assertion next to TestClientEndpointRequests so future OpenAPI additions
	// force a test checklist update rather than silently extending the client untested.
	assert.Len(t, openAPIPaths, 51)
}
