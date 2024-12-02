package main

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/charmbracelet/log"

	gitlab "github.com/xanzy/go-gitlab"
)

func (a app) run() {
	merge_requests, _, err := a.client.MergeRequests.ListMergeRequests(&gitlab.ListMergeRequestsOptions{
		ReviewerID: gitlab.ReviewerID(a.get_current_user().ID),
	})

	if err != nil {
		log.Errorf("Failed to get curent merge requests: %v", err)
	}

	log.Debugf("Found merge requsts assigned to: %d", len(merge_requests))

	for _, merge_request := range merge_requests {
		file, _, err := a.client.RepositoryFiles.GetFile(merge_request.ProjectID, "REVIEWERS", &gitlab.GetFileOptions{
			Ref: &a.conf.ReviewerFileRef,
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

		reviewer_users := a.get_users(reviewers_nicks)
		log.Debugf("Got %d reviewer users", len(reviewer_users))

		var sb strings.Builder
		sb.WriteString("/assign_reviewer")
		for _, reviewer := range reviewer_users {
			sb.WriteString(fmt.Sprintf(" @%s", reviewer.Username))
		}
		log.Debugf("Generated string is: %s", sb.String())

		_, req, err := a.client.Discussions.CreateMergeRequestDiscussion(merge_request.ProjectID, merge_request.IID, &gitlab.CreateMergeRequestDiscussionOptions{
			Body: gitlab.Ptr(fmt.Sprintf("Unassigning myself, assigning random reviewers\n%s", sb.String())),
		})
		if err != nil {
			log.Errorf("Failed to create a merge request discussion: %v; %v", err, req)
		}
	}
}
