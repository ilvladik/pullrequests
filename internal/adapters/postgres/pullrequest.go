package postgres

import (
	"context"
	"database/sql"
	"pullrequests/internal/domain"

	"github.com/jmoiron/sqlx"
)

type SQLPRRepo struct {
	db *sqlx.DB
}

func NewPRRepo(db *sqlx.DB) *SQLPRRepo {
	return &SQLPRRepo{db: db}
}

func (r *SQLPRRepo) Add(ctx context.Context, pullrequest *domain.PullRequest) error {
	query := `
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at)
		VALUES (:pull_request_id, :pull_request_name, :author_id, :status, :created_at)
	`

	_, err := sqlx.NamedExecContext(
		ctx,
		TxOrDb(ctx, r.db),
		query,
		r.toPRRow(pullrequest),
	)
	return err
}

func (r *SQLPRRepo) GetPullRequestByID(ctx context.Context, pullrequestID string) (*domain.PullRequest, error) {
	tx := TxOrDb(ctx, r.db)

	pullrequest := &domain.PullRequest{}
	query := `
        SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
        FROM pull_requests WHERE pull_request_id = $1
    `
	row := tx.QueryRowxContext(ctx, query, pullrequestID)

	var mergedAt sql.NullTime
	err := row.Scan(&pullrequest.PullRequestID, &pullrequest.PullRequestName, &pullrequest.AuthorID, &pullrequest.Status, &pullrequest.CreatedAt, &mergedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if mergedAt.Valid {
		pullrequest.MergedAt = &mergedAt.Time
	}

	return pullrequest, nil
}

func (r *SQLPRRepo) UpdatePullRequest(ctx context.Context, pullrequest *domain.PullRequest) error {
	query := `
        UPDATE pull_requests
        SET pull_request_name = :pull_request_name, author_id = :author_id,
            status = :status, created_at = :created_at, merged_at = :merged_at
        WHERE pull_request_id = :pull_request_id
    `

	_, err := sqlx.NamedExecContext(
		ctx,
		TxOrDb(ctx, r.db),
		query,
		r.toPRRow(pullrequest),
	)
	return err
}

func (r *SQLPRRepo) AddReviewer(ctx context.Context, pullrequestID, userID string) error {
	query := "INSERT INTO pull_request_reviewers (pull_request_id, user_id) VALUES (:pull_request_id, :user_id)"

	_, err := sqlx.NamedExecContext(
		ctx,
		TxOrDb(ctx, r.db),
		query,
		map[string]interface{}{
			"pull_request_id": pullrequestID,
			"user_id":         userID,
		},
	)
	return err
}

func (r *SQLPRRepo) GetReviewers(ctx context.Context, pullrequestID string) ([]domain.PullRequestReviewer, error) {
	tx := TxOrDb(ctx, r.db)
	query := "SELECT pull_request_id, user_id FROM pull_request_reviewers WHERE pull_request_id = $1"
	rows, err := tx.QueryxContext(ctx, query, pullrequestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviewers []domain.PullRequestReviewer
	for rows.Next() {
		var reviewer domain.PullRequestReviewer
		if err := rows.Scan(&reviewer.PullRequestID, &reviewer.UserID); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, reviewer)
	}
	return reviewers, nil
}

func (r *SQLPRRepo) RemoveReviewer(ctx context.Context, pullrequestID, userID string) error {
	query := "DELETE FROM pull_request_reviewers WHERE pull_request_id = $1 AND user_id = $2"

	_, err := TxOrDb(ctx, r.db).ExecContext(ctx, query, pullrequestID, userID)
	return err
}

func (r *SQLPRRepo) toPRRow(pullrequest *domain.PullRequest) map[string]interface{} {
	row := map[string]interface{}{
		"pull_request_id":   pullrequest.PullRequestID,
		"pull_request_name": pullrequest.PullRequestName,
		"author_id":         pullrequest.AuthorID,
		"status":            pullrequest.Status,
		"created_at":        pullrequest.CreatedAt,
	}

	if pullrequest.MergedAt != nil {
		row["merged_at"] = *pullrequest.MergedAt
	} else {
		row["merged_at"] = nil
	}

	return row
}

func (r *SQLPRRepo) GetUserAssignedPRs(ctx context.Context, userID string) ([]domain.PullRequest, error) {
	tx := TxOrDb(ctx, r.db)

	query := `
        SELECT
            pr.pull_request_id,
            pr.pull_request_name,
            pr.author_id,
            pr.status,
            pr.created_at,
            pr.merged_at
        FROM pull_requests pr
        JOIN pull_request_reviewers prr ON pr.pull_request_id = prr.pull_request_id
        WHERE prr.user_id = $1
        ORDER BY pr.created_at DESC
    `

	rows, err := tx.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pullrequests []domain.PullRequest
	for rows.Next() {
		var pullrequest domain.PullRequest
		var mergedAt sql.NullTime

		err := rows.Scan(
			&pullrequest.PullRequestID,
			&pullrequest.PullRequestName,
			&pullrequest.AuthorID,
			&pullrequest.Status,
			&pullrequest.CreatedAt,
			&mergedAt,
		)
		if err != nil {
			return nil, err
		}

		if mergedAt.Valid {
			pullrequest.MergedAt = &mergedAt.Time
		}

		pullrequests = append(pullrequests, pullrequest)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return pullrequests, nil
}
