// Discord conversation summary bot implemented to send data in webhook to Discord channel.
package main

import (
	"log"
	"os"

	"github.com/ethanthatonekid/discord_conversation_summary_bot/pkg/bot"
)

func main() {
	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		log.Fatalln("No $BOT_TOKEN given.")
	}

	b := bot.NewBot(token)
	defer b.Close()

	// Block forever.
	select {}
}
