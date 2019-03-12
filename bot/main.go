package main

import (
	"context"
	"flag"

	"github.com/wrfly/just_a_bot/client"
)

type config struct {
	Token string `default:"fake_token"`
	Redis string `default:"127.0.0.1:6379/1"`
}

var CONFIG = new(config)

func init() {
	flag.StringVar(&CONFIG.Token, "token", "<token>", "your github token")
	flag.StringVar(&CONFIG.Redis, "redis", "127.0.0.1:6379/3", "redis cache")
	flag.Parse()
}

func main() {
	c := client.New(CONFIG.Token)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	limitChan := make(chan struct{}, 10)
	userChan := make(chan string, 1000)

	go func() {
		for _, user := range c.RelatedUsers(ctx, "wrfly") {
			userChan <- user
		}
	}()

	go func() {
		for user := range userChan {
			c.Follow(user)

			limitChan <- struct{}{}
			go func(user string) {
				for _, user := range c.RelatedUsers(ctx, user) {
					userChan <- user
				}
				<-limitChan
			}(user)
		}
	}()

	<-make(chan bool)
}
