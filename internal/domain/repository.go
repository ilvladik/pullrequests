package domain

import (
	"context"
)

type TeamRepo interface {
	Add(ctx context.Context, team *Team) error
	AddTeamMember(ctx context.Context, teamName string, teamMember *TeamMember) error
	GetTeamByTeamName(ctx context.Context, teamName string) (*Team, error)
	GetTeamMembersByTeamName(ctx context.Context, teamName string) ([]TeamMember, error)
}

type UserRepo interface {
	GetUserByID(ctx context.Context, userID string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	GetActiveUsersByTeamName(ctx context.Context, teamName string) ([]User, error)
}

type PullRequestRepo interface {
	Add(ctx context.Context, pullrequest *PullRequest) error
	GetPullRequestByID(ctx context.Context, pullrequestID string) (*PullRequest, error)
	AddReviewer(ctx context.Context, pullrequestID, userID string) error
	RemoveReviewer(ctx context.Context, pullrequestID, userID string) error
	GetReviewers(ctx context.Context, pullrequestID string) ([]PullRequestReviewer, error)
	UpdatePullRequest(ctx context.Context, pullrequest *PullRequest) error
	GetUserAssignedPRs(ctx context.Context, userID string) ([]PullRequest, error)
}
