# Loops GO SDK

## Introduction

A Go SDK for interacting with [Loops's](https://loops.so) API.

## API Documentation

- Official Loops API documentation: [Loops API reference](https://app.loops.so/docs/api-reference/)
- This client targets the vendored [openapi.json](openapi.json) spec version `1.16.0`. For the latest Loops OpenAPI spec, see [https://app.loops.so/openapi.json](https://app.loops.so/openapi.json).

## Contributing

Contributions are welcome! If the Loops API changes, community PRs that update [openapi.json](openapi.json), models, client methods, examples, and endpoint tests are very welcome.

## Installation

```bash
go get github.com/tilebox/loops-go
```

## Usage

Below are a few examples of how to use the SDK to send API requests.
For some full, working examples, see the [examples](examples) directory.

**Create a client**:

```go
package main

import (
	"context"
	"log/slog"

	"github.com/tilebox/loops-go"
)

func main() {
	ctx := context.Background()
	client, err := loops.NewClient(loops.WithAPIKey("YOUR_LOOPS_API_KEY"))
	if err != nil {
		slog.Error("failed to create client", slog.Any("error", err.Error()))
		return
	}

	// now use the client to make requests
}
```

### Contacts

**Create a contact**
```go
contactID, err := client.CreateContact(ctx, &loops.Contact{
    Email:      "neil.armstrong@moon.space",
    FirstName:  loops.String("Neil"),
    LastName:   loops.String("Armstrong"),
    UserGroup:  loops.String("Astronauts"),
    Subscribed: true,
    // custom user defined properties for contacts
    Properties: map[string]any{
        "role": "Astronaut",
    },
})
if err != nil {
    slog.Error("failed to create contact", slog.Any("error", err.Error()))
    return
}
```

**Find a contact**
```go
contact, err := client.FindContact(ctx, &loops.ContactIdentifier{
    Email: loops.String("neil.armstrong@moon.space"),
})
if err != nil {
    slog.Error("failed to find contact", slog.Any("error", err.Error()))
    return
}
```

**Delete a contact**
```go
err = client.DeleteContact(ctx, &loops.ContactIdentifier{
    Email: loops.String("neil.armstrong@moon.space"),
})
if err != nil {
    slog.Error("failed to delete contact", slog.Any("error", err.Error()))
    return
}
```

**List contact properties**
```go
properties, err := client.GetContactProperties(ctx, loops.ContactPropertyListOptions{
    List: loops.ContactPropertyTypeCustom, // only return your team's custom properties
})
if err != nil {
    slog.Error("failed to get contact properties", slog.Any("error", err.Error()))
    return
}
```

### Events

**Send an event**
```go
err = client.SendEvent(ctx, &loops.Event{
    Email:     loops.String("neil.armstrong@moon.space"),
    EventName: "joinedMission",
    EventProperties: map[string]any{
        "mission": "Apollo 11",
    },
})
if err != nil {
    slog.Error("failed to send event", slog.Any("error", err.Error()))
    return
}
```

### Transactional emails

**Send a transactional email**

```go
err = client.SendTransactionalEmail(ctx, &loops.TransactionalRequest{
    TransactionalID: "cm...",
    Email:           "recipient@example.com",
    DataVariables: map[string]any{
        "name": "Recipient Name",
    },
})
if err != nil {
    slog.Error("failed to send transactional email", slog.Any("error", err.Error()))
    return
}
```

**List transactional emails**

```go
emailsPage, err := client.ListTransactionalEmails(ctx, loops.ListTransactionalEmailsOptions{
    PerPage: 10,
})
if err != nil {
    slog.Error("failed to list transactional emails", slog.Any("error", err.Error()))
    return
}

for _, email := range emailsPage.Data {
    slog.Info("transactional email", slog.String("id", email.ID), slog.String("name", email.Name))
}
```

For more complete flows, see the [transactional email lifecycle](examples/transactional-email-lifecycle) and [image upload](examples/upload-image-asset) examples.

## Authors

Created by [Tilebox](https://tilebox.com) - The Solar System’s #1 developer tool for space data management.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Development

### Testing

```bash
go test ./...
```

### Linting

```bash
golangci-lint run --fix  ./...
```
