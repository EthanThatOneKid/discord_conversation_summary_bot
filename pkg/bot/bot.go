package bot

import (
	"log"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/session"

	"github.com/ethanthatonekid/discord_conversation_summary_bot/pkg/store"
)

// ConversationSummaryUpdateEventHandler is a function that handles a ConversationSummaryUpdateEvent.
// Returns bool pointer that indicates whether or not the event should be recorded as successfully delivered.
type ConversationSummaryUpdateEventHandler func(gateway.ConversationSummaryUpdateEvent)

// Options is a set of options for a conversation summary bot.
type Options struct {
	Session                          *session.Session
	Store                            *store.Store
	OnConversationSummaryUpdateEvent ConversationSummaryUpdateEventHandler
}

// Setup sets up a new bot.
func Setup(o Options) {
	o.Session.AddHandler(func(c *gateway.MessageCreateEvent) {
		log.Println(c.Author.Username, "sent", c.Content)
	})

	// TODO: Send conversation summaries if storage confirms that the message was already sent.
}
