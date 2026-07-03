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

// Bool returns a pointer to the bool value passed in.
func Bool(v bool) *bool {
	return &v
}

// OptInStatus represents the double opt-in status of a contact.
type OptInStatus string

const (
	OptInStatusAccepted OptInStatus = "accepted"
	OptInStatusPending  OptInStatus = "pending"
	OptInStatusRejected OptInStatus = "rejected"
)

type Contact struct {
	ID           string          `json:"id,omitempty"`
	Email        string          `json:"email,omitempty"`
	FirstName    *string         `json:"firstName,omitempty"`
	LastName     *string         `json:"lastName,omitempty"`
	Source       *string         `json:"source,omitempty"`
	Subscribed   bool            `json:"subscribed,omitempty"`
	UserGroup    *string         `json:"userGroup,omitempty"`
	UserID       *string         `json:"userId,omitempty"`
	MailingLists map[string]bool `json:"mailingLists,omitempty"`
	OptInStatus  *OptInStatus    `json:"optInStatus,omitempty"`
	Properties   map[string]any  `json:"-"`
}

// MarshalJSON overrides the default json marshaller to add custom properties inline to the root object.
func (c *Contact) MarshalJSON() ([]byte, error) {
	data := map[string]any{}
	if c.ID != "" {
		data["id"] = c.ID
	}
	if c.Email != "" {
		data["email"] = c.Email
	}
	if c.Subscribed {
		data["subscribed"] = c.Subscribed
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

// UnmarshalJSON overrides the default json unmarshaller to add custom properties inline to the root object.
func (c *Contact) UnmarshalJSON(data []byte) error {
	values := map[string]any{}
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}
	if id, ok := values["id"].(string); ok {
		c.ID = id
		delete(values, "id")
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
	if mailingLists, ok := values["mailingLists"].(map[string]any); ok {
		c.MailingLists = make(map[string]bool)
		for k, v := range mailingLists {
			if b, ok := v.(bool); ok {
				c.MailingLists[k] = b
			}
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
	Key   string `json:"key"`
	Label string `json:"label"`
	Type  string `json:"type"`
}
type CustomField = ContactProperty

type MailingList struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPublic    bool   `json:"isPublic"`
}

type Pagination struct {
	TotalResults    int    `json:"totalResults"`
	ReturnedResults int    `json:"returnedResults"`
	PerPage         int    `json:"perPage"`
	TotalPages      int    `json:"totalPages"`
	NextCursor      string `json:"nextCursor,omitempty"`
	NextPage        string `json:"nextPage,omitempty"`
}

type Event struct {
	Email             *string         `json:"email,omitempty"`
	UserID            *string         `json:"userId,omitempty"`
	EventName         string          `json:"eventName"`
	EventProperties   map[string]any  `json:"eventProperties,omitempty"`
	MailingLists      map[string]bool `json:"mailingLists,omitempty"`
	Properties        map[string]any  `json:"-"`
	ContactProperties map[string]any  `json:"-"`
}

func (e *Event) MarshalJSON() ([]byte, error) {
	data := map[string]any{"eventName": e.EventName}
	if e.Email != nil {
		data["email"] = *e.Email
	}
	if e.UserID != nil {
		data["userId"] = *e.UserID
	}
	if e.EventProperties != nil {
		data["eventProperties"] = e.EventProperties
	}
	if e.MailingLists != nil {
		data["mailingLists"] = e.MailingLists
	}
	maps.Copy(data, e.ContactProperties)
	maps.Copy(data, e.Properties)
	return json.Marshal(data)
}

type TransactionalRequest struct {
	Email           string            `json:"email"`
	TransactionalID string            `json:"transactionalId"`
	AddToAudience   *bool             `json:"addToAudience,omitempty"`
	DataVariables   map[string]any    `json:"dataVariables,omitempty"`
	Attachments     []EmailAttachment `json:"attachments,omitempty"`
}
type SendTransactionalEmailRequest = TransactionalRequest

type EmailAttachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Data        string `json:"data"`
}

type ActivityCondition struct {
	Type   string `json:"type"`
	Action string `json:"action"`
	Negate bool   `json:"negate"`
	Target string `json:"target"`
	// The ID of the campaign, workflow, or workflow email.
	ID string `json:"id"`
}

type AddToListTriggerWorkflowNode struct {
	ID            string   `json:"id"`
	WorkflowID    string   `json:"workflowId"`
	TypeName      string   `json:"typeName"`
	NextNodeIDs   []string `json:"nextNodeIds"`
	MailingListID string   `json:"mailingListId"`
	ReEligible    bool     `json:"reEligible"`
}

type AudienceFilter struct {
	Match      string                    `json:"match"`
	Conditions []AudienceFilterCondition `json:"conditions"`
}

type AudienceFilterCondition map[string]any

type AudienceFilterWorkflowNode struct {
	ID                string          `json:"id"`
	WorkflowID        string          `json:"workflowId"`
	TypeName          string          `json:"typeName"`
	NextNodeIDs       []string        `json:"nextNodeIds"`
	AudienceFilter    *AudienceFilter `json:"audienceFilter,omitempty"`
	AudienceSegmentID *string         `json:"audienceSegmentId,omitempty"`
	AppliesDownstream bool            `json:"appliesDownstream"`
}

type AudienceSegment struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	// ISO 8601 timestamp.
	CreatedAt string `json:"createdAt"`
	// ISO 8601 timestamp.
	UpdatedAt string         `json:"updatedAt"`
	Filter    AudienceFilter `json:"filter"`
}

type AudienceSegmentFailureResponse struct {
	Message string `json:"message"`
}

type AudienceSegmentResponse map[string]any

type BlankTriggerWorkflowNode struct {
	ID          string   `json:"id"`
	WorkflowID  string   `json:"workflowId"`
	TypeName    string   `json:"typeName"`
	NextNodeIDs []string `json:"nextNodeIds"`
}

type BranchWorkflowNode struct {
	ID           string   `json:"id"`
	WorkflowID   string   `json:"workflowId"`
	TypeName     string   `json:"typeName"`
	NextNodeIDs  []string `json:"nextNodeIds"`
	EvalStrategy *string  `json:"evalStrategy,omitempty"`
}

type CampaignFailureResponse struct {
	Message string `json:"message"`
}

type CampaignListItem map[string]any

type CampaignResponse struct {
	ID                string             `json:"id"`
	Name              string             `json:"name"`
	Status            string             `json:"status"`
	CreatedAt         string             `json:"createdAt"`
	UpdatedAt         string             `json:"updatedAt"`
	EmailMessageID    *string            `json:"emailMessageId"`
	CampaignGroupID   *string            `json:"campaignGroupId"`
	MailingListID     *string            `json:"mailingListId"`
	AudienceSegmentID *string            `json:"audienceSegmentId"`
	AudienceFilter    AudienceFilter     `json:"audienceFilter"`
	Scheduling        CampaignScheduling `json:"scheduling"`
}

type CampaignScheduling struct {
	Method string `json:"method"`
	// ISO 8601 send time. Null when the method is `now`.
	Timestamp *string `json:"timestamp"`
}

type CampaignSchedulingRequest struct {
	Method    string  `json:"method"`
	Timestamp *string `json:"timestamp,omitempty"`
}

type CompleteUploadResponse struct {
	EmailAssetID string `json:"emailAssetId"`
	// The public URL of the uploaded asset.
	FinalURL string `json:"finalUrl"`
}

type Component struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// The component body serialized as LMX.
	LMX string `json:"lmx"`
}

