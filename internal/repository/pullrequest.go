package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nerfthisdev/backend-avito/internal/domain"
)

type PullRequestRepo struct {
	pool *pgxpool.Pool
}

func NewPullRequestRepo(pool *pgxpool.Pool) *PullRequestRepo {
	return &PullRequestRepo{pool: pool}
}

func (r *PullRequestRepo) Create(ctx context.Context, pr domain.PullRequest) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx,
		`INSERT INTO pull_requests (id, name, author_id, status)
		 VALUES ($1, $2, $3, $4)`,
		pr.ID, pr.Name, pr.AuthorID, pr.Status,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return ErrPRExists
		}
		return err
	}

	for _, reviewerID := range pr.AssignedReviewers {
		if _, err := tx.Exec(ctx,
			`INSERT INTO pull_request_reviewers (pr_id, user_id)
			 VALUES ($1, $2)`,
			pr.ID, reviewerID,
		); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (r *PullRequestRepo) GetByID(ctx context.Context, id string) (*domain.PullRequest, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT pr.id,
		       pr.name,
		       pr.author_id,
		       pr.status,
		       pr.created_at,
		       pr.merged_at,
		       COALESCE(array_agg(rev.user_id ORDER BY rev.user_id)
		                FILTER (WHERE rev.user_id IS NOT NULL), '{}') AS reviewers
		  FROM pull_requests pr
		  LEFT JOIN pull_request_reviewers rev ON rev.pr_id = pr.id
		 WHERE pr.id = $1
		 GROUP BY pr.id
	`, id)

	var (
		pr        domain.PullRequest
		status    string
		mergedAt  pgtype.Timestamptz
		reviewers []string
	)

	err := row.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &status, &pr.CreatedAt, &mergedAt, &reviewers)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	pr.Status = domain.PRStatus(status)
	if mergedAt.Valid {
		t := mergedAt.Time
		pr.MergedAt = &t
	}
	pr.AssignedReviewers = reviewers

	return &pr, nil
}

func (r *PullRequestRepo) SetMerged(ctx context.Context, id string, mergedAt time.Time) (*domain.PullRequest, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE pull_requests
		   SET status = 'MERGED',
		       merged_at = COALESCE(merged_at, $2)
		 WHERE id = $1
		 RETURNING id
	`, id, mergedAt)

	var prID string
	if err := row.Scan(&prID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return r.GetByID(ctx, prID)
}

func (r *PullRequestRepo) ReplaceReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var status string
	row := tx.QueryRow(ctx, `
		SELECT status
		  FROM pull_requests
		 WHERE id = $1
		 FOR UPDATE
	`, prID)
	if err := row.Scan(&status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	if domain.PRStatus(status) == domain.PRStatusMergerd {
		return domain.ErrPullRequestMerged
	}

	tag, err := tx.Exec(ctx, `
		DELETE FROM pull_request_reviewers
		 WHERE pr_id = $1 AND user_id = $2
	`, prID, oldReviewerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotAssigned
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO pull_request_reviewers (pr_id, user_id)
		VALUES ($1, $2)
	`, prID, newReviewerID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *PullRequestRepo) ListByReviewer(ctx context.Context, userID string) ([]domain.PullRequest, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT pr.id,
		       pr.name,
		       pr.author_id,
		       pr.status,
		       pr.created_at,
		       pr.merged_at,
		       COALESCE(array_agg(rev_all.user_id ORDER BY rev_all.user_id)
		                FILTER (WHERE rev_all.user_id IS NOT NULL), '{}') AS reviewers
		  FROM pull_requests pr
		  JOIN pull_request_reviewers rev ON rev.pr_id = pr.id AND rev.user_id = $1
		  LEFT JOIN pull_request_reviewers rev_all ON rev_all.pr_id = pr.id
		 GROUP BY pr.id
		 ORDER BY pr.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []domain.PullRequest
	for rows.Next() {
		var (
			pr        domain.PullRequest
			status    string
			mergedAt  pgtype.Timestamptz
			reviewers []string
		)
		if err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &status, &pr.CreatedAt, &mergedAt, &reviewers); err != nil {
			return nil, err
		}
		pr.Status = domain.PRStatus(status)
		if mergedAt.Valid {
			t := mergedAt.Time
			pr.MergedAt = &t
		}
		pr.AssignedReviewers = reviewers
		prs = append(prs, pr)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return prs, nil
}
