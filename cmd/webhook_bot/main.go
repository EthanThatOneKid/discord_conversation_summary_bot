// Discord conversation summary bot implemented to send data in webhook to Discord channel.
package main

import (
	"context"
	"log"
	"os"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/session"
	"github.com/joho/godotenv"
	"libdb.so/persist"
	"libdb.so/persist/driver/badgerdb"

	"github.com/ethanthatonekid/discord_conversation_summary_bot/pkg/bot"
	"github.com/ethanthatonekid/discord_conversation_summary_bot/pkg/store"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalln("Error loading .env file:", err)
	}

	storePath := os.Getenv("DB_PATH")
	if storePath == "" {
		log.Fatalln("No $DB_PATH given. Use :memory: for in-memory storage.")
	}

	m, err := persist.NewMustMap[string, store.Summary](badgerdb.Open, storePath)
	if err != nil {
		log.Fatalln("cannot create badgerdb-backed map:", err)
	}
	defer m.Close()

	st := store.NewStore(&m)

	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		log.Fatalln("No $DISCORD_BOT_TOKEN given.")
	}

	s := session.New("Bot " + token)

	// Add the needed Gateway intents.
	s.AddIntents(gateway.IntentGuildMessages)
	// s.AddIntents(gateway.IntentDirectMessages)

	bot.SetupBot(bot.SetupBotOptions{
		Session: s,
		Store:   st,
	})
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