type ComponentFailureResponse struct {
	Message string `json:"message"`
}

type ComponentResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// The component body serialized as LMX.
	LMX string `json:"lmx"`
}

type ComponentValidationFailureResponse struct {
	Message string `json:"message"`
	// The dynamic variables that the change would push into an email that cannot use them. Present only when the update was rejected for that reason.
	InvalidTags []string `json:"invalidTags,omitempty"`
}

type ContactDeleteRequest struct {
	Email  string `json:"email"`
	UserID string `json:"userId"`
}

type ContactDeleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ContactFailureResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ContactPropertyCreateRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type ContactPropertyFailureResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ContactPropertySuccessResponse struct {
	Success bool `json:"success"`
}

type ContactPropertyTriggerWorkflowNode struct {
	ID                   string                        `json:"id"`
	WorkflowID           string                        `json:"workflowId"`
	TypeName             string                        `json:"typeName"`
	NextNodeIDs          []string                      `json:"nextNodeIds"`
	ContactPropertyQuery *WorkflowContactPropertyQuery `json:"contactPropertyQuery"`
	ReEligible           bool                          `json:"reEligible"`
}

type ContactRequest struct {
	Email      string  `json:"email"`
	FirstName  *string `json:"firstName,omitempty"`
	LastName   *string `json:"lastName,omitempty"`
	Subscribed *bool   `json:"subscribed,omitempty"`
	UserGroup  *string `json:"userGroup,omitempty"`
	UserID     *string `json:"userId,omitempty"`
	// An object of mailing list IDs and boolean subscription statuses.
	MailingLists map[string]any `json:"mailingLists,omitempty"`
}

type ContactSuccessResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id"`
}

type ContactSuppressionRemovalQuota struct {
	Limit     float64 `json:"limit"`
	Remaining float64 `json:"remaining"`
}

type ContactSuppressionRemoveResponse struct {
	Success      bool                           `json:"success"`
	Message      string                         `json:"message"`
	RemovalQuota ContactSuppressionRemovalQuota `json:"removalQuota"`
}

