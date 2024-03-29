# discord_conversation_summary_bot

[![Go Reference](https://pkg.go.dev/badge/github.com/ethanthatonekid/discord_conversation_summary_bot.svg)](https://pkg.go.dev/github.com/ethanthatonekid/discord_conversation_summary_bot)

Discord conversation summary update event handler in Go.

## Development

Populate a `config.json` file with your secrets.

```json
{
  "token": "$DISCORD_USER_TOKEN",
  "webhooks": [
    {
      "url": "$DISCORD_WEBHOOK_URL"
    }
  ]
}
```

> [!NOTE]
> To get your user token, follow these steps (adapted from [diamondburned/gtkcord4's README.md](https://github.com/diamondburned/gtkcord4/blob/main/README.md)):
>
> 1. Press <kbd>F12</kbd> with Discord open (to open the Inspector).
> 2. Go to the Network tab then press <kbd>F5</kbd> to refresh the page.
> 3. Search `discord api` then look for the `Authorization` header in the right
>    column.
> 4. Copy its value (the token).

> [!WARNING]
> Logging in using username/email and password is strongly discouraged. This
> method is untested and may cause your account to be banned! Prefer using the
> token method above.

> [!NOTE]
> Using an unofficial client at all is against Discord's Terms of Service and
> may cause your account to be banned! Use at your own risk!

Run `go mod tidy` to install all required dependencies.

Invite the bot to your server with the following URL, replacing
`$CLIENT_ID` with your bot's client ID (with [permissions](https://discordapi.com/permissions.html#65536)):

```
https://discord.com/oauth2/authorize?client_id=$CLIENT_ID&scope=bot&permissions=65536
```

Start the bot with the following command:

```sh
go run .
```

---

Developed with 💖 by [**@EthanThatOneKid**](https://etok.codes/)
