package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/log"

	gitlab "github.com/xanzy/go-gitlab"
)

func (a app) loop() {
	for {
		a.run()
		log.Debug("Sleeping for 10 seconds...")
		time.Sleep(10 * time.Second)
	}
}

func (a app) run() {
	var err error

	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("pkg: %v", r)
			}
			log.Errorf("Uncaught panic occured: %v:", err)
		}
	}()

	merge_requests, _, err := a.client.MergeRequests.ListMergeRequests(&gitlab.ListMergeRequestsOptions{
		ReviewerID: gitlab.ReviewerID(a.get_current_user().ID),
	})

	if err != nil {
		log.Errorf("Failed to get assigned merge requests: %v", err)
		return
	}

	log.Debugf("Found %d assigned merge requests", len(merge_requests))

	for _, merge_request := range merge_requests {
		reviewers_nicks, err := a.get_reviewers_info(merge_request)
		if err != nil {
			log.Errorf("Failed to get reviewers info: %v", err)
			continue
		}
		err = shuffle_reviewers(&reviewers_nicks)
		if err != nil {
			log.Errorf("Failed to shuffle reviewers: %v", err)
			continue
		}

		reviewer_users := a.get_users(reviewers_nicks.Usernames)
		log.Debugf("Got %d reviewer users", len(reviewer_users))

		current_assigned_reviewers := len(merge_request.Reviewers)
		amount_of_users_to_assign := reviewers_nicks.ReviewThreshold - current_assigned_reviewers
		log.Debugf("reviewers_nicks.ReviewThreshold: %d", reviewers_nicks.ReviewThreshold)
		log.Debugf("There are %d reviewers already assigned", current_assigned_reviewers)
		log.Debugf("Assigning %d users", amount_of_users_to_assign)

		var sb strings.Builder
		sb.WriteString("/assign_reviewer")
		for i := 0; i < min(len(reviewer_users), amount_of_users_to_assign); i++ {
			sb.WriteString(fmt.Sprintf(" @%s", reviewer_users[i].Username))
		}
		sb.WriteString(fmt.Sprintf("\n/unassign_reviewer @%s", a.get_current_user().Username))
		log.Debugf("Generated string is: %s", sb.String())

		_, req, err := a.client.Discussions.CreateMergeRequestDiscussion(merge_request.ProjectID, merge_request.IID, &gitlab.CreateMergeRequestDiscussionOptions{
			Body: gitlab.Ptr(fmt.Sprintf("Unassigning myself, assigning random reviewers\n%s", sb.String())),
		})
		if err != nil {
			log.Errorf("Failed to create a merge request discussion: %v; %v", err, req)
		}
	}
}
