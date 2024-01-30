package bot

import (
	"log"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/session"
)

// SetupBot sets up a new bot.
func SetupBot(
	s *session.Session,
	// TODO: Pass persistent storage and conversation summary ingestion handler function, etc.
) {
	s.AddHandler(func(c *gateway.MessageCreateEvent) {
		log.Println(c.Author.Username, "sent", c.Content)
	})

	// TODO: Send conversation summaries if storage confirms that the message was already sent.
}

// https://github.com/diamondburned/arikawa/blob/v3.3.3/0-examples/undeleter/main.go
// https://github.com/diamondburned/arikawa/blob/dbc4ae8978dd01ca94b5f7364438ef8d01ea58b4/gateway/events.go#L671
// https://pkg.go.dev/libdb.so/persist#NewMustMap
