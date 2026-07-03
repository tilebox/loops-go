package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/tilebox/loops-go"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))

	client, err := loops.NewClient(loops.WithAPIKey("YOUR_LOOPS_API_KEY"))
	if err != nil {
		slog.Error("failed to create client", slog.Any("error", err.Error()))
		return
	}

	ctx := context.Background()

	transactionalID := os.Getenv("LOOPS_TRANSACTIONAL_ID")
	if transactionalID == "" {
		transactional, err := client.CreateTransactionalEmail(ctx, &loops.CreateTransactionalRequest{
			Name: "SDK lifecycle example",
		})
		if err != nil {
			slog.Error("failed to create transactional email", slog.Any("error", err.Error()))
			return
		}
		transactionalID = transactional.ID
		slog.Info("created transactional email", slog.String("id", transactional.ID), slog.String("name", transactional.Name))
	} else {
		slog.Info("using existing transactional email", slog.String("id", transactionalID))
	}

	draft, err := client.EnsureTransactionalEmailDraft(ctx, transactionalID)
	if err != nil {
		slog.Error("failed to ensure transactional email draft", slog.Any("error", err.Error()))
		return
	}
	if draft.DraftEmailMessageID == nil {
		slog.Error("transactional email has no draft email message")
		return
	}
	slog.Info("ensured draft", slog.String("emailMessageId", *draft.DraftEmailMessageID))

	message, err := client.GetEmailMessage(ctx, *draft.DraftEmailMessageID)
	if err != nil {
		slog.Error("failed to fetch draft email message", slog.Any("error", err.Error()))
		return
	}
	slog.Info("fetched draft email message", slog.String("id", message.ID), slog.String("revision", value(message.ContentRevisionID)))

	updated, err := client.UpdateEmailMessage(ctx, message.ID, &loops.UpdateEmailMessageRequest{
		ExpectedRevisionID: message.ContentRevisionID,
		Subject:            loops.String("Welcome to the mission, {{name}}"),
		PreviewText:        loops.String("Your mission briefing is ready."),
		LMX: loops.String(`<Email>
  <Text>Hello {{name}},</Text>
  <Text>Your mission briefing is ready.</Text>
</Email>`),
	})
	if err != nil {
		slog.Error("failed to update draft email message", slog.Any("error", err.Error()))
		return
	}
	slog.Info("updated draft email message", slog.String("id", updated.ID), slog.String("revision", value(updated.ContentRevisionID)))

	guardian, err := client.GetEmailMessageGuardian(ctx, updated.ID)
	if err != nil {
		slog.Error("failed to run Guardian checks", slog.Any("error", err.Error()))
		return
	}
	slog.Info("ran Guardian checks", slog.Int("errors", len(guardian.Errors)), slog.Int("warnings", len(guardian.Warnings)))

	// Sending previews and publishing are intentionally opt-in so this example is safe to run while experimenting.
	if previewEmail := os.Getenv("LOOPS_PREVIEW_EMAIL"); previewEmail != "" {
		preview, err := client.SendEmailMessagePreview(ctx, updated.ID, &loops.EmailMessagePreviewRequest{
			Emails: []string{previewEmail},
			DataVariables: map[string]any{
				"name": "Neil",
			},
		})
		if err != nil {
			slog.Error("failed to send preview", slog.Any("error", err.Error()))
			return
		}
		slog.Info("sent preview", slog.String("id", preview.ID), slog.String("email", previewEmail))
	}

	if os.Getenv("LOOPS_PUBLISH_TRANSACTIONAL") == "true" {
		published, err := client.PublishTransactionalEmailDraft(ctx, transactionalID)
		if err != nil {
			slog.Error("failed to publish transactional email draft", slog.Any("error", err.Error()))
			return
		}
		slog.Info("published transactional email", slog.String("id", published.ID))
	} else {
		slog.Info("skipped publish", slog.String("set", "LOOPS_PUBLISH_TRANSACTIONAL=true"))
	}
}

func value(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
