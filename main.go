package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/jhalter/mobius/hotline"
	"github.com/sashabaranov/go-openai"
	"gopkg.in/yaml.v3"
	"hotline-chat-gpt-bot/gptbot"
	"log/slog"
	"os"
)

func main() {
	srvAddr := flag.String("server", "localhost:5600", "Hotline server hostname:port")
	login := flag.String("login", "guest", "Hotline server login")
	pass := flag.String("pass", "", "Hotline server password")
	logLevel := flag.String("log-level", "info", "Log level")
	config := flag.String("config", "", "Path to config file")
	version := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *version {
		fmt.Printf("v%s\n", "TODO: Embed version during build")
		os.Exit(0)
	}

	if os.Getenv("OPENAI_API_KEY") == "" {
		fmt.Println("Missing OPENAI_API_KEY environment variable.")
		os.Exit(1)
	}

	logger := slog.New(
		slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: logLevels[*logLevel]},
		),
	)

	ctx := context.Background()
	// TODO: Implement context cancellation.
	// trap Ctrl+C and cancel the context
	//ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, os.Interrupt)
	//defer cancel()

	var botConfig gptbot.Config
	if *config != "" {
		fh, err := os.Open(*config)
		if err != nil {
			panic(err)
		}

		decoder := yaml.NewDecoder(fh)
		err = decoder.Decode(&botConfig)
		if err != nil {
			panic(err)
		}
	} else {
		botConfig = gptbot.DefaultConfig
	}

	openAIConfig := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	openAIConfig.AssistantVersion = "v2"

	bot, err := gptbot.New(
		ctx,
		botConfig,
		openai.NewClientWithConfig(openAIConfig),
		logger,
	)
	if err != nil {
		slog.Error("Error initializing bot", "error", err)
		os.Exit(1)
	}

	bot.HotlineClient.Pref.IconID = botConfig.IconID

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

	logger.InfoContext(ctx, "Started Mobius GPT Bot")

	// Connect to the Hotline server.
	err = bot.HotlineClient.Connect(*srvAddr, *login, *pass)
	if err != nil {
		logger.Error("Hotline connection error", "error", err)
		os.Exit(1)
	}

	// Get the initial username list.
	if err = bot.HotlineClient.Send(*hotline.NewTransaction(hotline.TranGetUserNameList, nil)); err != nil {
		logger.Error("Hotline connection error", "error", err)
		os.Exit(1)
	}

	// Get initial news posts so that we can answer questions related to news postings.
	if err = bot.HotlineClient.Send(*hotline.NewTransaction(hotline.TranGetMsgs, nil)); err != nil {
		logger.Error("Hotline connection error", "error", err)
		os.Exit(1)
	}

	// Enter transaction handling loop until exit.
	if err = bot.HotlineClient.HandleTransactions(ctx); err != nil {
		logger.Error("Hotline connection error", "error", err)
		os.Exit(1)
	}
}

var logLevels = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"error": slog.LevelError,
}
