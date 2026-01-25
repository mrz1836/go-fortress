package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// EventType represents GitHub event types.
type EventType string

const (
	EventPush              EventType = "push"
	EventPullRequest       EventType = "pull_request"
	EventPullRequestTarget EventType = "pull_request_target"
	EventWorkflowDispatch  EventType = "workflow_dispatch"
	EventSchedule          EventType = "schedule"
)

// PushEvent represents a GitHub push event payload.
type PushEvent struct {
	Ref        string     `json:"ref"`
	Before     string     `json:"before"`
	After      string     `json:"after"`
	Repository Repository `json:"repository"`
	Pusher     Actor      `json:"pusher"`
	Sender     Actor      `json:"sender"`
	Commits    []Commit   `json:"commits,omitempty"`
}

// PullRequestEvent represents a GitHub pull request event payload.
type PullRequestEvent struct {
	Action      string      `json:"action"`
	Number      int         `json:"number"`
	PullRequest PullRequest `json:"pull_request"`
	Repository  Repository  `json:"repository"`
	Sender      Actor       `json:"sender"`
}

// Repository represents a GitHub repository.
type Repository struct {
	ID       int64  `json:"id"`
	NodeID   string `json:"node_id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
	Owner    Actor  `json:"owner"`
	HTMLURL  string `json:"html_url"`
	CloneURL string `json:"clone_url"`
}

// Actor represents a GitHub user or organization.
type Actor struct {
	Login     string `json:"login"`
	ID        int64  `json:"id"`
	NodeID    string `json:"node_id"`
	AvatarURL string `json:"avatar_url"`
	Type      string `json:"type"`
}

// Commit represents a Git commit.
type Commit struct {
	ID        string `json:"id"`
	TreeID    string `json:"tree_id"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	Author    Author `json:"author"`
	Committer Author `json:"committer"`
}

// Author represents a commit author.
type Author struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// PullRequest represents a GitHub pull request.
type PullRequest struct {
	Number  int    `json:"number"`
	State   string `json:"state"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	Head    PRRef  `json:"head"`
	Base    PRRef  `json:"base"`
	HTMLURL string `json:"html_url"`
}

// PRRef represents a pull request branch reference.
type PRRef struct {
	Label string     `json:"label"`
	Ref   string     `json:"ref"`
	SHA   string     `json:"sha"`
	Repo  Repository `json:"repo"`
}

// DefaultPushEvent creates a standard push event for testing.
func DefaultPushEvent() *PushEvent {
	return &PushEvent{
		Ref:    "refs/heads/main",
		Before: "0000000000000000000000000000000000000000",
		After:  "abc123def456789012345678901234567890abcd",
		Repository: Repository{
			ID:       1,
			Name:     "test-repo",
			FullName: "test-owner/test-repo",
			Owner: Actor{
				Login: "test-owner",
				Type:  "User",
			},
		},
		Pusher: Actor{
			Login: "test-user",
		},
		Sender: Actor{
			Login: "test-user",
			Type:  "User",
		},
	}
}

// DefaultPullRequestEvent creates a standard PR event for testing.
func DefaultPullRequestEvent() *PullRequestEvent {
	return &PullRequestEvent{
		Action: "opened",
		Number: 1,
		PullRequest: PullRequest{
			Number: 1,
			State:  "open",
			Title:  "Test PR",
			Head: PRRef{
				Label: "test-owner:feature-branch",
				Ref:   "feature-branch",
				SHA:   "abc123",
				Repo: Repository{
					FullName: "test-owner/test-repo",
				},
			},
			Base: PRRef{
				Label: "test-owner:main",
				Ref:   "main",
				SHA:   "def456",
				Repo: Repository{
					FullName: "test-owner/test-repo",
				},
			},
		},
		Repository: Repository{
			Name:     "test-repo",
			FullName: "test-owner/test-repo",
		},
		Sender: Actor{
			Login: "test-user",
		},
	}
}

// ForkPullRequestEvent creates a PR event from a fork.
func ForkPullRequestEvent() *PullRequestEvent {
	event := DefaultPullRequestEvent()
	event.PullRequest.Head.Repo.FullName = "fork-owner/test-repo"
	event.PullRequest.Head.Label = "fork-owner:feature-branch"
	return event
}

// WriteEventFile writes an event payload to a temporary file.
func WriteEventFile(event interface{}, dir string) (string, error) {
	data, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling event: %w", err)
	}

	path := filepath.Join(dir, "event.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("writing event file: %w", err)
	}

	return path, nil
}

// LoadEventFile reads an event payload from a file.
func LoadEventFile(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading event file: %w", err)
	}

	var event map[string]interface{}
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("parsing event: %w", err)
	}

	return event, nil
}