type ContactSuppressionStatusResponse struct {
	Contact struct {
		ID     string  `json:"id"`
		Email  string  `json:"email"`
		UserID *string `json:"userId"`
	} `json:"contact"`
	IsSuppressed bool                           `json:"isSuppressed"`
	RemovalQuota ContactSuppressionRemovalQuota `json:"removalQuota"`
}

type ContactUpdateRequest struct {
	Email      *string `json:"email,omitempty"`
	FirstName  *string `json:"firstName,omitempty"`
	LastName   *string `json:"lastName,omitempty"`
	Subscribed *bool   `json:"subscribed,omitempty"`
	UserGroup  *string `json:"userGroup,omitempty"`
	UserID     *string `json:"userId,omitempty"`
	// An object of mailing list IDs and boolean subscription statuses.
	MailingLists map[string]any `json:"mailingLists,omitempty"`
}

type CreateCampaignRequest struct {
	// The campaign name.
	Name string `json:"name"`
	// The ID of the group to add this campaign to. Defaults to the team's default group when omitted.
	CampaignGroupID *string `json:"campaignGroupId,omitempty"`
	// The ID of the mailing list to send to.
	MailingListID *string `json:"mailingListId,omitempty"`
	// The ID of an audience segment. Setting this clears any `audienceFilter`.
	AudienceSegmentID *string                    `json:"audienceSegmentId,omitempty"`
	AudienceFilter    *AudienceFilter            `json:"audienceFilter,omitempty"`
	Scheduling        *CampaignSchedulingRequest `json:"scheduling,omitempty"`
}

type CreateCampaignResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	// The ID of the empty email message created for this campaign. Use `/email-messages/{emailMessageId}` to set its fields and LMX content.
	EmailMessageID *string `json:"emailMessageId"`
	// The `contentRevisionId` of the newly created email message. Pass this as `expectedRevisionId` on your first update.
	EmailMessageContentRevisionID *string            `json:"emailMessageContentRevisionId"`
	CampaignGroupID               *string            `json:"campaignGroupId"`
	MailingListID                 *string            `json:"mailingListId"`
	AudienceSegmentID             *string            `json:"audienceSegmentId"`
	AudienceFilter                AudienceFilter     `json:"audienceFilter"`
	Scheduling                    CampaignScheduling `json:"scheduling"`
}

type CreateComponentBody struct {
	// The component name.
	Name string `json:"name"`
	// The component body as an LMX string.
	LMX string `json:"lmx"`
}

type CreateGroupRequest struct {
	// The group name. Cannot be the reserved name "Unsorted".
	Name string `json:"name"`
	// An optional description for the group.
	Description *string `json:"description,omitempty"`
}

type CreateThemeBody struct {
	// The theme name.
	Name   string       `json:"name"`
	Styles *ThemeStyles `json:"styles,omitempty"`
}

type CreateTransactionalRequest struct {
	// The name of the transactional email.
	Name string `json:"name"`
	// The ID of the group to add this transactional email to. Defaults to the team's default group when omitted.
	TransactionalGroupID *string `json:"transactionalGroupId,omitempty"`
}

type CreateUploadRequest struct {
	// The MIME type of the file to upload. Supported types are `image/jpeg`, `image/png`, `image/gif` and `image/webp`.
	ContentType string `json:"contentType"`
	// The size of the file in bytes. Must be a positive integer no greater than 4,000,000 bytes.
	ContentLength int `json:"contentLength"`
}

type CreateUploadResponse struct {
	// The ID of the created asset. Pass this as `id` to `/uploads/{id}/complete` once the file has been uploaded.
	EmailAssetID string `json:"emailAssetId"`
	// The pre-signed URL to upload the file to with an HTTP `PUT` request. Send the same `Content-Type` and `Content-Length` used in the create request.
	PresignedURL string `json:"presignedUrl"`
}

type EmailMessageFailureResponse struct {
	Message string `json:"message"`
}

type EmailMessageGuardianResponse struct {
	// Validation errors. These must be resolved before the email can be published.
	Errors []GuardianRule `json:"errors"`
	// Validation warnings. These are advisory and do not block publishing.
	Warnings []GuardianRule `json:"warnings"`
}

type EmailMessagePreviewRequest struct {
	// One or more addresses to send the preview to.
	Emails []string `json:"emails"`
	// Contact property values to render. Accepted for campaign and workflow previews.
	ContactProperties map[string]string `json:"contactProperties,omitempty"`
	// Event property values to render. Accepted for workflow previews only.
	EventProperties map[string]string `json:"eventProperties,omitempty"`
	// Transactional data variables to render. Accepted for transactional previews only.
	DataVariables map[string]any `json:"dataVariables,omitempty"`
}

type EmailMessagePreviewResponse struct {
	// The ID of the email message the preview was sent for.
	ID string `json:"id"`
}

