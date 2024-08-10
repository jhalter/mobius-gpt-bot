package gptbot

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jhalter/mobius/hotline"
	"github.com/sashabaranov/go-openai"
	"golang.org/x/time/rate"
	"log/slog"
	"os"
	"sync"
	"text/template"
	"time"
)

var contextTTL = 1 * time.Minute

type Bot struct {
	Assistant     openai.Assistant
	OpenAPIClient *openai.Client
	HotlineClient *hotline.Client
	Threads       map[int]openai.Thread    // Hotline chat ID -> OpenAI Thread
	PMThreads     map[uint16]openai.Thread // Hotline user ID -> OpenAI Thread

	Config Config

	lastInteraction    time.Time
	lastInteractionMUX sync.Mutex

	toolCallHandlers map[string]toolCallHandleFunc

	Users    map[string]user
	usersMUX sync.Mutex

	FlatNews []string
	ChatLogs []PubChatLog
	Visitors []Visitor
}

// Visitor represents an observed visit by a user
type Visitor struct {
	Username string
	Time     time.Time

	limiter *rate.Limiter
}

// PubChatLog represents an observed public chat message
type PubChatLog struct {
	Username string
	Message  string
	Time     time.Time
}

func New(ctx context.Context, config Config, oc *openai.Client, logger *slog.Logger) (Bot, error) {
	assistantList, err := oc.ListAssistants(ctx, nil, nil, nil, nil)
	if err != nil {
		return Bot{}, err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return Bot{}, err
	}
	assistantName := "Hotline assistant-" + hostname

	var botAssistant openai.Assistant
	for _, assistant := range assistantList.Assistants {
		if *assistant.Name == assistantName {
			botAssistant = assistant
		}
	}

	if botAssistant.ID != "" {
		_, err := oc.DeleteAssistant(ctx, botAssistant.ID)
		if err != nil {
			return Bot{}, err
		}
	}

	assistant := defaultAssistant
	assistant.Name = &assistantName
	assistant.Model = config.Model

	create := func(name, t string) *template.Template {
		return template.Must(template.New(name).Parse(t))
	}
	buf := new(bytes.Buffer)
	t2 := create("t2", config.Instructions)
	err = t2.Execute(buf, map[string]string{
		"Name": config.Name,
	})
	if err != nil {
		return Bot{}, fmt.Errorf("error rendering instructions: %w", err)
	}

	inst := buf.String()
	assistant.Instructions = &inst

	botAssistant, err = oc.CreateAssistant(ctx, assistant)
	if err != nil {
		return Bot{}, err
	}

	pubChatThread, err := oc.CreateThread(ctx, openai.ThreadRequest{})
	if err != nil {
		return Bot{}, err
	}

	return Bot{
		Assistant:        botAssistant,
		Threads:          map[int]openai.Thread{0: pubChatThread},
		Users:            make(map[string]user),
		PMThreads:        make(map[uint16]openai.Thread),
		OpenAPIClient:    oc,
		Config:           config,
		HotlineClient:    hotline.NewClient(config.Name, logger),
		toolCallHandlers: make(map[string]toolCallHandleFunc),
		lastInteraction:  time.Now(),
	}, nil
}
