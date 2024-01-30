package bot

import (
	"log"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/session"
)

// ConversationSummaryUpdateEventHandler is a function that handles a ConversationSummaryUpdateEvent.
// Returns bool pointer that indicates whether or not the event should be recorded as successfully delivered.
type ConversationSummaryUpdateEventHandler func(*gateway.ConversationSummaryUpdateEvent)

// Options is a set of options for a conversation summary bot.
type Options struct {
	Session                          *session.Session
	OnConversationSummaryUpdateEvent ConversationSummaryUpdateEventHandler
}

// Setup sets up a new bot.
func Setup(o Options) {
	o.Session.AddHandler(func(c *gateway.MessageCreateEvent) {
		log.Println(c.Author.Username, "sent", c.Message.Content)
	})

	o.Session.AddHandler(func(c *gateway.ConversationSummaryUpdateEvent) {
		for _, summary := range c.Summaries {
			log.Printf("Conversation summary update event: %s", summary.ID)
			log.Printf("Summary: %s", summary.ShortSummary)
		}

		o.OnConversationSummaryUpdateEvent(c)
	})
}
