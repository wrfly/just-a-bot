package client

import (
	"context"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/go-redis/redis"
	"github.com/google/go-github/v24/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type Cli struct {
	cli *github.Client
}

func New(token string, redisCli *redis.Client) *Cli {
	if token == "" {
		return &Cli{github.NewClient(nil)}
	}

	ctx := context.WithValue(context.Background(),
		oauth2.HTTPClient,
		&http.Client{
			Transport: &roundTripper{
				tp:    http.DefaultTransport,
				redis: redisCli,
			},
		},
	)

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	return &Cli{
		cli: github.NewClient(oauth2.NewClient(ctx, ts)),
	}
}

// TODO: add cache here
func (c *Cli) followers(ctx context.Context, user string) <-chan *github.User {
	ch := make(chan *github.User)
	go func() {
		for i := 1; ; i++ {
			users, _, err := c.cli.Users.
				ListFollowers(ctx, user, &github.ListOptions{
					Page:    i,
					PerPage: 100,
				})
			if err != nil {
				if i == 1 {
					logrus.Errorf("list [%s] followers error: %s", user, err)
				}
				break
			}
			if len(users) == 0 {
				break
			}
			for _, u := range users {
				ch <- u
			}
		}
		close(ch)
	}()
	return ch
}

// TODO: add cache here
func (c *Cli) following(ctx context.Context, user string) <-chan *github.User {
	ch := make(chan *github.User)
	go func() {
		for i := 1; ; i++ {
			users, _, err := c.cli.Users.
				ListFollowing(ctx, user, &github.ListOptions{
					Page:    i,
					PerPage: 100,
				})
			if err != nil {
				if i == 1 {
					logrus.Errorf("list [%s] following error: %s", user, err)
				}
				break
			}
			if len(users) == 0 {
				break
			}
			for _, u := range users {
				ch <- u
			}
		}
		close(ch)
	}()

	return ch
}

func (c *Cli) RelatedUsers(ctx context.Context, user string) []string {
	var wg sync.WaitGroup
	users := make(chan *github.User)

	wg.Add(2)
	go func() {
		for user := range c.followers(ctx, user) {
			users <- user
		}
		wg.Done()
	}()
	go func() {
		for user := range c.following(ctx, user) {
			users <- user
		}
		wg.Done()
	}()

	relatedMap := make(map[string]bool)
	go func() {
		for user := range users {
			relatedMap[user.GetLogin()] = true
		}
	}()

	wg.Wait()

	related := make([]string, 0, len(relatedMap))
	for user := range relatedMap {
		related = append(related, user)
	}

	logrus.Infof("%s related users %d", user, len(related))

	return related
}

func (c *Cli) Follow(user string) {
	logrus.Infof("follow %s", user)
	resp, err := c.cli.Users.Follow(context.Background(), user)
	if err != nil {
		bs, _ := ioutil.ReadAll(resp.Body)
		logrus.Errorf("follow %s err: %s [%s]", user, bs, err)
	}
}
