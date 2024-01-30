// Discord conversation summary bot implemented to send data in webhook to Discord channel.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/diamondburned/arikawa/v3/api/webhook"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/session"
	"github.com/joho/godotenv"
	"libdb.so/persist"
	"libdb.so/persist/driver/badgerdb"

	"github.com/ethanthatonekid/discord_conversation_summary_bot/pkg/bot"
	"github.com/ethanthatonekid/discord_conversation_summary_bot/pkg/store"
)

func formatMention(id discord.UserID) string {
	return fmt.Sprintf("<@%s>", id.String())
}

func formatMentions(userIDs []discord.UserID) string {
	mentions := []string{}
	for _, id := range userIDs {
		mentions = append(mentions, formatMention(id))
	}

	return strings.Join(mentions, ", ")
}

func formatMessageURL(guildID discord.GuildID, channelID discord.ChannelID, messageID discord.MessageID) string {
	return fmt.Sprintf("https://discord.com/channels/%d/%d/%d", guildID, channelID, messageID)
}

func formatChannel(channelID discord.ChannelID) string {
	return fmt.Sprintf("<#%s>", channelID.String())
}

func formatSummaryRange(guildID discord.GuildID, channelID discord.ChannelID, summary gateway.ConversationSummary) string {
	m1 := formatMessageURL(guildID, channelID, summary.StartID)
	m2 := formatMessageURL(guildID, channelID, summary.EndID)
	amountBetween := summary.Count - 2
	if amountBetween > 0 {
		return fmt.Sprintf("%s\n[%d messages]\n%s", m1, amountBetween, m2)
	}

	return fmt.Sprintf("%s\n%s", m1, m2)
}

func makeExecuteDataWithSummaries(guildID discord.GuildID, channelID discord.ChannelID, summaries []gateway.ConversationSummary) webhook.ExecuteData {
	embeds := []discord.Embed{}
	for _, summary := range summaries {
		embeds = append(embeds, discord.Embed{
			Title: summary.ShortSummary,
			Fields: []discord.EmbedField{
				{Name: "Topic", Value: summary.Topic, Inline: true},
				{Name: "Channel", Value: formatChannel(channelID), Inline: true},
				{Name: "Messages", Value: formatSummaryRange(guildID, channelID, summary), Inline: true},
				{Name: "People", Value: formatMentions(summary.People), Inline: true},
			},
		})
	}

	return webhook.ExecuteData{Embeds: embeds}
}

func executeWebhook(webhookURL string, data webhook.ExecuteData) (*discord.Message, error) {
	c, err := webhook.NewFromURL(webhookURL)
	if err != nil {
		return nil, err
	}

	return c.ExecuteAndWait(data)
}

func executeWebhookWithSummaries(webhookURL string, guildID discord.GuildID, channelID discord.ChannelID, summaries []gateway.ConversationSummary) (*discord.Message, error) {
	return executeWebhook(webhookURL, makeExecuteDataWithSummaries(guildID, channelID, summaries))
}

func paginate[T any](s []T, pageSize int) (pages [][]T) {
	for i := 0; i < len(s); i += pageSize {
		pages = append(pages, s[i:min(i+pageSize, len(s))])
	}

	return
}

func executeWebhooksWithEvent(webhookURL string, event gateway.ConversationSummaryUpdateEvent) ([]*discord.Message, error) {
	webhookID, _, err := webhook.ParseURL(webhookURL)
	if err != nil {
		return nil, err
	}

	// 10 embeds is the limit per Discord message.
	// Each conversation summary is rendered as 1 embed.
	summaryGroups := paginate(event.Summaries, 10)
	messages := []*discord.Message{}
	for _, summaryGroup := range summaryGroups {
		m, err := executeWebhookWithSummaries(webhookURL, event.GuildID, event.ChannelID, summaryGroup)
		if err != nil {
			return nil, err
		}

		if m.WebhookID != webhookID {
			return messages, fmt.Errorf("webhook ID does not match")
		}

		messages = append(messages, m)
	}

	return messages, nil
}

func handleConversationSummaryUpdateEvent(webhookURL string, event gateway.ConversationSummaryUpdateEvent) {
	messages, err := executeWebhooksWithEvent(webhookURL, event)
	if err != nil {
		log.Println("Failed to execute webhook:", err)
		return
	}

	log.Println("Executed webhook(s):")
	for _, m := range messages {
		log.Println(m.ID)
	}
}

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

	webhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	if webhookURL == "" {
		log.Fatalln("No $DISCORD_WEBHOOK_URL given.")
	}

	bot.Setup(bot.Options{
		Session: s,
		Store:   st,
		OnConversationSummaryUpdateEvent: func(event gateway.ConversationSummaryUpdateEvent) {
			handleConversationSummaryUpdateEvent(webhookURL, event)
		},
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
