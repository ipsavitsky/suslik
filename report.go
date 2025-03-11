package main

import (
	"fmt"

	"github.com/charmbracelet/log"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func (a app) ReportErrorOnMergeRequest(mr *gitlab.BasicMergeRequest, error error) {
	_, _, err := a.client.Notes.CreateMergeRequestNote(mr.ProjectID, mr.IID, &gitlab.CreateMergeRequestNoteOptions{
		Body: gitlab.Ptr(fmt.Sprintf("🚨 %v", error)),
	})
	if err != nil {
		log.Errorf("Failed to create a merge request note: %v", err)
	}
}
