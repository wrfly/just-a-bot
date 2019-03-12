package client

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/go-redis/redis"
)

func TestClient(t *testing.T) {
	opt, _ := redis.ParseURL("redis://127.0.0.1:6379/5")
	rCli := redis.NewClient(opt)

	c := New(os.Getenv("TOKEN"), rCli)

	users := c.RelatedUsers(context.Background(), "wrfly")
	// for _, user := range users {
	// 	fmt.Println(user)
	// }
	fmt.Println(len(users))
	c.Follow("wrfly")
}
