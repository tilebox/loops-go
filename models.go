package loops

import (
	"encoding/json"
	"errors"
	"maps"
)

// String returns a pointer to the string value passed in.
func String(v string) *string {
	return &v
}

// OptInStatus represents the double opt-in status of a contact.
type OptInStatus string

const (
	OptInStatusAccepted OptInStatus = "accepted"
	OptInStatusPending  OptInStatus = "pending"
	OptInStatusRejected OptInStatus = "rejected"
)

// Contact defines model for Contact.
type Contact struct {
	// The contact's ID.
	ID string `json:"id,omitempty"`
	// The contact's email address.
	Email string `json:"email,omitempty"`
	// The contact's first name.
	FirstName *string `json:"firstName,omitempty"`
	// The contact's last name.
	LastName *string `json:"lastName,omitempty"`
	// The source the contact was created from.
	Source *string `json:"source,omitempty"`
	// Whether the contact will receive campaign and loops emails.
	Subscribed bool `json:"subscribed,omitempty"`
	// The contact's user group (used to segemnt users when sending emails).
	UserGroup *string `json:"userGroup,omitempty"`
	// A unique user ID (for example, from an external application).
	UserID *string `json:"userId,omitempty"`
	// Mailing lists the contact is subscribed to.
	MailingLists map[string]bool `json:"mailingLists,omitempty"`
	// Double opt-in status.
	OptInStatus *OptInStatus `json:"optInStatus,omitempty"`
	// Custom properties for the contact.
	Properties map[string]any `json:"-"` // there is no "customProperties", we need to inline add them to the json
}

// MarshalJSON overrides the default json marshaller to add custom properties inline to the root object
func (c *Contact) MarshalJSON() ([]byte, error) {
	data := map[string]any{
		"id":         c.ID,
		"email":      c.Email,
		"subscribed": c.Subscribed,
	}
	if c.FirstName != nil {
		data["firstName"] = *c.FirstName
	}
	if c.LastName != nil {
		data["lastName"] = *c.LastName
	}
	if c.Source != nil {
		data["source"] = *c.Source
	}
	if c.UserGroup != nil {
		data["userGroup"] = *c.UserGroup
	}
	if c.UserID != nil {
		data["userId"] = *c.UserID
	}
	if c.MailingLists != nil {
		data["mailingLists"] = c.MailingLists
	}
	if c.OptInStatus != nil {
		data["optInStatus"] = *c.OptInStatus
	}
	maps.Copy(data, c.Properties)
	return json.Marshal(data)
}

// UnmarshalJSON overrides the default json unmarshaller to add custom properties inline to the root object
func (c *Contact) UnmarshalJSON(data []byte) error {
	values := map[string]any{}
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}

	if id, ok := values["id"].(string); ok {
		c.ID = id
		delete(values, "id")
	} else {
		return errors.New("missing or invalid 'id' field")
	}

	if email, ok := values["email"].(string); ok {
		c.Email = email
		delete(values, "email")
	} else {
		return errors.New("missing or invalid 'email' field")
	}

	if subscribed, ok := values["subscribed"].(bool); ok {
		c.Subscribed = subscribed
		delete(values, "subscribed")
	} else {
		return errors.New("missing or invalid 'subscribed' field")
	}

	if firstName, ok := values["firstName"].(string); ok {
		c.FirstName = &firstName
		delete(values, "firstName")
	}

	if lastName, ok := values["lastName"].(string); ok {
		c.LastName = &lastName
		delete(values, "lastName")
	}

	if source, ok := values["source"].(string); ok {
		c.Source = &source
		delete(values, "source")
	}

	if userGroup, ok := values["userGroup"].(string); ok {
		c.UserGroup = &userGroup
		delete(values, "userGroup")
	}

	if userID, ok := values["userId"].(string); ok {
		c.UserID = &userID
		delete(values, "userId")
	}

	mailingLists, ok := values["mailingLists"].(map[string]any)
	if ok {
		c.MailingLists = make(map[string]bool)
		for k, v := range mailingLists {
			c.MailingLists[k] = v.(bool)
		}
		delete(values, "mailingLists")
	}

	if optInStatus, ok := values["optInStatus"].(string); ok {
		status := OptInStatus(optInStatus)
		c.OptInStatus = &status
		delete(values, "optInStatus")
	}

	c.Properties = make(map[string]any)
	maps.Copy(c.Properties, values)
	return nil
}

