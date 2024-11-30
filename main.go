package main

import (
	"log"

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
	return user
}

func main() {
	app := app{
		client: get_gitlab_client("<token>"),
	}

	list_merge_request_options := &gitlab.ListMergeRequestsOptions{
		ApproverIDs: gitlab.ApproverIDs([]int{app.get_current_user().ID}),
	}

	merge_requests, _, err := app.client.MergeRequests.ListMergeRequests(list_merge_request_options)
	if err != nil {
		log.Fatalf("Failed to get curent merge requests: %v", err)
	}

	log.Println(merge_requests)
}
