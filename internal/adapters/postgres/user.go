package postgres

import (
	"context"
	"database/sql"
	"pullrequests/internal/domain"

	"github.com/jmoiron/sqlx"
)

type SQLUserRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) *SQLUserRepo {
	return &SQLUserRepo{db: db}
}

func (r *SQLUserRepo) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	tx := TxOrDb(ctx, r.db)

	user := &domain.User{}
	query := "SELECT user_id, username, team_name, is_active FROM users WHERE user_id = $1"
	row := tx.QueryRowxContext(ctx, query, userID)

	if err := row.Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *SQLUserRepo) UpdateUserActive(ctx context.Context, userID string, isActive bool) error {
	query := "UPDATE users SET is_active = :is_active WHERE user_id = :user_id"

	_, err := sqlx.NamedExecContext(
		ctx,
		TxOrDb(ctx, r.db),
		query,
		map[string]interface{}{
			"user_id":   userID,
			"is_active": isActive,
		},
	)
	return err
}

func (r *SQLUserRepo) GetActiveUsersByTeamName(ctx context.Context, teamName string) ([]domain.User, error) {
	tx := TxOrDb(ctx, r.db)
	query := "SELECT user_id, username, team_name, is_active FROM users WHERE team_name = $1 AND is_active = true"
	rows, err := tx.QueryxContext(ctx, query, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
