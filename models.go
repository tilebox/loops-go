package loops

import (
	"encoding/json"
	"errors"
)

// String returns a pointer to the string value passed in.
func String(v string) *string {
	return &v
}

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
	// Custom properties for the contact.
	CustomProperties map[string]interface{} `json:"-"` // there is no "customProperties", we need to inline add them to the json
}

// MarshalJSON overrides the default json marshaller to add custom properties inline to the root object
func (c *Contact) MarshalJSON() ([]byte, error) {
	data := map[string]interface{}{
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
	for k, v := range c.CustomProperties {
		data[k] = v
	}
	return json.Marshal(data)
}

// UnmarshalJSON overrides the default json unmarshaller to add custom properties inline to the root object
func (c *Contact) UnmarshalJSON(data []byte) error {
	values := map[string]interface{}{}
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

	mailingLists, ok := values["mailingLists"].(map[string]interface{})
	if ok {
		c.MailingLists = make(map[string]bool)
		for k, v := range mailingLists {
			c.MailingLists[k] = v.(bool)
		}
		delete(values, "mailingLists")
	}

	c.CustomProperties = make(map[string]interface{})
	for k, v := range values {
		c.CustomProperties[k] = v
	}
	return nil
}

type ContactIdentifier struct {
	Email  *string `json:"email,omitempty"`
	UserID *string `json:"userId,omitempty"`
}

type MailingList struct {
	// The ID of the list.
	ID string `json:"id"`
	// The name of the list.
	Name string `json:"name"`
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
	ContactProperties map[string]interface{} `json:"contactProperties,omitempty"`
	// Event properties, made available in emails triggered by the event.
	EventProperties *map[string]interface{} `json:"eventProperties,omitempty"`
	// An object of mailing list IDs and boolean subscription statuses.
	MailingLists *map[string]interface{} `json:"mailingLists,omitempty"`
}

type TransactionalEmail struct {
	// The ID of the transactional email to send.
	TransactionalID string `json:"transactionalId"`
	// The email address of the recipient
	Email string `json:"email"`
	// Create a contact in your audience using the provided email address (if one doesn't already exist).
	AddToAudience *bool `json:"addToAudience,omitempty"`
	// Data variables as defined by the transational email template.
	DataVariables *map[string]interface{} `json:"dataVariables,omitempty"`
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

type CustomField struct {
	// The property's name key
	Key string `json:"key"`
	// The human-friendly label for this property
	Label string `json:"label"`
	// The type of property (one of string, number, boolean or date)
	Type string `json:"type"`
}

type APIKeyInfo struct {
	Success bool `json:"success"`
	// The name of the team the API key belongs to.
	TeamName string `json:"teamName"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type IDResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id"`
}

type MessageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