type EmailMessageResponse struct {
	ID string `json:"id"`
	// The campaign this email message belongs to. Present only when the message belongs to a campaign (mutually exclusive with `transactionalId`).
	CampaignID *string `json:"campaignId,omitempty"`
	// The transactional email this email message belongs to. Present only when the message belongs to a transactional email (mutually exclusive with `campaignId`).
	TransactionalID *string `json:"transactionalId,omitempty"`
	Subject         string  `json:"subject"`
	PreviewText     string  `json:"previewText"`
	FromName        string  `json:"fromName"`
	FromEmail       string  `json:"fromEmail"`
	ReplyToEmail    string  `json:"replyToEmail"`
	// Only present when set.
	CCEmail *string `json:"ccEmail,omitempty"`
	// Only present when set.
	BCCEmail *string `json:"bccEmail,omitempty"`
	// Only present when set.
	LanguageCode *string `json:"languageCode,omitempty"`
	// The rendering format of the email.
	EmailFormat string `json:"emailFormat"`
	// The email body serialized as LMX.
	LMX string `json:"lmx"`
	// The current content revision. Pass this as `expectedRevisionId` on your next update.
	ContentRevisionID *string `json:"contentRevisionId"`
	UpdatedAt         string  `json:"updatedAt"`
	// Fallback values for contact properties. Only present when set.
	ContactPropertiesFallbacks map[string]string `json:"contactPropertiesFallbacks,omitempty"`
	// Fallback values for event properties. Only present when set.
	EventPropertiesFallbacks map[string]string `json:"eventPropertiesFallbacks,omitempty"`
	// Fallback values for data variables. Only present when set.
	DataVariablesFallbacks map[string]string `json:"dataVariablesFallbacks,omitempty"`
	// Non-fatal issues raised while compiling the submitted LMX. Only present on update responses when warnings were produced.
	Warnings []struct {
		Rule     string  `json:"rule"`
		Severity string  `json:"severity"`
		Message  string  `json:"message"`
		Path     *string `json:"path,omitempty"`
	} `json:"warnings,omitempty"`
}

type EventFailureResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type EventSuccessResponse struct {
	Success bool `json:"success"`
}

type EventTriggerWorkflowNode struct {
	ID              string                  `json:"id"`
	WorkflowID      string                  `json:"workflowId"`
	TypeName        string                  `json:"typeName"`
	NextNodeIDs     []string                `json:"nextNodeIds"`
	EventName       *string                 `json:"eventName,omitempty"`
	EventProperties []WorkflowEventProperty `json:"eventProperties,omitempty"`
	ReEligible      bool                    `json:"reEligible"`
}

type ExitActionWorkflowNode struct {
	ID          string   `json:"id"`
	WorkflowID  string   `json:"workflowId"`
	TypeName    string   `json:"typeName"`
	NextNodeIDs []string `json:"nextNodeIds"`
}

type ExperimentBranchWorkflowNode struct {
	ID             string                 `json:"id"`
	WorkflowID     string                 `json:"workflowId"`
	TypeName       string                 `json:"typeName"`
	NextNodeIDs    []string               `json:"nextNodeIds"`
	SamplingRate   float64                `json:"samplingRate"`
	URL            *string                `json:"url,omitempty"`
	ExperimentID   *string                `json:"experimentId,omitempty"`
	ExperimentType WorkflowExperimentType `json:"experimentType"`
}

type GroupFailureResponse struct {
	Message string `json:"message"`
}

type GroupResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type GuardianRule struct {
	// The identifier of the Guardian rule that fired.
	Rule string `json:"rule"`
	// A human-readable title for the rule.
	Title string `json:"title"`
	// A human-readable explanation of the rule.
	Description string `json:"description"`
	// The individual items that triggered the rule.
	Items []struct {
		Label    string  `json:"label"`
		CodeName *string `json:"codeName,omitempty"`
	} `json:"items"`
}

type IdempotencyKeyFailureResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ListAudienceSegmentsResponse struct {
	Pagination Pagination        `json:"pagination"`
	Data       []AudienceSegment `json:"data"`
}

type ListCampaignsResponse struct {
	Pagination Pagination         `json:"pagination"`
	Data       []CampaignResponse `json:"data"`
}

type ListComponentsResponse struct {
	Pagination Pagination  `json:"pagination"`
	Data       []Component `json:"data"`
}

type ListGroupsResponse struct {
	Pagination Pagination      `json:"pagination"`
	Data       []GroupResponse `json:"data"`
}

type ListThemesResponse struct {
	Pagination Pagination `json:"pagination"`
	Data       []Theme    `json:"data"`
}

type ListTransactionalsResourceResponse struct {
	Pagination Pagination                   `json:"pagination"`
	Data       []TransactionalEmailResource `json:"data"`
}

