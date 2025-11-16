package postgres

import (
	"context"
	"database/sql"
	"pullrequests/internal/domain"

	"github.com/jmoiron/sqlx"
)

type SQLTeamRepo struct {
	db *sqlx.DB
}

func NewTeamRepo(db *sqlx.DB) *SQLTeamRepo {
	return &SQLTeamRepo{db: db}
}

func (r *SQLTeamRepo) Add(ctx context.Context, team *domain.Team) error {
	query := "INSERT INTO teams (team_name) VALUES (:team_name)"

	_, err := sqlx.NamedExecContext(
		ctx,
		TxOrDb(ctx, r.db),
		query,
		r.toTeamRow(team),
	)
	return err
}

func (r *SQLTeamRepo) AddTeamMember(
	ctx context.Context,
	teamName string,
	teamMember *domain.TeamMember) error {
	query := "INSERT INTO users (user_id, username, team_name, is_active) VALUES (:user_id, :username, :team_name, :is_active)"

	_, err := sqlx.NamedExecContext(
		ctx,
		TxOrDb(ctx, r.db),
		query,
		r.toTeamMemberRow(teamName, teamMember),
	)

	return err
}

func (r *SQLTeamRepo) GetTeamByTeamName(ctx context.Context, teamName string) (*domain.Team, error) {
	tx := TxOrDb(ctx, r.db)

	team := &domain.Team{}
	query := "SELECT team_name FROM teams WHERE team_name = $1"
	row := tx.QueryRowxContext(ctx, query, teamName)

	if err := row.Scan(&team.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return team, nil
}

func (r *SQLTeamRepo) GetTeamMembersByTeamName(ctx context.Context, teamName string) ([]domain.TeamMember, error) {
	tx := TxOrDb(ctx, r.db)
	query := "SELECT user_id, username, is_active FROM users WHERE team_name = $1"
	rows, err := tx.QueryxContext(ctx, query, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := []domain.TeamMember{}
	for rows.Next() {
		var m domain.TeamMember
		if err := rows.Scan(&m.UserID, &m.Username, &m.IsActive); err != nil {
			return nil, err
		}
		members = append(members, m)
	}

	return members, nil
}

func (r *SQLTeamRepo) toTeamRow(team *domain.Team) map[string]interface{} {
	return map[string]interface{}{
		"team_name": team.Name,
	}
}

func (r *SQLTeamRepo) toTeamMemberRow(teamName string, member *domain.TeamMember) map[string]interface{} {
	return map[string]interface{}{
		"user_id":   member.UserID,
		"username":  member.Username,
		"team_name": teamName,
		"is_active": member.IsActive,
	}
}
