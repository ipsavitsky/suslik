package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"testing"

	"gitlab.com/gitlab-org/api/client-go"
)

var TestGitlabClient *gitlab.Client

func init() {
	token, found := os.LookupEnv("SUSLIK_GITLAB_TOKEN")

	if !found {
		panic("no token found")
	}

	client, err := gitlab.NewClient(
		token,
		gitlab.WithBaseURL("http://localhost:9999/api/v4"),
	)
	if err != nil {
		panic("failed to create test client: " + err.Error())
	}

	TestGitlabClient = client
}

func CreateProject(t *testing.T) *gitlab.Project {
	t.Helper()

	options := &gitlab.CreateProjectOptions{
		Name:                 gitlab.Ptr("Test project"),
		Description:          gitlab.Ptr("Suslik integration test"),
		Visibility:           gitlab.Ptr(gitlab.PublicVisibility),
		InitializeWithReadme: gitlab.Ptr(true),
	}

	project, _, err := TestGitlabClient.Projects.CreateProject(options)
	if err != nil {
		t.Fatalf("could not create test project: %v", err)
	}

	t.Cleanup(func() {
		if _, err := TestGitlabClient.Projects.DeleteProject(project.ID, nil); err != nil {
			t.Fatalf("could not cleanup test project: %v", err)
		}
	})

	return project
}

func CreateUserWithToken(t *testing.T, username string) (*gitlab.User, *gitlab.PersonalAccessToken) {
	t.Helper()

	options := &gitlab.CreateUserOptions{
		Name:     &username,
		Username: &username,
		Email:    gitlab.Ptr(fmt.Sprintf("%s@example.invalid", username)),
		Password: gitlab.Ptr("insecure1111"),
	}

	user, _, err := TestGitlabClient.Users.CreateUser(options)
	if err != nil {
		t.Fatalf("could not create test user: %v", err)
	}

	t.Cleanup(func() {
		if _, err := TestGitlabClient.Users.DeleteUser(user.ID, nil); err != nil {
			t.Fatalf("could not cleanup user: %v", err)
		}
	})

	token_options := &gitlab.CreatePersonalAccessTokenOptions{
		Name: gitlab.Ptr("test_token"),
		Scopes: &[]string{"api", "read_user", "write_repository"},
	}

	token, _, err := TestGitlabClient.Users.CreatePersonalAccessToken(user.ID, token_options)
	if err != nil {
		t.Fatalf("could not create token for user: %v", err)
	}

	return user, token
}

func CreateBranchInProject(t *testing.T, project *gitlab.Project, branchName string) *gitlab.Branch {
	t.Helper()

	options := &gitlab.CreateBranchOptions{
		Branch: &branchName,
		Ref:    gitlab.Ptr("main"),
	}

	branch, _, err := TestGitlabClient.Branches.CreateBranch(project.ID, options)
	if err != nil {
		t.Fatalf("error creating branch: %v", err)
	}

	return branch
}

func AddFileToProject(t *testing.T, project *gitlab.Project, fileName string, fileContents string, branch string) *gitlab.FileInfo {
	t.Helper()

	options := &gitlab.CreateFileOptions{
		Branch:        &branch,
		Encoding:      gitlab.Ptr("base64"),
		Content:       gitlab.Ptr(base64.StdEncoding.EncodeToString([]byte(fileContents))),
		CommitMessage: gitlab.Ptr(fmt.Sprintf("Add %s", fileName)),
	}
	file, _, err := TestGitlabClient.RepositoryFiles.CreateFile(project.ID, fileName, options)
	if err != nil {
		t.Fatalf("unable to create a repository file: %v", err)
	}

	return file
}

func CreateMergeRequestWithReviewer(t *testing.T, project *gitlab.Project, reviewer *gitlab.User, sourceBranch string, targetBranch string) *gitlab.MergeRequest {
	t.Helper()

	options := &gitlab.CreateMergeRequestOptions{
		Title:        gitlab.Ptr("Test merge request"),
		Description:  gitlab.Ptr("Test merge request"),
		SourceBranch: &sourceBranch,
		TargetBranch: &targetBranch,
		ReviewerIDs:  &[]int{reviewer.ID},
	}

	mergeRequest, _, err := TestGitlabClient.MergeRequests.CreateMergeRequest(project.ID, options)
	if err != nil {
		t.Fatalf("could not create merge request: %v", err)
	}

	return mergeRequest
}

func GetReviewersOnMergeRequest(t *testing.T, project *gitlab.Project, mr *gitlab.MergeRequest) []*gitlab.MergeRequestReviewer {
	t.Helper()

	reviewers, _, err := TestGitlabClient.MergeRequests.GetMergeRequestReviewers(project.ID, mr.IID)
	if err != nil {
		t.Fatalf("could not get reviewers: %v", err)
	}

	return reviewers
}

func TestBasicFunctionality(t *testing.T) {
	project := CreateProject(t)
	user, _ := CreateUserWithToken(t, "test_user")

	reviewersFile := fmt.Sprintf(`reviewThreshold: 1
usernames:
  - %s`, user.Username)

	AddFileToProject(t, project, "reviewers.yaml", reviewersFile, project.DefaultBranch)
	branch := CreateBranchInProject(t, project, "test-branch")
	AddFileToProject(t, project, "test", "this is a test file", branch.Name)
	suslik_account, suslik_token := CreateUserWithToken(t, "suslik")
	mr := CreateMergeRequestWithReviewer(t, project, suslik_account, branch.Name, project.DefaultBranch)

	conf := Config{
		Token:           suslik_token.Token,
		BaseURL:         "http://localhost:9999",
		ReviewerFileRef: "main",
		PollDelay:       0,
	}

	app := app{
		client: getGitlabClient(conf.Token, conf.BaseURL),
		conf:   conf,
	}

	app.run()

	reviewersAfterRun := GetReviewersOnMergeRequest(t, project, mr)
	if reviewersAfterRun[0].User.ID != user.ID {
		t.Fatalf("Assigned wrong user: %d(%s) != %d(%s)", reviewersAfterRun[0].User.ID, reviewersAfterRun[0].User.Username, user.ID, user.Username)
	}
}