type ListTransactionalsResponse struct {
	Pagination Pagination           `json:"pagination"`
	Data       []TransactionalEmail `json:"data,omitempty"`
}

type ListWorkflowsResponse struct {
	Pagination Pagination        `json:"pagination"`
	Data       []WorkflowSummary `json:"data"`
}

type OptInCondition struct {
	Type   string  `json:"type"`
	Status *string `json:"status"`
}

type PropertyCondition struct {
	Type string `json:"type"`
	// The contact property name.
	Key      string `json:"key"`
	Operator string `json:"operator"`
	// The comparison value. Omitted for value-less operators (e.g. `isTrue`, `empty`). A `{ from, to }` object for `between`.
	Value any `json:"value,omitempty"`
}

type SendEmailActionWorkflowNode struct {
	ID             string   `json:"id"`
	WorkflowID     string   `json:"workflowId"`
	TypeName       string   `json:"typeName"`
	NextNodeIDs    []string `json:"nextNodeIds"`
	EmailMessageID string   `json:"emailMessageId"`
	Subject        string   `json:"subject"`
}

type SignupTriggerWorkflowNode struct {
	ID          string   `json:"id"`
	WorkflowID  string   `json:"workflowId"`
	TypeName    string   `json:"typeName"`
	NextNodeIDs []string `json:"nextNodeIds"`
}

type SimplifiedAddToListTriggerWorkflowNode struct {
	TypeName      string   `json:"typeName"`
	NextNodeIDs   []string `json:"nextNodeIds"`
	MailingListID *string  `json:"mailingListId"`
	ReEligible    bool     `json:"reEligible"`
}

type SimplifiedAudienceFilterWorkflowNode struct {
	TypeName    string   `json:"typeName"`
	NextNodeIDs []string `json:"nextNodeIds"`
}

type SimplifiedBlankTriggerWorkflowNode struct {
	TypeName    string   `json:"typeName"`
	NextNodeIDs []string `json:"nextNodeIds"`
}

type SimplifiedBranchWorkflowNode struct {
	TypeName    string   `json:"typeName"`
	NextNodeIDs []string `json:"nextNodeIds"`
}

type SimplifiedContactPropertyTriggerWorkflowNode struct {
	TypeName             string                        `json:"typeName"`
	NextNodeIDs          []string                      `json:"nextNodeIds"`
	ContactPropertyQuery *WorkflowContactPropertyQuery `json:"contactPropertyQuery"`
	ReEligible           bool                          `json:"reEligible"`
}

type SimplifiedEventTriggerWorkflowNode struct {
	TypeName    string   `json:"typeName"`
	NextNodeIDs []string `json:"nextNodeIds"`
	EventName   *string  `json:"eventName"`
	ReEligible  bool     `json:"reEligible"`
}

type SimplifiedExitActionWorkflowNode struct {
	TypeName    string   `json:"typeName"`
	NextNodeIDs []string `json:"nextNodeIds"`
}

type SimplifiedExperimentBranchWorkflowNode struct {
	TypeName     string   `json:"typeName"`
	NextNodeIDs  []string `json:"nextNodeIds"`
	SamplingRate float64  `json:"samplingRate"`
}

type SimplifiedSendEmailActionWorkflowNode struct {
	TypeName       string   `json:"typeName"`
	NextNodeIDs    []string `json:"nextNodeIds"`
	EmailMessageID *string  `json:"emailMessageId"`
	Subject        string   `json:"subject"`
}

type SimplifiedSignupTriggerWorkflowNode struct {
	TypeName    string   `json:"typeName"`
	NextNodeIDs []string `json:"nextNodeIds"`
}

type SimplifiedTimerActionWorkflowNode struct {
	TypeName    string            `json:"typeName"`
	NextNodeIDs []string          `json:"nextNodeIds"`
	Amount      float64           `json:"amount"`
	Unit        WorkflowTimerUnit `json:"unit"`
}

type SimplifiedVariantWorkflowNode struct {
	TypeName    string   `json:"typeName"`
	NextNodeIDs []string `json:"nextNodeIds"`
	VariantID   string   `json:"variantId"`
	IsControl   bool     `json:"isControl"`
}

type SimplifiedWorkflow struct {
	ID            string                            `json:"id"`
	Status        string                            `json:"status"`
	Name          *string                           `json:"name,omitempty"`
	Description   *string                           `json:"description,omitempty"`
	Emoji         *string                           `json:"emoji,omitempty"`
	MailingListID *string                           `json:"mailingListId"`
	RootNodeID    *string                           `json:"rootNodeId"`
	Nodes         map[string]SimplifiedWorkflowNode `json:"nodes"`
}

type SimplifiedWorkflowNode map[string]any

