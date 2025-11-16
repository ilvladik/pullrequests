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
	UpdateUserActive(ctx context.Context, userID string, isActive bool) error
	GetActiveUsersByTeamName(ctx context.Context, teamName string) ([]User, error)
}
