package main

import (
	"encoding/base64"
	"errors"

	"github.com/charmbracelet/log"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/exp/rand"
	"gopkg.in/yaml.v3"
)

type ReviewersInfo struct {
	ReviewThreshold int
	Usernames []string
}

func (a app) get_reviewers_usernames(merge_request *gitlab.MergeRequest) (ReviewersInfo, error) {
	var ri ReviewersInfo
	file, _, err := a.client.RepositoryFiles.GetFile(merge_request.ProjectID, "reviewers.yaml", &gitlab.GetFileOptions{
		Ref: &a.conf.ReviewerFileRef,
	})
	if err != nil {
		return ri, errors.Join(errors.New("Failed to get reviewers file"), err)
	}

	file_contents, err := base64.StdEncoding.DecodeString(file.Content)
	if err != nil {
		return ri, errors.Join(errors.New("Could not decode contents of file"), err)
	}

	yaml.Unmarshal(file_contents, &ri)
	return ri, nil
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

func shuffle_reviewers(reviewers *ReviewersInfo) error {
	if reviewers == nil {
		return errors.New("Reviewers list is nil")
	}
	for i := range reviewers.Usernames {
		j := rand.Intn(i + 1)
		reviewers.Usernames[i], reviewers.Usernames[j] = reviewers.Usernames[j], reviewers.Usernames[i]
	}
	return nil
}
