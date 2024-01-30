// Discord conversation summary bot implemented to send data in webhook to Discord channel.
package main

import (
	"context"
	"log"
	"os"

	"github.com/diamondburned/arikawa/v3/session"
	"github.com/ethanthatonekid/discord_conversation_summary_bot/pkg/bot"
)

func main() {
	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		log.Fatalln("No $DISCORD_BOT_TOKEN given.")
	}

	s := session.New("Bot " + token)

	// Add the needed Gateway intents.
	// s.AddIntents(gateway.IntentGuildMessages)
	// s.AddIntents(gateway.IntentDirectMessages)

	bot.SetupBot(s)
	if err := s.Open(context.Background()); err != nil {
		log.Fatalln("Failed to connect:", err)
	}
	defer s.Close()

	u, err := s.Me()
	if err != nil {
		log.Fatalln("Failed to get myself:", err)
	}

	log.Println("Started as", u.Username)

	// Block forever.
	select {}
}
