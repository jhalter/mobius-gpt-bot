package main

import (
	"context"
	"github.com/Netflix/go-env"
	"github.com/jhalter/mobius/hotline"
	"github.com/sashabaranov/go-openai"
	"hotline-chat-gpt-bot/gptbot"
	"log"
	"log/slog"
	"os"
	"time"
)

type Environment struct {
	APIKey     string `env:"OPENAI_API_KEY,required=true"`
	ServerAddr string `env:"SERVER_ADDR"`
	Login      string `env:"SERVER_LOGIN,default=guest"`
	Password   string `env:"SERVER_PASS"`

	BotConfig gptbot.Config
}

// Value swapped in by go-releaser at build time
var version = "dev"

func main() {
	var environment Environment
	_, err := env.UnmarshalFromEnviron(&environment)
	if err != nil {
		log.Fatal(err)
	}

	logger := slog.New(
		slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{
				ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
					if a.Key == slog.TimeKey {
						// Remove the milliseconds from the time field to save a few columns.
						a.Value = slog.StringValue(a.Value.Time().Format(time.RFC3339))
					}
					return a
				},
			},
		),
	)

	ctx := context.Background()
	bot, err := gptbot.New(
		ctx,
		environment.BotConfig,
		openai.NewClientWithConfig(openai.DefaultConfig(environment.APIKey)),
		logger,
	)
	if err != nil {
		slog.Error("Error initializing bot", "error", err)
		os.Exit(1)
	}

	bot.HotlineClient.Pref.IconID = environment.BotConfig.IconID

	// Register transaction handlers for transaction types that we should act on.
	bot.HotlineClient.HandleFunc(hotline.TranChatMsg, bot.HandleClientChatMsg)
	bot.HotlineClient.HandleFunc(hotline.TranNotifyChangeUser, bot.TranNotifyChangeUser)
	bot.HotlineClient.HandleFunc(hotline.TranNotifyChatDeleteUser, bot.HandleNotifyDeleteUser)
	bot.HotlineClient.HandleFunc(hotline.TranGetUserNameList, bot.HandleClientGetUserNameList)
	bot.HotlineClient.HandleFunc(hotline.TranKeepAlive, bot.HandleKeepAlive)
	bot.HotlineClient.HandleFunc(hotline.TranGetMsgs, bot.TranGetMsgs)
	bot.HotlineClient.HandleFunc(hotline.TranServerMsg, bot.HandleServerMsg)
	bot.HotlineClient.HandleFunc(hotline.TranInviteToChat, bot.HandleInviteToChat)
	bot.HotlineClient.HandleFunc(hotline.TranGetClientInfoText, bot.TranGetClientInfoText)

	// Register tool call functions.
	bot.HandleToolCallFunc("get_chat_log", bot.GetChatLog)
	bot.HandleToolCallFunc("get_user_log", bot.GetUserLog)
	bot.HandleToolCallFunc("get_flat_news", bot.GetFlatNews)
	bot.HandleToolCallFunc("get_release_info", bot.GetReleaseInfo)
	bot.HandleToolCallFunc("get_hotline_release_info", bot.GetHotlineReleaseInfo)

	logger.InfoContext(ctx, "Started Mobius GPT Bot", "version", version, "model", environment.BotConfig.Model)

	// Connect to the Hotline server.
	err = bot.HotlineClient.Connect(environment.ServerAddr, environment.Login, environment.Password)
	if err != nil {
		logger.Error("Hotline connection error", "error", err)
		os.Exit(1)
	}

	// Get the initial username list.
	if err = bot.HotlineClient.Send(hotline.NewTransaction(hotline.TranGetUserNameList, [2]byte{})); err != nil {
		logger.Error("Hotline connection error", "error", err)
		os.Exit(1)
	}

	// Get initial news posts so that we can answer questions related to news postings.
	if err = bot.HotlineClient.Send(hotline.NewTransaction(hotline.TranGetMsgs, [2]byte{})); err != nil {
		logger.Error("Hotline connection error", "error", err)
		os.Exit(1)
	}

	// Enter transaction handling loop until exit.
	if err = bot.HotlineClient.HandleTransactions(ctx); err != nil {
		logger.Error("Hotline connection error", "error", err)
		os.Exit(1)
	}
}
