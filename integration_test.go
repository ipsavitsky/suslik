package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"testing"
	"time"

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
		panic("failed to create test client: " + err.Error()) // lintignore: R009 // TODO: Resolve this tfproviderlint issue
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

func CreateUser(t *testing.T) *gitlab.User {
	t.Helper()

	options := &gitlab.CreateUserOptions{
		Name:     gitlab.Ptr("gaba"),
		Username: gitlab.Ptr("goo"),
		Email:    gitlab.Ptr("test@example.invalid"),
		Password: gitlab.Ptr("hehehahu"),
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

	return user
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

func TestBasicFunctionality(t *testing.T) {
	project := CreateProject(t)
	user := CreateUser(t)

	reviewersFile := fmt.Sprintf(`reviewThreshold: 1
usernames:
  - %s`, user.Username)

	AddFileToProject(t, project, "reviewers.yaml", reviewersFile, project.DefaultBranch)
	branch := CreateBranchInProject(t, project, "test-branch")
	AddFileToProject(t, project, "test", "this is a test file", branch.Name)

	CreateMergeRequestWithReviewer(t, project, user, branch.Name, project.DefaultBranch)

	time.Sleep(10 * time.Second)

	token, found := os.LookupEnv("SUSLIK_GITLAB_TOKEN")

	if !found {
		t.Fail()
	}

	conf := Config{
		Token:           token,
		BaseURL:         "http://localhost:9999",
		ReviewerFileRef: "main",
		PollDelay:       0,
	}

	app := app{
		client: getGitlabClient(conf.Token, conf.BaseURL),
		conf:   conf,
	}

	app.run()
}
