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

	err = client.SendEvent(ctx, &loops.Event{
		Email:     loops.String("neil.armstrong@moon.space"),
		EventName: "joinedMission",
		EventProperties: &map[string]interface{}{
			"mission": "Apollo 11",
		},
	})
	if err != nil {
		slog.Error("failed to send event", slog.Any("error", err.Error()))
		return
	}
	slog.Info("sent event")
}