type Theme struct {
	ID     string      `json:"id"`
	Name   string      `json:"name"`
	Styles ThemeStyles `json:"styles"`
	// Whether this theme is the team's default.
	IsDefault bool `json:"isDefault"`
	// ISO 8601 timestamp.
	CreatedAt string `json:"createdAt"`
	// ISO 8601 timestamp.
	UpdatedAt string `json:"updatedAt"`
}

type ThemeFailureResponse struct {
	Message string `json:"message"`
}

type ThemeResponse struct {
	ID     string      `json:"id"`
	Name   string      `json:"name"`
	Styles ThemeStyles `json:"styles"`
	// Whether this theme is the team's default.
	IsDefault bool `json:"isDefault"`
	// ISO 8601 timestamp.
	CreatedAt string `json:"createdAt"`
	// ISO 8601 timestamp.
	UpdatedAt string `json:"updatedAt"`
}

type ThemeStyles struct {
	BackgroundColor       *string  `json:"backgroundColor,omitempty"`
	BackgroundXPadding    *float64 `json:"backgroundXPadding,omitempty"`
	BackgroundYPadding    *float64 `json:"backgroundYPadding,omitempty"`
	BodyColor             *string  `json:"bodyColor,omitempty"`
	BodyXPadding          *float64 `json:"bodyXPadding,omitempty"`
	BodyYPadding          *float64 `json:"bodyYPadding,omitempty"`
	BodyFontFamily        *string  `json:"bodyFontFamily,omitempty"`
	BodyFontCategory      *string  `json:"bodyFontCategory,omitempty"`
	BorderColor           *string  `json:"borderColor,omitempty"`
	BorderWidth           *float64 `json:"borderWidth,omitempty"`
	BorderRadius          *float64 `json:"borderRadius,omitempty"`
	ButtonBodyColor       *string  `json:"buttonBodyColor,omitempty"`
	ButtonBodyXPadding    *float64 `json:"buttonBodyXPadding,omitempty"`
	ButtonBodyYPadding    *float64 `json:"buttonBodyYPadding,omitempty"`
	ButtonBorderColor     *string  `json:"buttonBorderColor,omitempty"`
	ButtonBorderWidth     *float64 `json:"buttonBorderWidth,omitempty"`
	ButtonBorderRadius    *float64 `json:"buttonBorderRadius,omitempty"`
	ButtonTextColor       *string  `json:"buttonTextColor,omitempty"`
	ButtonTextFormat      *float64 `json:"buttonTextFormat,omitempty"`
	ButtonTextFontSize    *float64 `json:"buttonTextFontSize,omitempty"`
	DividerColor          *string  `json:"dividerColor,omitempty"`
	DividerBorderWidth    *float64 `json:"dividerBorderWidth,omitempty"`
	TextBaseColor         *string  `json:"textBaseColor,omitempty"`
	TextBaseFontSize      *float64 `json:"textBaseFontSize,omitempty"`
	TextBaseLineHeight    *float64 `json:"textBaseLineHeight,omitempty"`
	TextBaseLetterSpacing *float64 `json:"textBaseLetterSpacing,omitempty"`
	TextLinkColor         *string  `json:"textLinkColor,omitempty"`
	Heading1Color         *string  `json:"heading1Color,omitempty"`
	Heading1FontSize      *float64 `json:"heading1FontSize,omitempty"`
	Heading1LineHeight    *float64 `json:"heading1LineHeight,omitempty"`
	Heading1LetterSpacing *float64 `json:"heading1LetterSpacing,omitempty"`
	Heading2Color         *string  `json:"heading2Color,omitempty"`
	Heading2FontSize      *float64 `json:"heading2FontSize,omitempty"`
	Heading2LineHeight    *float64 `json:"heading2LineHeight,omitempty"`
	Heading2LetterSpacing *float64 `json:"heading2LetterSpacing,omitempty"`
	Heading3Color         *string  `json:"heading3Color,omitempty"`
	Heading3FontSize      *float64 `json:"heading3FontSize,omitempty"`
	Heading3LineHeight    *float64 `json:"heading3LineHeight,omitempty"`
	Heading3LetterSpacing *float64 `json:"heading3LetterSpacing,omitempty"`
}

type TimerActionWorkflowNode struct {
	ID          string            `json:"id"`
	WorkflowID  string            `json:"workflowId"`
	TypeName    string            `json:"typeName"`
	NextNodeIDs []string          `json:"nextNodeIds"`
	Amount      float64           `json:"amount"`
	Unit        WorkflowTimerUnit `json:"unit"`
}

