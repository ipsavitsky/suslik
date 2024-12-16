package main

import (
	"fmt"
	"testing"

	tu "github.com/ipsavitsky/suslik/testutil"
)

func TestBasicFunctionality(t *testing.T) {
	project := tu.CreateProject(t)
	user, _ := tu.CreateUserWithToken(t, "test_user")
	tu.AddUserToProject(t, user, project)

	reviewersFile := fmt.Sprintf(`reviewThreshold: 1
usernames:
  - %s`, user.Username)

	tu.AddFileToProject(t, project, "reviewers.yaml", reviewersFile, project.DefaultBranch)
	branch := tu.CreateBranchInProject(t, project, "test-branch")
	tu.AddFileToProject(t, project, "test", "this is a test file", branch.Name)
	suslik_account, suslik_token := tu.CreateUserWithToken(t, "suslik")
	tu.AddUserToProject(t, suslik_account, project)
	mr := tu.CreateMergeRequestWithReviewer(t, project, suslik_account, branch.Name, project.DefaultBranch)

	conf := Config{
		Token:           suslik_token.Token,
		BaseURL:         "http://localhost:9999/api/v4",
		ReviewerFileRef: "main",
		PollDelay:       0,
	}

	app := app{
		client: getGitlabClient(conf.Token, conf.BaseURL),
		conf:   conf,
	}

	app.run()

	reviewersAfterRun := tu.GetReviewersOnMergeRequest(t, project, mr)
	for _, reviewer := range reviewersAfterRun {
		t.Logf("Reviewer username: %s", reviewer.User.Username)
	}
	if reviewersAfterRun[0].User.ID != user.ID {
		t.Fatalf("Assigned wrong user: %d(%s) != %d(%s)", reviewersAfterRun[0].User.ID, reviewersAfterRun[0].User.Username, user.ID, user.Username)
	}
}
