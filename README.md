# discord_conversation_summary_bot

[![Go Reference](https://pkg.go.dev/badge/github.com/ethanthatonekid/discord_conversation_summary_bot.svg)](https://pkg.go.dev/github.com/ethanthatonekid/discord_conversation_summary_bot)

Discord conversation summary update event handler in Go with example bot
implementations.

## Development

Copy `.env.example` to `.env` and populate it with your bot's token and the
webhook URL to send messages to.

Run `go mod tidy` to install all required dependencies.

Invite the bot to your server with the following URL, replacing
`$CLIENT_ID` with your bot's client ID:

```
https://discord.com/oauth2/authorize?client_id=$CLIENT_ID&scope=bot
```

Start the bot with the following command:

```sh
go run cmd/webhook_bot/main.go
```

<!-- TODO: Document `ngrok` usage for local development. -->

---

Developed with 💖 by [**@EthanThatOneKid**](https://etok.codes/)
