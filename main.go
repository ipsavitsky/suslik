package main

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/charmbracelet/log"

	gitlab "github.com/xanzy/go-gitlab"
)

type app struct {
	client *gitlab.Client
}

func get_gitlab_client(token string) *gitlab.Client {
	git, err := gitlab.NewClient(token)
	if err != nil {
		log.Errorf("Failed to create client: %v", err)
	}
	return git
}

func (a *app) get_current_user() *gitlab.User {
	user, _, err := a.client.Users.CurrentUser()
	if err != nil {
		log.Errorf("Failed to get current user: %v", err)
	}
	log.Debugf("Name: %s", user.Name)
	return user
}

func main() {
	log.SetLevel(log.DebugLevel)

	conf := parseConfig("conf.toml")

	app := app{
		client: get_gitlab_client(conf.Token),
	}

	merge_requests, _, err := app.client.MergeRequests.ListMergeRequests(&gitlab.ListMergeRequestsOptions{
		ReviewerID: gitlab.ReviewerID(app.get_current_user().ID),
	})

	if err != nil {
		log.Errorf("Failed to get curent merge requests: %v", err)
	}

	log.Debugf("Found merge requsts assigned to: %d", len(merge_requests))

	for _, merge_request := range merge_requests {
		project_id := merge_request.ProjectID
		file, _, err := app.client.RepositoryFiles.GetFile(project_id, "REVIEWERS", &gitlab.GetFileOptions{
			Ref: &conf.ReviewerFileRef,
		})
		if err != nil {
			log.Error(err)
			continue
		}

		file_contents, err := base64.StdEncoding.DecodeString(file.Content)
		if err != nil {
			log.Fatalf("Could not decode contents of file: %v", file)
		}

		reviewers_nicks := strings.Split(string(file_contents), "\n")
		log.Debugf("Got reviewers: %v", reviewers_nicks)

		shuffle_reviewers(&reviewers_nicks)

		reviewer_users := app.get_users(reviewers_nicks)
		log.Debugf("Got %d reviewer users", len(reviewer_users))


		var sb strings.Builder
		sb.WriteString("/assign_reviewer")
		for _, reviewer := range reviewer_users {
			sb.WriteString(fmt.Sprintf(" @%s", reviewer.Username))
		}
		log.Debugf("Generated string is: %s", sb.String())

		_, req, err := app.client.Discussions.CreateMergeRequestDiscussion(merge_request.ProjectID, merge_request.IID, &gitlab.CreateMergeRequestDiscussionOptions{
			Body: gitlab.Ptr(fmt.Sprintf("Unassigning myself, assigning random reviewers\n%s",sb.String())),
		})
		if err != nil {
			log.Errorf("Failed to create a merge request discussion: %v; %v", err, req)
		}
	}
}
