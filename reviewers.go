package main

import (
	"encoding/base64"
	"errors"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/exp/rand"
)

type ReviewersInfo []string

func (a app) get_reviewers_usernames(merge_request *gitlab.MergeRequest) (ReviewersInfo, error) {
	file, _, err := a.client.RepositoryFiles.GetFile(merge_request.ProjectID, "REVIEWERS", &gitlab.GetFileOptions{
		Ref: &a.conf.ReviewerFileRef,
	})
	if err != nil {
		return nil, errors.Join(errors.New("Failed to get reviewers file"), err)
	}

	file_contents, err := base64.StdEncoding.DecodeString(file.Content)
	if err != nil {
		return nil, errors.Join(errors.New("Could not decode contents of file"), err)
	}

	reviewers_nicks := strings.Split(string(file_contents), "\n")
	log.Debugf("Got reviewers: %v", reviewers_nicks)
	return reviewers_nicks, nil
}

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

func shuffle_reviewers(reviewers *[]string) error {
	if reviewers == nil {
		return errors.New("Reviewers list is nil")
	}
	for i := range *reviewers {
		j := rand.Intn(i + 1)
		(*reviewers)[i], (*reviewers)[j] = (*reviewers)[j], (*reviewers)[i]
	}
	return nil
}
