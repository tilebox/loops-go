package loops

import (
	"context"
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
		CustomProperties: map[string]interface{}{
			"companyRole": "Developer",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "cm3n4kiua02c0t839btycnwe1", contactID)
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
	assert.Equal(t, "cm3n4kiua02c0t839btycnwe1", contactID)
}

func TestFindContact(t *testing.T) {
	client := newReplayTestClient(t, "find-contact.replay.json")
	contact, err := client.FindContact(context.Background(), &ContactIdentifier{
		Email: String("new-test-mail@example.com"),
	})
	require.NoError(t, err)
	assert.Equal(t, "cm3n4kiua02c0t839btycnwe1", contact.ID)
	assert.Equal(t, "new-test-mail@example.com", contact.Email)
	assert.Equal(t, "Test", *contact.FirstName)
	assert.Equal(t, "User", *contact.LastName)
	assert.Equal(t, "user_123", *contact.UserID)

	companyRole, ok := contact.CustomProperties["companyRole"]
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
	assert.Equal(t, "cm3n4kiua02c0t839btycnwe1", contact.ID)
	assert.Equal(t, "new-test-mail@example.com", contact.Email)
	assert.Equal(t, "Test", *contact.FirstName)
	assert.Equal(t, "User", *contact.LastName)
	assert.Equal(t, "user_123", *contact.UserID)

	companyRole, ok := contact.CustomProperties["companyRole"]
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
	require.Len(t, mailingLists, 1)
	assert.Equal(t, "cm3n274xf027h0mi33t4qhrdg", mailingLists[0].ID)
	assert.Equal(t, "Newsletter", mailingLists[0].Name)
	assert.True(t, mailingLists[0].IsPublic)
}

func TestSendEvent(t *testing.T) {
	client := newReplayTestClient(t, "send-event.replay.json")
	err := client.SendEvent(context.Background(), &Event{
		Email:     String("neil.armstrong@moon.space"),
		EventName: "joinedMission",
		EventProperties: &map[string]interface{}{
			"mission": "Apollo 11",
		},
	})
	require.NoError(t, err)
}

func TestSendTransactionalEmail(t *testing.T) {
	client := newReplayTestClient(t, "send-transactional-email.replay.json")
	err := client.SendTransactionalEmail(context.Background(), &TransactionalEmail{
		TransactionalID: "cm3n2vjux00cgeyeflew9ly2w",
		Email:           "test@example.com",
		DataVariables: &map[string]interface{}{
			"name": "Mr. Test",
		},
	})
	require.NoError(t, err)
}

func TestGetCustomFields(t *testing.T) {
	client := newReplayTestClient(t, "get-custom-fields.replay.json")
	customFields, err := client.GetCustomFields(context.Background())
	require.NoError(t, err)
	require.Len(t, customFields, 1)
	assert.Equal(t, "role", customFields[0].Key)
	assert.Equal(t, "Role", customFields[0].Label)
	assert.Equal(t, "string", customFields[0].Type)
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
