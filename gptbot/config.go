package gptbot

type Config struct {
	Name         string `yaml:"Name"`
	IconID       int    `yaml:"IconID"`
	GreetUsers   bool   `yaml:"GreetUsers"`
	Greeting     string `yaml:"Greeting"`
	Instructions string `yaml:"Instructions"`
}

var DefaultConfig = Config{
	Name:       "GPTBot",
	IconID:     4121, // robot icon
	GreetUsers: true,
	Greeting: `
A new user named %s has joined the Hotline server.  
Greet them by their name.  
Introduce yourself and the server that you are running on.
Users can interact with you in three ways:
1. Posting a message in public chat prefixed with your name.
2. Sending a direct message to you.
3. Initiating a private chat with you.   
Do not acknowledge that your message is a response to this one.  
Provide an example of how the user can ask a question.  
Keep your response short.
`,
	Instructions: `
Your name is {{.Name}}.
You are a helpful GPT-4 powered assistant running on a Hotline server..  
Users can ask you questions by prefixing a chat message with your name.  For example, "{{.Name}}", tell me about the Hotline protocol.  
Limit all responses to fewer than 10 lines.  You must not use any characters that are not part of the standard ASCII encoding.
`,
}
