package loops

// String returns a pointer to the string value passed in.
func String(v string) *string {
	return &v
}

// Contact defines model for Contact.
type Contact struct {
	// The contact's ID.
	Id string `json:"id,omitempty"`
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
	UserId *string `json:"userId,omitempty"`
	// Mailing lists the contact is subscribed to.
	MailingLists map[string]interface{} `json:"mailingLists,omitempty"`
}

type ContactIdentifier struct {
	Email  *string `json:"email,omitempty"`
	UserId *string `json:"userId,omitempty"`
}

type MailingList struct {
	// The ID of the list.
	Id string `json:"id"`
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
	UserId *string `json:"userId,omitempty"`
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
	TransactionalId string `json:"transactionalId"`
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

type ApiKeyInfo struct {
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
