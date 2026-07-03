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

	err = client.SendTransactionalEmail(ctx, &loops.TransactionalRequest{
		TransactionalID: "cm3n2vjux00cgeyeflew9ly2w",
		Email:           "neil.armstrong@moon.space",
		DataVariables: map[string]any{
			"name": "Mr. Armstrong",
		},
	})
	if err != nil {
		slog.Error("failed to send transactional email", slog.Any("error", err.Error()))
		return
	}
	slog.Info("sent transactional email")
}
