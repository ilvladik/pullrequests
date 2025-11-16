package domain

import "context"

type TeamRepo interface {
	Add(ctx context.Context, team *Team) error
	AddTeamMember(ctx context.Context, teamName string, teamMember *TeamMember) error
	GetTeamByTeamName(ctx context.Context, teamName string) (*Team, error)
	GetTeamMembersByTeamName(ctx context.Context, teamName string) ([]TeamMember, error)
}