type ContactIdentifier struct {
	Email  *string `json:"email,omitempty"`
	UserID *string `json:"userId,omitempty"`
}

type ContactProperty struct {
	// The property's name key
	Key string `json:"key"`
	// The human-friendly label for this property
	Label string `json:"label"`
	// The type of property (one of string, number, boolean or date)
	Type string `json:"type"`
}

// Deprecated: Use ContactProperty instead.
type CustomField = ContactProperty

type ContactPropertyCreate struct {
	// The property's name key (must be in camelCase, like `planName`)
	Name string `json:"name"`
	// The type of property (one of string, number, boolean or date)
	Type string `json:"type"`
}

type MailingList struct {
	// The ID of the list.
	ID string `json:"id"`
	// The name of the list.
	Name string `json:"name"`
	// The description of the list.
	Description string `json:"description"`
	// Whether the list is public (true) or private (false).
	// See: https://loops.so/docs/contacts/mailing-lists#list-visibility
	IsPublic bool `json:"isPublic"`
}

type Event struct {
	// The contact's email address
	Email *string `json:"email,omitempty"`
	// The contact's unique user ID. This must already have been added to your contact in Loops.
	UserID *string `json:"userId,omitempty"`
	// The name of the event
	EventName string `json:"eventName"`
	// Properties to update the contact with, including custom properties.
	ContactProperties map[string]any `json:"contactProperties,omitempty"`
	// Event properties, made available in emails triggered by the event.
	EventProperties *map[string]any `json:"eventProperties,omitempty"`
	// An object of mailing list IDs and boolean subscription statuses.
	MailingLists *map[string]any `json:"mailingLists,omitempty"`
}

type TransactionalEmail struct {
	// The ID of the transactional email to send.
	TransactionalID string `json:"transactionalId"`
	// The email address of the recipient
	Email string `json:"email"`
	// Create a contact in your audience using the provided email address (if one doesn't already exist).
	AddToAudience *bool `json:"addToAudience,omitempty"`
	// Data variables as defined by the transitional email template.
	DataVariables *map[string]any `json:"dataVariables,omitempty"`
	// File(s) to be sent along with the email message.
	Attachments *[]EmailAttachment `json:"attachments,omitempty"`
}

type EmailAttachment struct {
	// Filename The name of the file, shown in email clients.
	Filename string `json:"filename"`
	// ContentType The MIME type of the file.
	ContentType string `json:"contentType"`
	// Data The base64-encoded content of the file.
	Data string `json:"data"`
}

type TransactionalEmailInfo struct {
	// The ID of the transactional email.
	ID string `json:"id"`
	// The name of the transactional email.
	Name string `json:"name"`
	// The last time the transactional email was updated.
	LastUpdated string `json:"lastUpdated"`
	// The data variables used in the transactional email.
	DataVariables []string `json:"dataVariables"`
}

type TransactionalEmailList struct {
	Data       []*TransactionalEmailInfo `json:"data"`
	Pagination Pagination                `json:"pagination"`
}

type Pagination struct {
	// Total results found.
	TotalResults int `json:"totalResults"`
	// The number of results returned in this response.
	ReturnedResults int `json:"returnedResults"`
	// The maximum number of results requested.
	PerPage int `json:"perPage"`
	// Total number of pages.
	TotalPages int `json:"totalPages"`
	// The next cursor (for retrieving the next page of results using the cursor parameter), or empty string if there are no further pages.
	NextCursor string `json:"nextCursor,omitempty"`
	// The next page (for retrieving the next page of results using the page parameter), or empty string if there are no further pages.
	NextPage string `json:"nextPage,omitempty"`
}

type APIKeyInfo struct {
	Success bool `json:"success"`
	// The name of the team the API key belongs to.
	TeamName string `json:"teamName"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Success bool `json:"success"`
}

type IDResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id"`
}

type MessageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
