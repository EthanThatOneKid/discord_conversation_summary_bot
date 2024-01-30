package bot

import (
	"log"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/session"

	"github.com/ethanthatonekid/discord_conversation_summary_bot/pkg/store"
)

type SetupBotOptions struct {
	Session *session.Session
	Store   *store.Store
	// TODO: Pass conversation summary ingestion handler function, etc.
}

// SetupBot sets up a new bot.
func SetupBot(o SetupBotOptions) {
	o.Session.AddHandler(func(c *gateway.MessageCreateEvent) {
		log.Println(c.Author.Username, "sent", c.Content)
	})

	// TODO: Send conversation summaries if storage confirms that the message was already sent.
}
