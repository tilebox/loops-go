package main

import (
	"context"
	"log/slog"

	"github.com/tilebox/loops-go"
)

func main() {
	client, err := loops.NewClient(loops.WithAPIKey("YOUR_LOOPS_API_KEY"))
	if err != nil {
		slog.Error("failed to create client", slog.Any("error", err.Error()))
		return
	}

	ctx := context.Background()

	err = client.SendTransactionalEmail(ctx, &loops.TransactionalEmail{
		TransactionalID: "cm3n2vjux00cgeyeflew9ly2w",
		Email:           "lukas.bindreiter@tilebox.com",
		DataVariables: &map[string]interface{}{
			"name": "Mr. Lukas",
		},
	})
	if err != nil {
		slog.Error("failed to send transactional email", slog.Any("error", err.Error()))
		return
	}
	slog.Info("sent transactional email")
}
