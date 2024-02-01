// Discord conversation summary bot implemented to send data in webhook to Discord channel.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/diamondburned/arikawa/v3/api/webhook"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/session"
)

func formatMention(id discord.UserID) string {
	return fmt.Sprintf("<@%s>", id.String())
}

func formatMentions(userIDs []discord.UserID) string {
	mentions := []string{}
	for _, id := range userIDs {
		mentions = append(mentions, formatMention(id))
	}

	return strings.Join(mentions, " ")
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
			URL:         formatMessageURL(guildID, channelID, summary.StartID),
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

// webhookConfig is the configuration for a webhook.
type webhookConfig struct {
	// URL is the webhook URL to send the conversation summaries to.
	URL string `json:"url"`

	// GuildIDs is the list of guild IDs allowed to send conversation summaries to the webhook.
	// If empty, all guilds are allowed.
	GuildIDs []discord.GuildID `json:"guild_ids"`

	// ChannelIDs is the list of channel IDs allowed to send conversation summaries to the webhook.
	// If empty, all channels are allowed.
	ChannelIDs []discord.ChannelID `json:"channel_ids"`
}

// config is the configuration for the bot.
type config struct {
	Token    string          `json:"token"`
	Webhooks []webhookConfig `json:"webhooks"`
}

func (c *config) unmarshal(data []byte) error {
	return json.Unmarshal(data, c)
}

func mustConfig(path string) *config {
	file, err := os.ReadFile(path)
	if err != nil {
		log.Fatalln("Failed to read config file:", err)
	}

	var c config
	if err := c.unmarshal(file); err != nil {
		log.Fatalln("Failed to unmarshal config:", err)
	}

	return &c
}

func webhookURLsByEvent(c *config, event *gateway.ConversationSummaryUpdateEvent) []string {
	urls := []string{}
	for _, wc := range c.Webhooks {
		if len(wc.GuildIDs) > 0 && !slices.Contains(wc.GuildIDs, event.GuildID) {
			continue
		}

		if len(wc.ChannelIDs) > 0 && !slices.Contains(wc.ChannelIDs, event.ChannelID) {
			continue
		}

		urls = append(urls, wc.URL)
	}

	return urls
}

func main() {
	c := mustConfig("config.json")
	s := session.New(c.Token)

	// Add the needed Gateway intents.
	s.AddIntents(gateway.IntentGuildMessages)

	// Add the conversation summary update event handler.
	s.AddHandler(func(event *gateway.ConversationSummaryUpdateEvent) {
		webhookURLs := webhookURLsByEvent(c, event)
		var wg sync.WaitGroup
		for _, webhookURL := range webhookURLs {
			wg.Add(1)
			go func(webhookURL string) {
				messages, err := executeWebhooksWithEvent(webhookURL, event)
				if err != nil {
					log.Println("Failed to execute webhook:", err)
					return
				}

				log.Println("Executed webhook(s):")
				for _, m := range messages {
					log.Println(m.ID)
				}
			}(webhookURL)
		}
		wg.Wait()
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
