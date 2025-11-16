package domain

import "time"

type PullRequestStatus string

const (
	PRStatusOpen   PullRequestStatus = "OPEN"
	PRStatusMerged PullRequestStatus = "MERGED"
)

type PullRequest struct {
	PullRequestID   string
	PullRequestName string
	AuthorID        string
	Status          PullRequestStatus
	CreatedAt       time.Time
	MergedAt        *time.Time
}

type PullRequestReviewer struct {
	PullRequestID string
	UserID        string
}
