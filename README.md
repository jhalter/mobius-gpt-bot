# Mobius-GPT-bot

Mobius-GPT-bot is an experimental, cross-platform [Hotline](https://en.wikipedia.org/wiki/Hotline_Communications) chat
bot powered by the OpenAI ChatGPT-4 [assistants](https://platform.openai.com/docs/api-reference/assistants) API.

## Features

Mobius-GPT-bot can greet visitors to a Hotline server and respond to user interactions with OpenAI ChatGPT-4 generated
responses.

Users can interact with the bot by:

1. Posting a message in public chat prefixed with the bot's name
2. Sending a direct message to the bot
3. Initiating a private chat with the bot

The bot can make calls to external sources as part of the response generation. This currently includes accessing the
Hotline server news, public chat history, and visitor history.

This enables interactions like:

* _Summarize the recent chat history_
* _Translate the last news post to Spanish_
* _Who has visited the server recently?_

## ⚠️ Warning ⚠️

This software depends on the commercial [OpenAI](https://platform.openai.com/overview) ChatGPT APIs and costs money to
run and operate. The exact costs vary depending on a number of factors and may be difficult to predict. This software is
currently in an experimental phase with limited safeguards against abusive users and rife with bugs and inefficiencies,
so it's important that you set low spending limits on your OpenAI account to prevent billing surprises.

## Setup

The bot is configured through the following environment variables:

| Name               | Required |   Default   |                                                                                                                                                                                                                                                                            Description |
|:-------------------|----------|:-----------:|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------:|
| `OPENAI_API_KEY`   | X        |             |                                                                                                                                                                                                                                                                         OpenAI API Key |
| `OPENAI_MODEL`     |          | gpt-4o-mini |                                                                                                                                                                                OpenAI Model.  Refer to https://platform.openai.com/docs/models/gpt-4o for pricing and capability info. |
| `SERVER_ADDR`      | X        |             |                                                                                                                                                                                                                                     Server Address.  Example: mobius.trtphotl.com:5500 |
| `SERVER_LOGIN`     | X        |    guest    |                                                                                                                                                                                                                                                                  Hotline account login |
| `SERVER_PASS`      | X        |             |                                                                                                                                                                                                                                                               Hotline account password |
| `BOT_NAME`         |          |  Durandal   |                                                                                                                                                                                                                                                                   Name of Hotline user |
| `BOT_ICON`         |          |    4121     |                                                                                                                                                                                                   Hotline icon ID.  Refer to https://wiki.preterhuman.net/Hotline_Icons for valid IDs. |
| `BOT_INSTRUCTIONS` | X        |    False    | Instructions for the OpenAI assistant.  Example: You are Durandal, from the game Marathon. Respond as the character. The tone, mood,and formatting of your responses should accurately reflect your identity as Durandal. Keep your responses short. You are  slightly malfunctioning. |
| `BOT_GREET_USERS`  |          |    False    |                                                                                                                                                                                                                      Set to true to greet users who join the server. May be expensive! |
| `BOT_GREETING`     |          |             |                                                                                                                                                                                                           A new user named %s has joined the Hotline server. Greet them by their name. |

1. Create a new OpenAI [Project API key](https://platform.openai.com/api-keys) for use with the bot.
2. Define values for the required environment variables. Usage examples below will assume the required environment
   variables exist.
3. Create a Hotline user account with the following permissions:
    * Send Message
    * Private Chat
    * Public Chat
    * Can Get User Info
4. Acquire the binary through your preferred method:
    * Build it with `go build .`
    * Download a pre-compiled binary from the releases page
    * Run the Docker image
5. Run it:

* From binary:

```
./hotline-chat-gpt-bot
```

* From Docker image:

```
docker run --pull=always --rm \
-e OPENAI_API_KEY=$OPENAI_API_KEY \
-e SERVER_ADDR=$SERVER_ADDR \
-e SERVER_LOGIN=$SERVER_LOGIN \
-e SERVER_PASS=$SERVER_PASS \
-e BOT_INSTRUCTIONS=$BOT_INSTRUCTIONS \
-e BOT_NAME=$BOT_NAME \
-e OPENAI_MODEL=$OPENAI_MODEL \
-e BOT_GREETING=$BOT_GREETING \
-e BOT_GREET_USERS=$BOT_GREET_USERS \
ghcr.io/jhalter/mobius-gpt-bot:latest
```

## License

[MIT](https://raw.githubusercontent.com/jhalter/mobius-gpt-bot/master/LICENSE)
