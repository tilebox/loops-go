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
		Properties: map[string]any{
			"companyRole": "Developer",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "cmk6vyub00c7b0i04dlregeit", contactID)
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
	assert.Equal(t, "cmk6vyub00c7b0i04dlregeit", contactID)
}

func TestFindContact(t *testing.T) {
	client := newReplayTestClient(t, "find-contact.replay.json")
	contact, err := client.FindContact(context.Background(), &ContactIdentifier{
		Email: String("new-test-mail@example.com"),
	})
	require.NoError(t, err)
	assert.Equal(t, "cmk6vyub00c7b0i04dlregeit", contact.ID)
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
	assert.Equal(t, "cmk6vyub00c7b0i04dlregeit", contact.ID)
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
	require.Len(t, allProperties, 14)
	assert.Equal(t, "firstName", allProperties[0].Key)
	assert.Equal(t, "First Name", allProperties[0].Label)
	assert.Equal(t, "string", allProperties[0].Type)
	assert.Equal(t, "lastName", allProperties[1].Key)

	customProperties, err := client.GetContactProperties(context.Background(), ContactPropertyListOptions{
		List: ContactPropertyTypeCustom,
	})
	require.NoError(t, err)
	require.Len(t, customProperties, 2)
	assert.Equal(t, "heardAboutChannel", customProperties[0].Key)
	assert.Equal(t, "Heard About Channel", customProperties[0].Label)
	assert.Equal(t, "string", customProperties[0].Type)
	assert.Equal(t, "companyRole", customProperties[1].Key)
}

func TestCreateContactProperty(t *testing.T) {
	client := newReplayTestClient(t, "create-contact-property.replay.json")
	err := client.CreateContactProperty(context.Background(), &ContactPropertyCreate{
		Name: "planName",
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
		EventProperties: &map[string]any{
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
		DataVariables: &map[string]any{
			"name": "Mr. Test",
		},
	})
	require.NoError(t, err)
}

func TestGetDedicatedSendingIPs(t *testing.T) {
	client := newReplayTestClient(t, "get-dedicated-sending-ips.replay.json")
	ips, err := client.GetDedicatedSendingIPs(context.Background())
	require.NoError(t, err)
	require.Len(t, ips, 5)
	assert.Contains(t, ips[0], "221.169")
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
	assert.Equal(t, "2024-11-18T14:32:35.586Z", response.Data[0].LastUpdated)
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
