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
		CustomProperties: map[string]interface{}{
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
	assert.Equal(t, "blue", c.CustomProperties["favoriteColor"])
	assert.Equal(t, "John", *c.FirstName)
	assert.Equal(t, "Doe", *c.LastName)
	require.Len(t, c.MailingLists, 1)
	list123, ok := c.MailingLists["list_123"]
	assert.True(t, ok)
	assert.True(t, list123)
}