type TransactionalDraftResponse struct {
	ID                  string  `json:"id"`
	Name                string  `json:"name"`
	DraftEmailMessageID *string `json:"draftEmailMessageId"`
	// The `contentRevisionId` of the draft email message. Pass this as `expectedRevisionId` on your first update via `/email-messages/{emailMessageId}`.
	DraftEmailMessageContentRevisionID *string `json:"draftEmailMessageContentRevisionId"`
	PublishedEmailMessageID            *string `json:"publishedEmailMessageId"`
	// The ID of the group this transactional email belongs to. Returned when creating a transactional email.
	TransactionalGroupID *string `json:"transactionalGroupId,omitempty"`
	CreatedAt            string  `json:"createdAt"`
	UpdatedAt            string  `json:"updatedAt"`
	// Data variable names used by the published email. Empty for unpublished transactional emails.
	DataVariables []string `json:"dataVariables"`
}

type TransactionalEmail struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	LastUpdated   string   `json:"lastUpdated"`
	DataVariables []string `json:"dataVariables"`
}

type TransactionalEmailResource struct {
	ID                      string  `json:"id"`
	Name                    string  `json:"name"`
	DraftEmailMessageID     *string `json:"draftEmailMessageId"`
	PublishedEmailMessageID *string `json:"publishedEmailMessageId"`
	// The ID of the group this transactional email belongs to.
	TransactionalGroupID *string `json:"transactionalGroupId"`
	CreatedAt            string  `json:"createdAt"`
	UpdatedAt            string  `json:"updatedAt"`
	// Data variable names used by the published email. Empty for unpublished transactional emails.
	DataVariables []string `json:"dataVariables"`
}

type TransactionalFailure2Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Path    string `json:"path"`
}

type TransactionalFailure3Response struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	ErrorValue struct {
		Path    *string `json:"path,omitempty"`
		Message *string `json:"message,omitempty"`
	} `json:"error"`
}

type TransactionalFailure4Response struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	ErrorValue struct {
		Path   *string `json:"path,omitempty"`
		Reason *string `json:"reason,omitempty"`
	} `json:"error"`
}

type TransactionalFailure5Response struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	ErrorValue struct {
		Path    *string `json:"path,omitempty"`
		Message *string `json:"message,omitempty"`
	} `json:"error"`
	TransactionalID string `json:"transactionalId"`
}

type TransactionalFailureResponse struct {
	Message string `json:"message"`
}

type TransactionalResponse struct {
	ID                      string  `json:"id"`
	Name                    string  `json:"name"`
	DraftEmailMessageID     *string `json:"draftEmailMessageId"`
	PublishedEmailMessageID *string `json:"publishedEmailMessageId"`
	// The ID of the group this transactional email belongs to.
	TransactionalGroupID *string `json:"transactionalGroupId"`
	CreatedAt            string  `json:"createdAt"`
	UpdatedAt            string  `json:"updatedAt"`
	// Data variable names used by the published email. Empty for unpublished transactional emails.
	DataVariables []string `json:"dataVariables"`
}

type TransactionalSendFailureResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type TransactionalSuccessResponse struct {
	Success bool `json:"success"`
}

type UpdateCampaignRequest struct {
	Name *string `json:"name,omitempty"`
	// The ID of the group to move this campaign to.
	CampaignGroupID *string `json:"campaignGroupId,omitempty"`
	// The ID of the mailing list to send to.
	MailingListID *string `json:"mailingListId,omitempty"`
	// The ID of an audience segment. Setting this clears any `audienceFilter`.
	AudienceSegmentID *string                    `json:"audienceSegmentId,omitempty"`
	AudienceFilter    *AudienceFilter            `json:"audienceFilter,omitempty"`
	Scheduling        *CampaignSchedulingRequest `json:"scheduling,omitempty"`
}

type UpdateComponentBody struct {
	Name *string `json:"name,omitempty"`
	// The component body as an LMX string.
	LMX *string `json:"lmx,omitempty"`
}

type UpdateComponentResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// The component body serialized as LMX.
	LMX string `json:"lmx"`
	// The number of emails using this component that were updated by the body change. `0` when only the name changed.
	AffectedEmailCount float64 `json:"affectedEmailCount"`
}

