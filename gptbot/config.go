package gptbot

type Config struct {
	Name         string `yaml:"Name" env:"BOT_NAME,default=Durandal"`
	IconID       int    `yaml:"IconID" env:"BOT_ICON,default=4121"`
	GreetUsers   bool   `yaml:"GreetUsers" env:"BOT_GREET_USERS,default=false"`
	Greeting     string `yaml:"Greeting" env:"BOT_GREETING"`
	Instructions string `yaml:"Instructions" env:"BOT_INSTRUCTIONS"`

	Model string `env:"MODEL,default=gpt-4o-mini"`
}
