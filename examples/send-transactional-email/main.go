package main

import (
	"context"
	"github.com/tilebox/loops-go"
	"log/slog"
)

func main() {
	client, err := loops.NewClient(loops.WithApiKey("YOUR_LOOPS_API_KEY"))
	if err != nil {
		slog.Error("failed to create client", slog.Any("error", err.Error()))
		return
	}

	ctx := context.Background()

	err = client.SendTransactionalEmail(ctx, &loops.TransactionalEmail{
		TransactionalId: "cm3n2vjux00cgeyeflew9ly2w",
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
