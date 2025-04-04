package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/log"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func (a app) loop() {
	runTicker := time.NewTicker(a.conf.PollDelay)
	for {
		a.run()
		log.Debugf("Sleeping...")
		<-runTicker.C
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
		Scope:      gitlab.Ptr("all"),
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
			a.ReportErrorOnMergeRequest(mergeRequest, err)
			continue
		}
		err = shuffleReviewers(&reviewersInfo)
		if err != nil {
			log.Errorf("Failed to shuffle reviewers: %v", err)
			continue
		}

		reviewerUsers, warnings := a.getUsers(reviewersInfo.Usernames)
		log.Debugf("Got %d reviewer users", len(reviewerUsers))

		var reviewersIDs []int
		for _, existingReviewer := range mergeRequest.Reviewers {
			if existingReviewer.ID == a.getCurrentUser().ID {
				log.Debug("Skipping bot user", "id", existingReviewer.ID, "username", existingReviewer.Username)
				continue
			}
			if existingReviewer.ID == mergeRequest.Author.ID {
				log.Debug("Skipping MR author", "id", existingReviewer.ID, "username", existingReviewer.Username)
			}
			reviewersIDs = append(reviewersIDs, existingReviewer.ID)
		}

		currentAssignedReviewers := len(reviewersIDs)
		amountOfUsersToAssign := reviewersInfo.ReviewThreshold - currentAssignedReviewers
		log.Debugf("There are %d reviewers already assigned", currentAssignedReviewers)
		log.Debugf("Assigning %d users", amountOfUsersToAssign)

		var reviewersFormattedUsernames []string

		i := 0
		added_reviewers := 0
		for added_reviewers < min(amountOfUsersToAssign, len(reviewerUsers)) {
			if reviewerUsers[i].ID == mergeRequest.Author.ID {
				log.Debug("Skipping MR author in reviewer list", "id", reviewerUsers[i].ID, "username", reviewerUsers[i].ID)
				i++
				continue
			}
			reviewersIDs = append(reviewersIDs, reviewerUsers[i].ID)
			reviewersFormattedUsernames = append(reviewersFormattedUsernames, "`@"+reviewerUsers[i].Username+"`")
			added_reviewers++
			i++
		}

		assignmentBody := fmt.Sprintf("Unassigning myself, assigning random reviewers (%s)<br>", strings.Join(reviewersFormattedUsernames, ", "))

		if len(warnings) != 0 {
			assignmentBody += fmt.Sprintf("⚠ Warnings:\n")
			for _, warning := range warnings {
				assignmentBody += fmt.Sprintf("- %s\n", warning)
			}
		}

		_, req, err := a.client.Notes.CreateMergeRequestNote(mergeRequest.ProjectID, mergeRequest.IID, &gitlab.CreateMergeRequestNoteOptions{
			Body: &assignmentBody,
		})
		if err != nil {
			log.Errorf("Failed to create a merge request note: %v; %v", err, req)
			continue
		}

		_, req, err = a.client.MergeRequests.UpdateMergeRequest(mergeRequest.ProjectID, mergeRequest.IID, &gitlab.UpdateMergeRequestOptions{
			ReviewerIDs: gitlab.Ptr(reviewersIDs),
		})
		if err != nil {
			log.Errorf("Failed to assign reviewers: %v; %v", err, req)
			continue
		}
	}
}
