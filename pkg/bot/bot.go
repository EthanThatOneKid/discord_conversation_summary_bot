package bot

import (
	"context"
	"log"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/session"
)

// NewBot creates a new bot with the given options.
func NewBot(
	token string,
	// TODO: Pass persistent storage and conversation summary ingestion handler function, etc.
) *session.Session {
	s := session.New("Bot " + token)
	s.AddHandler(func(c *gateway.MessageCreateEvent) {
		log.Println(c.Author.Username, "sent", c.Content)
	})

	// TODO: Send conversation summaries if storage confirms that the message was already sent.

	// Add the needed Gateway intents.
	// s.AddIntents(gateway.IntentGuildMessages)
	// s.AddIntents(gateway.IntentDirectMessages)

	if err := s.Open(context.Background()); err != nil {
		log.Fatalln("Failed to connect:", err)
	}

	u, err := s.Me()
	if err != nil {
		log.Fatalln("Failed to get myself:", err)
	}

	log.Println("Started as", u.Username)
	return s
}

// https://github.com/diamondburned/arikawa/blob/v3.3.3/0-examples/undeleter/main.go
// https://github.com/diamondburned/arikawa/blob/dbc4ae8978dd01ca94b5f7364438ef8d01ea58b4/gateway/events.go#L671
// https://pkg.go.dev/libdb.so/persist#NewMustMap