type UpdateEmailMessageRequest struct {
	// The `contentRevisionId` you last fetched. Used for optimistic concurrency — the request is rejected with 409 if the server's revision has advanced.
	ExpectedRevisionID *string `json:"expectedRevisionId,omitempty"`
	Subject            *string `json:"subject,omitempty"`
	PreviewText        *string `json:"previewText,omitempty"`
	FromName           *string `json:"fromName,omitempty"`
	// The sender username (without `@` or domain). The team's sending domain is appended automatically.
	FromEmail *string `json:"fromEmail,omitempty"`
	// Reply-to email. Must be empty or a valid email address.
	ReplyToEmail *string `json:"replyToEmail,omitempty"`
	// CC email address. Requires the team to have CC/BCC enabled.
	CCEmail *string `json:"ccEmail,omitempty"`
	// BCC email address. Requires the team to have CC/BCC enabled.
	BCCEmail *string `json:"bccEmail,omitempty"`
	// Language code for the email. Requires translation to be enabled for the team.
	LanguageCode *string `json:"languageCode,omitempty"`
	// The rendering format of the email.
	EmailFormat *string `json:"emailFormat,omitempty"`
	// The email body serialized as LMX. Styles must be embedded in the LMX `<Style />` tag.
	LMX *string `json:"lmx,omitempty"`
	// Fallback values for contact properties, keyed by property name. Per-key merge: a string value sets the fallback, a null value deletes it, and keys omitted from the map are left unchanged.
	ContactPropertiesFallbacks map[string]*string `json:"contactPropertiesFallbacks,omitempty"`
	// Fallback values for event properties, keyed by property name. Per-key merge: a string value sets the fallback, a null value deletes it, and keys omitted from the map are left unchanged.
	EventPropertiesFallbacks map[string]*string `json:"eventPropertiesFallbacks,omitempty"`
	// Fallback values for data variables, keyed by variable name. Per-key merge: a string value sets the fallback, a null value deletes it, and keys omitted from the map are left unchanged.
	DataVariablesFallbacks map[string]*string `json:"dataVariablesFallbacks,omitempty"`
}

type UpdateGroupRequest struct {
	// The group name. Cannot be the reserved name "Unsorted".
	Name *string `json:"name,omitempty"`
	// A description for the group.
	Description *string `json:"description,omitempty"`
}

type UpdateThemeBody struct {
	Name   *string      `json:"name,omitempty"`
	Styles *ThemeStyles `json:"styles,omitempty"`
}

type UpdateThemeResponse struct {
	ID     string      `json:"id"`
	Name   string      `json:"name"`
	Styles ThemeStyles `json:"styles"`
	// Whether this theme is the team's default.
	IsDefault bool `json:"isDefault"`
	// ISO 8601 timestamp.
	CreatedAt string `json:"createdAt"`
	// ISO 8601 timestamp.
	UpdatedAt string `json:"updatedAt"`
	// The number of emails using this theme that are affected by the style change. `0` when only the name changed.
	AffectedEmailCount float64 `json:"affectedEmailCount"`
}

type UpdateTransactionalRequest struct {
	Name *string `json:"name,omitempty"`
	// The ID of the group to move this transactional email to.
	TransactionalGroupID *string `json:"transactionalGroupId,omitempty"`
}

type UploadFailureResponse struct {
	Message string `json:"message"`
	// Present when the request was rejected for an unsupported `contentType`. Lists the accepted MIME types.
	SupportedContentTypes []string `json:"supportedContentTypes,omitempty"`
	// Present when the upload exceeds the size limit. The maximum allowed size in bytes.
	MaxBytes *int `json:"maxBytes,omitempty"`
}

type UploadLimitExceededFailureResponse struct {
	Message *string `json:"message,omitempty"`
	// The maximum number of uploads allowed per window.
	MaxUploads *int `json:"maxUploads,omitempty"`
	// The number of hours in the upload limit window.
	WindowHours *int `json:"windowHours,omitempty"`
}

type VariantWorkflowNode struct {
	ID          string   `json:"id"`
	WorkflowID  string   `json:"workflowId"`
	TypeName    string   `json:"typeName"`
	NextNodeIDs []string `json:"nextNodeIds"`
	VariantID   *string  `json:"variantId,omitempty"`
	IsControl   *bool    `json:"isControl,omitempty"`
}

type WorkflowContactPropertyComparison struct {
	Value    any    `json:"value"`
	Operator string `json:"operator"`
}

type WorkflowContactPropertyQuery struct {
	Key string                            `json:"key"`
	Is  WorkflowContactPropertyComparison `json:"is"`
	Was WorkflowContactPropertyComparison `json:"was"`
}

type WorkflowEventProperty struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type WorkflowExperimentType string

const (
	WorkflowExperimentTypeWebhook   WorkflowExperimentType = "webhook"
	WorkflowExperimentTypeAutosplit WorkflowExperimentType = "autosplit"
)

type WorkflowFailureResponse struct {
	Message string `json:"message"`
}

type WorkflowNode map[string]any

type WorkflowSummary struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type WorkflowTimerUnit string

const (
	WorkflowTimerUnitM WorkflowTimerUnit = "m"
	WorkflowTimerUnitH WorkflowTimerUnit = "h"
	WorkflowTimerUnitD WorkflowTimerUnit = "d"
	WorkflowTimerUnitS WorkflowTimerUnit = "s"
)

type (
	ContactPropertyCreate  = ContactPropertyCreateRequest
	TransactionalEmailInfo = TransactionalEmail
	TransactionalEmailList = ListTransactionalsResponse
)

type APIKeyInfo struct {
	Success  bool   `json:"success"`
	TeamName string `json:"teamName"`
}

type errorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
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
