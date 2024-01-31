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
)

func formatMention(id discord.UserID) string {
	return fmt.Sprintf("<@%s>", id.String())
}

func formatMentions(userIDs []discord.UserID) string {
	mentions := []string{}
	for _, id := range userIDs {
		mentions = append(mentions, formatMention(id))
	}

	return strings.Join(mentions, "")
}

func formatMessageURL(guildID discord.GuildID, channelID discord.ChannelID, messageID discord.MessageID) string {
	return fmt.Sprintf("https://discord.com/channels/%d/%d/%d", guildID, channelID, messageID)
}

func formatSummaryRange(guildID discord.GuildID, channelID discord.ChannelID, summary gateway.ConversationSummary) string {
	m1 := formatMessageURL(guildID, channelID, summary.StartID)
	m2 := formatMessageURL(guildID, channelID, summary.EndID)
	amountBetween := summary.Count - 2
	if amountBetween > 0 {
		return fmt.Sprintf("%s +%d → %s", m1, amountBetween, m2)
	}

	return fmt.Sprintf("%s → %s", m1, m2)
}

func formatPeopleEmbedFieldName(amountPeople int) string {
	if amountPeople == 1 {
		return "Person"
	}

	return fmt.Sprintf("%d people", amountPeople)
}

func formatMessagesEmbedFieldName(amountMessages int) string {
	if amountMessages == 1 {
		return "Message"
	}

	return fmt.Sprintf("%d messages", amountMessages)
}

func makeExecuteDataWithSummaries(guildID discord.GuildID, channelID discord.ChannelID, summaries []gateway.ConversationSummary) webhook.ExecuteData {
	embeds := []discord.Embed{}
	for _, summary := range summaries {
		embeds = append(embeds, discord.Embed{
			Title:       summary.Topic,
			Description: summary.ShortSummary,
			Fields: []discord.EmbedField{
				{Name: formatPeopleEmbedFieldName(len(summary.People)),
					Value: formatMentions(summary.People)},
				{Name: formatMessagesEmbedFieldName(summary.Count),
					Value: formatSummaryRange(guildID, channelID, summary)},
			},
			Footer: &discord.EmbedFooter{
				Text: fmt.Sprintf("Summary ID: %d", summary.ID),
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

func executeWebhooksWithEvent(webhookURL string, event *gateway.ConversationSummaryUpdateEvent) ([]*discord.Message, error) {
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

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalln("Error loading .env file:", err)
	}

	token := os.Getenv("DISCORD_USER_TOKEN")
	if token == "" {
		log.Fatalln("No $DISCORD_USER_TOKEN given.")
	}

	s := session.New(token)

	// Add the needed Gateway intents.
	s.AddIntents(gateway.IntentGuildMessages)

	webhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	if webhookURL == "" {
		log.Fatalln("No $DISCORD_WEBHOOK_URL given.")
	}

	// Add the conversation summary update event handler.
	s.AddHandler(func(event *gateway.ConversationSummaryUpdateEvent) {
		messages, err := executeWebhooksWithEvent(webhookURL, event)
		if err != nil {
			log.Println("Failed to execute webhook:", err)
			return
		}

		log.Println("Executed webhook(s):")
		for _, m := range messages {
			log.Println(m.ID)
		}
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
