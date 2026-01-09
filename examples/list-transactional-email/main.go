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
	emailsPage, err := client.ListTransactionalEmails(ctx, loops.ListTransactionalEmailsOptions{
		PerPage: 10,
	})
	if err != nil {
		slog.Error("failed to list transactional emails", slog.Any("error", err.Error()))
		return
	}
	slog.Info("transactional emails summary", slog.Int("count", len(emailsPage.Data)))

	for _, email := range emailsPage.Data {
		slog.Info("transactional email", slog.String("id", email.ID), slog.String("name", email.Name))
	}
}
