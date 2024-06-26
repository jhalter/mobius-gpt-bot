package gptbot

import "golang.org/x/time/rate"

type user struct {
	account         string
	ipAddr          string
	limiter         *rate.Limiter
	greetingLimiter *rate.Limiter
}

// const perIPRateLimit = rate.Limit(0.0005) // ~ 2 greets per hour
const perIPRateLimit = rate.Limit(0.1005) // ~ 2 greets per hour

func NewUser(account string) user {
	return user{
		account:         account,
		greetingLimiter: rate.NewLimiter(perIPRateLimit, 1),
	}
}
