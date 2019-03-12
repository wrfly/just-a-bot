package client

import (
	"context"
	"fmt"

	"github.com/google/go-github/v24/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type Cli struct {
	cli *github.Client
}

func New(token string) *Cli {
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
	for i := 1; ; i++ {
		users, _, err := c.cli.Users.
			ListFollowers(ctx, user, &github.ListOptions{
				Page:    i,
				PerPage: 100,
			})
		if err != nil {
			if len(all) == 0 {
				return nil, fmt.Errorf("list follower error: %s", err)
			}
			break
		}
		if len(users) == 0 {
			break
		}
		all = append(all, users...)
	}
	return all, nil
}

func (c *Cli) following(ctx context.Context, user string) ([]*github.User, error) {
	all := []*github.User{}
	for i := 1; ; i++ {
		users, _, err := c.cli.Users.
			ListFollowing(ctx, user, &github.ListOptions{
				Page:    i,
				PerPage: 100,
			})
		if err != nil {
			if len(all) == 0 {
				return nil, fmt.Errorf("list following error: %s", err)
			}
			break
		}
		if len(users) == 0 {
			break
		}
		all = append(all, users...)
	}
	return all, nil
}

func (c *Cli) RelatedUsers(ctx context.Context, user string) []string {
	users := []*github.User{}

	if x, err := c.followers(ctx, user); err != nil {
		logrus.Errorf("list [%s] follower error: %s", user, err)
	} else {
		logrus.Infof("%s has %d followers", user, len(x))
		users = append(users, x...)
	}

	if x, err := c.following(ctx, user); err != nil {
		logrus.Errorf("list [%s] following error: %s", user, err)
	} else {
		logrus.Infof("%s following %d", user, len(x))
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

	logrus.Infof("%s related users %d", user, len(related))

	return related
}

func (c *Cli) Follow(user string) {
	logrus.Infof("follow %s", user)
	// c.cli.Users.Follow(context.Background(), user)
}
