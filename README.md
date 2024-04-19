# Mobius-GPT-bot

Mobius-GPT-bot is an experimental, cross-platform [Hotline](https://en.wikipedia.org/wiki/Hotline_Communications) chat bot powered by the OpenAI ChatGPT-4 [assistants](https://platform.openai.com/docs/api-reference/assistants) API.

## Features

Mobius-GPT-bot can greet visitors to a Hotline server and respond to user interactions with OpenAI ChatGPT-4 generated responses.

Users can interact with the bot by:

1. Posting a message in public chat prefixed with the bot's name
2. Sending a direct message to the bot
3. Initiating a private chat with the bot

The bot can make calls to external sources as part of the response generation.  This currently includes accessing the Hotline server news, public chat history, and visitor history.

This enables interactions like:

* _Summarize the recent chat history_
* _Translate the last news post to Spanish_
* _Who has visited the server recently?_


## ⚠️ Warning  ⚠️

This software depends on the commercial [OpenAI](https://platform.openai.com/overview) ChatGPT-4 APIs and costs money to run and operate. The exact costs vary depending on a number of factors and may be difficult to predict. This software is currently in an experimental phase with limited safeguards against abusive users and rife with bugs and inefficiencies, so it's important that you set low spending limits on your OpenAI account to prevent billing surprises.

## Setup

1. Create a new OpenAI [Project API key](https://platform.openai.com/api-keys) and set it as the `OPENAI_API_KEY` environment variable.
2. (Optional) If you'd like to enable visitor greetings, create a new Hotline user account with the following permissions:
   * Send Message
   * Private Chat
   * Public Chat
   * Can Get User Info
3. (Optional) If you'd like to customize the name, icon, identity, greeting, etc, copy `example-config.yaml` from this repo and edit to your preference.
4. Acquire the binary through your preferred method:
   * Build it with `go build .`
   * Download a pre-compiled binary from the releases page 
   * Run the Docker image
5. Run it:
* From binary:

`./mobius-gpt-bot -server=your-server:5500 -login botuser -pass botpass`

* From Docker image:
```
export OPENAI_API_KEY=<your OPENAI_API_KEY>
docker run --pull=always --rm -e OPENAI_API_KEY=$OPENAI_API_KEY ghcr.io/jhalter/mobius-gpt-bot:latest -server=your-server:5500 -login botuser -pass botpass
```

To run the Docker image with a custom config file, use a bind mount to point to a directory on the host system that has the config.yaml:

```
docker run --pull=always --rm -e OPENAI_API_KEY=$OPENAI_API_KEY -v /home/jhalter/bot-config-dir:/usr/local/etc ghcr.io/jhalter/mobius-gpt-bot:latest -server=your-server:5500 -login botuser -pass botpass -config /usr/local/etc/config.yaml
```

## License

[MIT](https://raw.githubusercontent.com/jhalter/mobius-gpt-bot/master/LICENSE)
