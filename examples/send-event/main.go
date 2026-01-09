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

	err = client.SendEvent(ctx, &loops.Event{
		Email:     loops.String("neil.armstrong@moon.space"),
		EventName: "joinedMission",
		EventProperties: &map[string]any{
			"mission": "Apollo 11",
		},
	})
	if err != nil {
		slog.Error("failed to send event", slog.Any("error", err.Error()))
		return
	}
	slog.Info("sent event")
}
