# Loops GO SDK

## Introduction

A Go SDK for interacting with [Loops's](https://loops.so) API.

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
	"github.com/tilebox/loops-go"
	"log/slog"
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

**Create a contact**
```go
contactID, err := client.CreateContact(ctx, &loops.Contact{
    Email:      "neil.armstrong@moon.space",
    FirstName:  loops.String("Neil"),
    LastName:   loops.String("Armstrong"),
    UserGroup:  loops.String("Astronauts"),
    Subscribed: true,
})
if err != nil {
    slog.Error("failed to create contact", slog.Any("error", err.Error()))
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

### Events

**Send an event**
```go
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
```

### Transactional emails

**Send a transactional email**

```go
err = client.SendTransactionalEmail(ctx, &loops.TransactionalEmail{
    TransactionalId: "cm...",
    Email:           "recipient@example.com",
    DataVariables: &map[string]interface{}{
        "name": "Recipient Name",
    },
})
if err != nil {
    slog.Error("failed to send transactional email", slog.Any("error", err.Error()))
    return
}
```

## API Documentation

The API documentation is part of the official Loops Documentation and can be found [here](https://app.loops.so/docs/api-reference/).

## Contributing

Contributions are welcome! Especially if the loops API is updated, please feel free to open PRs for new or updated endpoints.

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