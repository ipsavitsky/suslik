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
		log.Fatalf("Failed to create client: %v", err)
	}
	return git
}

func (a *app) get_current_user() *gitlab.User {
	user, _, err := a.client.Users.CurrentUser()
	if err != nil {
		log.Fatalf("Failed to get current user: %v", err)
	}
	log.Debugf("Name: %s", user.Name)
	return user
}

func main() {
	log.SetLevel(log.DebugLevel)

	app := app{
		client: get_gitlab_client("<token>"),
	}

	list_merge_request_options := &gitlab.ListMergeRequestsOptions{
		ReviewerID: gitlab.ReviewerID(app.get_current_user().ID),
	}

	merge_requests, _, err := app.client.MergeRequests.ListMergeRequests(list_merge_request_options)
	if err != nil {
		log.Fatalf("Failed to get curent merge requests: %v", err)
	}

	log.Debugf("Found merge requsts assigned to: %d", len(merge_requests))

	for _, merge_request := range merge_requests {
		project_id := merge_request.ProjectID
		file, _, err := app.client.RepositoryFiles.GetFile(project_id, "REVIEWERS", &gitlab.GetFileOptions{})
		if err != nil {
			log.Error(err)
			continue
		}
		log.Debugf("File contents: %s", file)

		text := "Unassigning myself, assigning random reviewers"
		options := &gitlab.CreateMergeRequestDiscussionOptions{
			Body: &text,
		}

		log.Debugf("project id: %d", merge_request.ProjectID)
		log.Debugf("merge_request iid: %d", merge_request.IID)
		_, req, err := app.client.Discussions.CreateMergeRequestDiscussion(merge_request.ProjectID, merge_request.IID, options)

		if err != nil {
			log.Errorf("Failed to create a merge request discussion: %v; %s", err, req)
		}
	}
}
