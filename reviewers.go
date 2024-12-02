package main

import (
	"github.com/charmbracelet/log"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/exp/rand"
)

func (a app) get_users(reviewers []string) []*gitlab.User {
	var users []*gitlab.User
	for _, reviewer := range reviewers {
		queried_users, _, err := a.client.Users.ListUsers(&gitlab.ListUsersOptions{
			Username: &reviewer,
		})

		if err != nil {
			log.Warnf("Error querying for user %s: %v", reviewer, err)
			continue
		}

		if len(queried_users) == 0 {
			log.Warnf("Found no users for the username %s, skipping", reviewer)
			continue
		}

		if len(queried_users) != 1 {
			log.Warnf("Found more then 1 match on %s (%d); assuming first match", reviewer, len(queried_users))
		}

		if queried_users[0].Username != reviewer {
			log.Warnf("First match is not an exact match (%s != %s), skipping", queried_users[0].Username, reviewer)
			continue
		}

		users = append(users, queried_users[0])
	}

	return users
}

func shuffle_reviewers(reviewers *[]string) {
	for i := range *reviewers {
		j := rand.Intn(i + 1)
		(*reviewers)[i], (*reviewers)[j] = (*reviewers)[j], (*reviewers)[i]
	}
}
