package gptbot

type Config struct {
	Name         string `env:"BOT_NAME,default=Durandal"`
	IconID       int    `env:"BOT_ICON,default=4121"`
	GreetUsers   bool   `env:"BOT_GREET_USERS,default=false"`
	Greeting     string `env:"BOT_GREETING"`
	Instructions string `env:"BOT_INSTRUCTIONS",default=""`
	Model        string `env:"OPENAI_MODEL,default=gpt-4o-mini"`
}
