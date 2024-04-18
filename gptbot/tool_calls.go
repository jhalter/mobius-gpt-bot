package gptbot

import (
	"context"
	"encoding/json"
	"github.com/sashabaranov/go-openai"
	"io"
	"net/http"
	"time"
)

type toolCallHandleFunc func(call openai.ToolCall) (openai.ToolOutput, error)

var runSleepInterval = 4 * time.Second

func (b *Bot) RunLoop(ctx context.Context, threadID string) (r openai.Run, err error) {
	run, err := b.OpenAPIClient.CreateRun(ctx, threadID, openai.RunRequest{
		AssistantID: b.Assistant.ID,
	})
	if err != nil {
		return r, err
	}

	b.HotlineClient.Logger.Info("CreateRun", "runID", run.ID, "threadID", run.ThreadID, "status", run.Status)

	var newRun openai.Run
	for {
		newRun, err = b.OpenAPIClient.RetrieveRun(ctx, run.ThreadID, run.ID)
		if err != nil {
			return r, err
		}

		if newRun.Status == openai.RunStatusCompleted {
			break
		}
		if newRun.Status == openai.RunStatusRequiresAction {
			var toolOutputs []openai.ToolOutput
			for _, toolCall := range newRun.RequiredAction.SubmitToolOutputs.ToolCalls {
				b.HotlineClient.Logger.InfoContext(ctx, "Required toolCall", "type", toolCall.Function.Name)

				toolOutput, err := b.toolCallHandlers[toolCall.Function.Name](toolCall)
				if err != nil {
					panic(err)
				}

				toolOutputs = append(toolOutputs, toolOutput)
			}
			_, err = b.OpenAPIClient.SubmitToolOutputs(ctx, newRun.ThreadID, newRun.ID, openai.SubmitToolOutputsRequest{
				ToolOutputs: toolOutputs,
			},
			)
			if err != nil {
				return r, err
			}
		}

		time.Sleep(runSleepInterval)
	}

	return newRun, err
}

func (b *Bot) GetUserLog(toolCall openai.ToolCall) (openai.ToolOutput, error) {
	jbytes, err := json.Marshal(b.Visitors)
	if err != nil {
		return openai.ToolOutput{}, err
	}

	return openai.ToolOutput{
		ToolCallID: toolCall.ID,
		Output:     jbytes,
	}, nil
}

func (b *Bot) GetChatLog(toolCall openai.ToolCall) (openai.ToolOutput, error) {
	var chatLogs []PubChatLog
	if len(b.ChatLogs) > 0 {
		chatLogs = b.ChatLogs[:len(b.ChatLogs)-1]
	}

	jbytes, err := json.Marshal(chatLogs)
	if err != nil {
		return openai.ToolOutput{}, err
	}

	return openai.ToolOutput{
		ToolCallID: toolCall.ID,
		Output:     jbytes,
	}, nil
}

func (b *Bot) GetFlatNews(toolCall openai.ToolCall) (openai.ToolOutput, error) {
	jbytes, err := json.Marshal(b.FlatNews)
	if err != nil {
		return openai.ToolOutput{}, err
	}

	return openai.ToolOutput{
		ToolCallID: toolCall.ID,
		Output:     jbytes,
	}, nil
}

type ReleaseInfo struct {
	Body      string `json:"body"`
	HtmlURL   string `json:"html_url"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

func (b *Bot) GetReleaseInfo(toolCall openai.ToolCall) (openai.ToolOutput, error) {
	resp, err := http.Get("https://api.github.com/repos/jhalter/mobius/releases/latest")
	if err != nil {
		return openai.ToolOutput{}, err
	}

	body, err := io.ReadAll(resp.Body)
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode > 299 {
		return openai.ToolOutput{}, err
	}

	var releaseInfo ReleaseInfo
	err = json.Unmarshal(body, &releaseInfo)
	if err != nil {
		return openai.ToolOutput{}, err
	}

	jBytes, err := json.Marshal(releaseInfo)
	if err != nil {
		return openai.ToolOutput{}, err
	}

	return openai.ToolOutput{
		ToolCallID: toolCall.ID,
		Output:     jBytes,
	}, nil
}

func (b *Bot) GetHotlineReleaseInfo(toolCall openai.ToolCall) (openai.ToolOutput, error) {
	resp, err := http.Get("https://api.github.com/repos/mierau/hotline/releases/latest")
	if err != nil {
		return openai.ToolOutput{}, err
	}

	body, err := io.ReadAll(resp.Body)
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode > 299 {
		return openai.ToolOutput{}, err
	}

	var releaseInfo ReleaseInfo
	err = json.Unmarshal(body, &releaseInfo)
	if err != nil {
		return openai.ToolOutput{}, err
	}

	jBytes, err := json.Marshal(releaseInfo)
	if err != nil {
		return openai.ToolOutput{}, err
	}

	return openai.ToolOutput{
		ToolCallID: toolCall.ID,
		Output:     jBytes,
	}, nil
}
