package loops

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContactMarshalJSONCustomPropertiesInlined(t *testing.T) {
	c := Contact{
		ID:         "123",
		Email:      "test@example.com",
		Subscribed: true,
		MailingLists: map[string]bool{
			"list_123": true,
		},
		Properties: map[string]any{
			"favoriteColor": "blue",
		},
	}
	data, err := json.Marshal(&c)
	require.NoError(t, err)
	assert.JSONEq(t, `{"id":"123","email":"test@example.com","subscribed":true,"favoriteColor":"blue","mailingLists":{"list_123":true}}`, string(data))
}

func TestContactUnmarshalJSONCustomPropertiesInlined(t *testing.T) {
	c := Contact{}

	data := []byte(`{"id":"123","email":"test@example.com","subscribed":true,"favoriteColor":"blue","firstName":"John","lastName":"Doe","mailingLists":{"list_123":true}}`)
	err := json.Unmarshal(data, &c)
	require.NoError(t, err)
	assert.Equal(t, "123", c.ID)
	assert.Equal(t, "test@example.com", c.Email)
	assert.True(t, c.Subscribed)
	assert.Equal(t, "blue", c.Properties["favoriteColor"])
	assert.Equal(t, "John", *c.FirstName)
	assert.Equal(t, "Doe", *c.LastName)
	require.Len(t, c.MailingLists, 1)
	list123, ok := c.MailingLists["list_123"]
	assert.True(t, ok)
	assert.True(t, list123)
}

func TestEventMarshalJSONCustomPropertiesInlined(t *testing.T) {
	e := Event{
		Email:     String("test@example.com"),
		EventName: "joined",
		EventProperties: map[string]any{
			"source": "test",
		},
		MailingLists: map[string]bool{
			"list_123": true,
		},
		Properties: map[string]any{
			"plan": "pro",
		},
		ContactProperties: map[string]any{
			"companyRole": "Developer",
		},
	}

	data, err := json.Marshal(&e)
	require.NoError(t, err)
	assert.JSONEq(t, `{"email":"test@example.com","eventName":"joined","eventProperties":{"source":"test"},"mailingLists":{"list_123":true},"plan":"pro","companyRole":"Developer"}`, string(data))
}

func TestTransactionalRequestMarshalJSON(t *testing.T) {
	request := TransactionalRequest{
		Email:           "test@example.com",
		TransactionalID: "transactional_123",
		AddToAudience:   Bool(true),
		DataVariables: map[string]any{
			"name": "Test User",
		},
		Attachments: []EmailAttachment{
			{
				Filename:    "hello.txt",
				ContentType: "text/plain",
				Data:        "aGVsbG8=",
			},
		},
	}

	data, err := json.Marshal(&request)
	require.NoError(t, err)
	assert.JSONEq(t, `{"email":"test@example.com","transactionalId":"transactional_123","addToAudience":true,"dataVariables":{"name":"Test User"},"attachments":[{"filename":"hello.txt","contentType":"text/plain","data":"aGVsbG8="}]}`, string(data))
}

func TestOptionalObjectRequestFieldsOmittedWhenNil(t *testing.T) {
	campaign, err := json.Marshal(&CreateCampaignRequest{Name: "Campaign"})
	require.NoError(t, err)
	assert.JSONEq(t, `{"name":"Campaign"}`, string(campaign))

	updatedCampaign, err := json.Marshal(&UpdateCampaignRequest{Name: String("Updated")})
	require.NoError(t, err)
	assert.JSONEq(t, `{"name":"Updated"}`, string(updatedCampaign))

	theme, err := json.Marshal(&CreateThemeBody{Name: "Theme"})
	require.NoError(t, err)
	assert.JSONEq(t, `{"name":"Theme"}`, string(theme))

	updatedTheme, err := json.Marshal(&UpdateThemeBody{Name: String("Updated")})
	require.NoError(t, err)
	assert.JSONEq(t, `{"name":"Updated"}`, string(updatedTheme))
}

func TestUploadRequestMarshalJSON(t *testing.T) {
	request := CreateUploadRequest{
		ContentType:   "image/png",
		ContentLength: 42,
	}

	data, err := json.Marshal(&request)
	require.NoError(t, err)
	assert.JSONEq(t, `{"contentType":"image/png","contentLength":42}`, string(data))
}
