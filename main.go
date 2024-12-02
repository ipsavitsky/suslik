package main

import (
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
		log.Debugf("File contents: %s", file)

		log.Debugf("project id: %d", merge_request.ProjectID)
		log.Debugf("merge_request iid: %d", merge_request.IID)
		_, req, err := app.client.Discussions.CreateMergeRequestDiscussion(merge_request.ProjectID, merge_request.IID, &gitlab.CreateMergeRequestDiscussionOptions{
			Body: gitlab.Ptr("Unassigning myself, assigning random reviewers"),
		})

		if err != nil {
			log.Errorf("Failed to create a merge request discussion: %v; %v", err, req)
		}
	}
}
