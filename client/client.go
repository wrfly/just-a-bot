package client

import (
	"context"

	"github.com/google/go-github/v24/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type Cli struct {
	cli *github.Client
}

func NewClient(token string) *Cli {
	if token == "" {
		return &Cli{github.NewClient(nil)}
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	return &Cli{
		cli: github.NewClient(
			oauth2.NewClient(
				context.Background(),
				ts,
			),
		),
	}
}

func (c *Cli) followers(ctx context.Context, user string) ([]*github.User, error) {
	all := []*github.User{}
	for i := 0; ; i++ {
		users, _, err := c.cli.Users.
			ListFollowers(ctx, user, &github.ListOptions{
				Page:    i,
				PerPage: 100,
			})
		if len(users) == 0 || err != nil {
			break
		}
		all = append(all, users...)
	}
	return all, nil
}

func (c *Cli) following(ctx context.Context, user string) ([]*github.User, error) {
	all := []*github.User{}
	for i := 0; ; i++ {
		users, _, err := c.cli.Users.
			ListFollowing(ctx, user, &github.ListOptions{
				Page:    i,
				PerPage: 100,
			})
		if len(users) == 0 || err != nil {
			break
		}
		all = append(all, users...)
	}
	return all, nil
}

func (c *Cli) RelatedUsers(ctx context.Context, startUser string) []string {
	users := []*github.User{}

	if x, err := c.followers(ctx, startUser); err != nil {
		logrus.Errorf("get [%s] follower error: %s", startUser, err)
	} else {
		users = append(users, x...)
	}

	if x, err := c.following(ctx, startUser); err != nil {
		logrus.Errorf("get [%s] following error: %s", startUser, err)
	} else {
		users = append(users, x...)
	}

	relatedMap := make(map[string]bool)
	for _, user := range users {
		relatedMap[user.GetLogin()] = true
	}

	related := make([]string, 0, len(relatedMap))
	for user := range relatedMap {
		related = append(related, user)
	}

	return related
}
