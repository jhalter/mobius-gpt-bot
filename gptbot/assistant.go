package gptbot

import "github.com/sashabaranov/go-openai"

type funcProps struct {
	Properties struct {
	} `json:"properties"`
	Required []interface{} `json:"required"`
	Type     string        `json:"type"`
}

var defaultAssistant = openai.AssistantRequest{
	Model:       openai.GPT4TurboPreview,
	Description: nil,
	Tools: []openai.AssistantTool{
		{
			Type: openai.AssistantToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "get_user_log",
				Description: "Get list of visitors to the server",
				Parameters: funcProps{
					Properties: struct{}{},
					Required:   make([]interface{}, 0),
					Type:       "object",
				},
			},
		},
		{
			Type: openai.AssistantToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "get_chat_log",
				Description: "Get most recent chat messages",
				Parameters: funcProps{
					Properties: struct{}{},
					Required:   make([]interface{}, 0),
					Type:       "object",
				},
			},
		},
		{
			Type: openai.AssistantToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "get_flat_news",
				Description: "Get news/message board posts",
				Parameters: funcProps{
					Properties: struct{}{},
					Required:   make([]interface{}, 0),
					Type:       "object",
				},
			},
		},
		{
			Type: openai.AssistantToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "get_release_info",
				Description: "Get latest Mobius software release info",
				Parameters: funcProps{
					Properties: struct{}{},
					Required:   make([]interface{}, 0),
					Type:       "object",
				},
			},
		},
		{
			Type: openai.AssistantToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "get_hotline_release_info",
				Description: "Get latest Hotline software release info",
				Parameters: funcProps{
					Properties: struct{}{},
					Required:   make([]interface{}, 0),
					Type:       "object",
				},
			},
		},
	},
	FileIDs:  nil,
	Metadata: nil,
}
