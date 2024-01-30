package main

import (
	"fmt"
	"strings"

	"github.com/diamondburned/arikawa/v3/api/webhook"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
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

// func formatDescription(channelID discord.ChannelID, guildID discord.GuildID, messageID discord.MessageID) string {
// 	return fmt.Sprintf("↩ %s", formatMessageURL(event.ChannelID,  guildID, messageID))
// }

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

func executeWebhooksWithEvent(webhookURL string, event gateway.ConversationSummaryUpdateEvent) ([]*discord.Message, error) {
	webhookID, _, err := webhook.ParseURL(webhookURL)
	if err != nil {
		return nil, err
	}

	embedsLimit := 10
	messages := []*discord.Message{}
	for i := 0; i < len(event.Summaries); i += embedsLimit {
		summaries := event.Summaries[i : i+embedsLimit]
		m, err := executeWebhookWithSummaries(webhookURL, event.GuildID, event.ChannelID, summaries)
		if err != nil {
			return nil, err
		}

		if m.WebhookID != webhookID {
			return nil, fmt.Errorf("webhook ID does not match")
		}

		messages = append(messages, m)
	}

	return messages, nil
}
