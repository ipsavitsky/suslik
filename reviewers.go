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
	ReviewThreshold int `yaml:"reviewThreshold"`
	Usernames []string  `yaml:"usernames"`
}

func (a app) getReviewersInfo(mergeRequest *gitlab.MergeRequest) (ReviewersInfo, error) {
	var ri ReviewersInfo
	file, _, err := a.client.RepositoryFiles.GetFile(mergeRequest.ProjectID, "reviewers.yaml", &gitlab.GetFileOptions{
		Ref: &a.conf.ReviewerFileRef,
	})
	if err != nil {
		return ri, errors.Join(errors.New("failed to get reviewers file"), err)
	}

	fileContents, err := base64.StdEncoding.DecodeString(file.Content)
	if err != nil {
		return ri, errors.Join(errors.New("could not decode contents of file"), err)
	}

	err = yaml.Unmarshal(fileContents, &ri)
	if err != nil {
		return ri, errors.Join(errors.New("failed unmarshalling a file"), err)
	}

	return ri, nil
}

func (a app) getUsers(reviewers []string) []*gitlab.User {
	var users []*gitlab.User
	for _, reviewer := range reviewers {
		queriedUsers, _, err := a.client.Users.ListUsers(&gitlab.ListUsersOptions{
			Username: &reviewer,
		})

		if err != nil {
			log.Warnf("Error querying for user %s: %v", reviewer, err)
			continue
		}

		if len(queriedUsers) == 0 {
			log.Warnf("Found no users for the username %s, skipping", reviewer)
			continue
		}

		if len(queriedUsers) != 1 {
			log.Warnf("Found more then 1 match on %s (%d); assuming first match", reviewer, len(queriedUsers))
		}

		if queriedUsers[0].Username != reviewer {
			log.Warnf("First match is not an exact match (%s != %s), skipping", queriedUsers[0].Username, reviewer)
			continue
		}

		users = append(users, queriedUsers[0])
	}

	return users
}

func shuffleReviewers(reviewers *ReviewersInfo) error {
	if reviewers == nil {
		return errors.New("reviewers list is nil")
	}
	for i := range reviewers.Usernames {
		j := rand.Intn(i + 1)
		reviewers.Usernames[i], reviewers.Usernames[j] = reviewers.Usernames[j], reviewers.Usernames[i]
	}
	return nil
}
