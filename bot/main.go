package main

import (
	"context"
	"flag"

	"github.com/sirupsen/logrus"

	"github.com/go-redis/redis"

	"github.com/wrfly/just_a_bot/client"
)

type config struct {
	Token string `default:"fake_token"`
	Redis string `default:"127.0.0.1:6379/1"`
}

var CONFIG = new(config)

func init() {
	flag.StringVar(&CONFIG.Token, "token", "<token>", "your github token")
	flag.StringVar(&CONFIG.Redis, "redis", "redis://127.0.0.1:6379/3", "redis cache")
	flag.Parse()
}

func main() {
	// build redis cli
	opts, err := redis.ParseURL(CONFIG.Redis)
	if err != nil {
		logrus.Fatalf("bad redis conn: %s", err)
	}
	redisCli := redis.NewClient(opts)
	if redisCli.Ping().Err() != nil {
		logrus.Fatal("bad redis, cannot ping")
	}

	c := client.New(CONFIG.Token, redisCli)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	userChan := make(chan string, 1e6)
	userChan <- "wrfly"

	nextUserChan := make(chan string, 1e6)

	go func() {
		for user := range nextUserChan {
			for _, user := range c.RelatedUsers(ctx, user) {
				userChan <- user
			}
		}
	}()

	go func() {
		for user := range userChan {
			c.Follow(user)
			nextUserChan <- user
		}
	}()

	<-make(chan bool)
}
