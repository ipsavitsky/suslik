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

	mergeRequests, _, err := a.client.MergeRequests.ListMergeRequests(&gitlab.ListMergeRequestsOptions{
		ReviewerID: gitlab.ReviewerID(a.getCurrentUser().ID),
	})

	if err != nil {
		log.Errorf("Failed to get assigned merge requests: %v", err)
		return
	}

	log.Debugf("Found %d assigned merge requests", len(mergeRequests))

	for _, mergeRequest := range mergeRequests {
		reviewersInfo, err := a.getReviewersInfo(mergeRequest)
		if err != nil {
			log.Errorf("Failed to get reviewers info: %v", err)
			continue
		}
		err = shuffleReviewers(&reviewersInfo)
		if err != nil {
			log.Errorf("Failed to shuffle reviewers: %v", err)
			continue
		}

		reviewerUsers := a.getUsers(reviewersInfo.Usernames)
		log.Debugf("Got %d reviewer users", len(reviewerUsers))

		currentAssignedReviewers := len(mergeRequest.Reviewers)
		amountOfUsersToAssign := reviewersInfo.ReviewThreshold - currentAssignedReviewers
		log.Debugf("There are %d reviewers already assigned", currentAssignedReviewers)
		log.Debugf("Assigning %d users", amountOfUsersToAssign)

		var sb strings.Builder
		sb.WriteString("/assign_reviewer")
		for i := 0; i < min(len(reviewerUsers), amountOfUsersToAssign); i++ {
			sb.WriteString(fmt.Sprintf(" @%s", reviewerUsers[i].Username))
		}
		sb.WriteString(fmt.Sprintf("\n/unassign_reviewer @%s", a.getCurrentUser().Username))
		log.Debugf("Generated string is: %s", sb.String())

		_, req, err := a.client.Discussions.CreateMergeRequestDiscussion(mergeRequest.ProjectID, mergeRequest.IID, &gitlab.CreateMergeRequestDiscussionOptions{
			Body: gitlab.Ptr(fmt.Sprintf("Unassigning myself, assigning random reviewers\n%s", sb.String())),
		})
		if err != nil {
			log.Errorf("Failed to create a merge request discussion: %v; %v", err, req)
		}
	}
}
