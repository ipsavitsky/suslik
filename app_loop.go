package main

import (
	"fmt"
	"os"
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
		project, _, err := a.client.Projects.GetProject(mergeRequest.ProjectID, &gitlab.GetProjectOptions{})
		if err != nil {
			log.Error("Could not get project", "err", err, "mr", mergeRequest.IID)
		}

		logger := log.Default().WithPrefix(fmt.Sprintf("[%s!%d] %s", project.NameWithNamespace, mergeRequest.ID, mergeRequest.Title))

		reviewersInfo, err := a.getReviewersInfo(mergeRequest)
		if err != nil {
			logger.Error("Failed to get reviewers info", "err", err)
			continue
		}
		err = shuffleReviewers(&reviewersInfo)
		if err != nil {
			logger.Error("Failed to shuffle reviewers", "err", err)
			continue
		}

		reviewerUsers, warnings := a.getUsers(reviewersInfo.Usernames)
		logger.Debugf("Got %d reviewer users", len(reviewerUsers))

		var reviewersIDs []int
		for _, existingReviewer := range mergeRequest.Reviewers {
			if existingReviewer.ID == a.getCurrentUser().ID {
				logger.Debug("Skipping bot user", "id", existingReviewer.ID, "username", existingReviewer.Username)
				continue
			}
			if existingReviewer.ID == mergeRequest.Author.ID {
				logger.Debug("Skipping MR author", "id", existingReviewer.ID, "username", existingReviewer.Username)
			}
			reviewersIDs = append(reviewersIDs, existingReviewer.ID)
		}

		currentAssignedReviewers := len(reviewersIDs)
		amountOfUsersToAssign := reviewersInfo.ReviewThreshold - currentAssignedReviewers
		logger.Debugf("There are %d reviewers already assigned", currentAssignedReviewers)
		logger.Debugf("Assigning %d users", amountOfUsersToAssign)

		var reviewersFormattedUsernames []string

		i := 0
		added_reviewers := 0
		for added_reviewers < min(amountOfUsersToAssign, len(reviewerUsers)) {
			if reviewerUsers[i].ID == mergeRequest.Author.ID {
				logger.Debug("Skipping MR author in reviewer list", "id", reviewerUsers[i].ID, "username", reviewerUsers[i].Username)
				i++
				continue
			}

			for _, reviewer := range reviewersIDs {
				if reviewerUsers[i].ID == reviewer {
					logger.Debug("Reviewer already assigned", "id", reviewer, "username", reviewerUsers[i].Username)
					i++
					continue
				}
			}

			reviewersIDs = append(reviewersIDs, reviewerUsers[i].ID)
			reviewersFormattedUsernames = append(reviewersFormattedUsernames, "`@"+reviewerUsers[i].Username+"`")
			added_reviewers++
			i++
		}

		assignmentBody := fmt.Sprintf("Unassigning myself, assigning random reviewers (%s)<br />", strings.Join(reviewersFormattedUsernames, ", "))

		var job_id string
		var job_url string
		if a.conf.Mode == CI {
			job_id = os.Getenv("CI_JOB_ID")
			job_url = os.Getenv("CI_JOB_URL")
		}

		if len(warnings) != 0 {
			assignmentBody += fmt.Sprintf("⚠ Warnings:<br />")
			for _, warning := range warnings {
				assignmentBody += fmt.Sprintf("- %s<br />", warning)
			}
		}

		if a.conf.Mode == CI {
			assignmentBody += fmt.Sprintf("<br />*Note generated by job [%s](%s)*", job_id, job_url)
		}

		_, _, err = a.client.Notes.CreateMergeRequestNote(mergeRequest.ProjectID, mergeRequest.IID, &gitlab.CreateMergeRequestNoteOptions{
			Body: &assignmentBody,
		})
		if err != nil {
			logger.Error("Failed to create a merge request note", "err", err)
			continue
		}

		_, _, err = a.client.MergeRequests.UpdateMergeRequest(mergeRequest.ProjectID, mergeRequest.IID, &gitlab.UpdateMergeRequestOptions{
			ReviewerIDs: gitlab.Ptr(reviewersIDs),
		})
		if err != nil {
			logger.Error("Failed to assign reviewers", "err", err)
			continue
		}
	}
}
