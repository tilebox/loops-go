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

	// create a contact
	contactID, err := client.CreateContact(ctx, &loops.Contact{
		Email:      "neil.armstrong@moon.space",
		FirstName:  loops.String("Neil"),
		LastName:   loops.String("Armstrong"),
		Subscribed: true,
	})
	if err != nil {
		slog.Error("failed to create contact", slog.Any("error", err.Error()))
		return
	}
	slog.Info("Created contact", slog.String("id", contactID))

	// find a contact
	contact, err := client.FindContact(ctx, &loops.ContactIdentifier{
		Email: loops.String("neil.armstrong@moon.space"),
	})
	if err != nil {
		slog.Error("failed to find contact", slog.Any("error", err.Error()))
		return
	}
	slog.Info("Found contact", slog.String("id", contact.ID), slog.String("email", contact.Email))

	// update a contact, specify a user group
	_, err = client.UpdateContact(ctx, &loops.Contact{
		Email:     "neil.armstrong@moon.space",
		UserGroup: loops.String("Astronauts"),
	})
	if err != nil {
		slog.Error("failed to update contact", slog.Any("error", err.Error()))
		return
	}

	// delete a contact
	err = client.DeleteContact(ctx, &loops.ContactIdentifier{
		Email: loops.String("neil.armstrong@moon.space"),
	})
	if err != nil {
		slog.Error("failed to delete contact", slog.Any("error", err.Error()))
		return
	}
}
