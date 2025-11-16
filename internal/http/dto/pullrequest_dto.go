package dto

import (
	"time"

	"github.com/nerfthisdev/backend-avito/internal/domain"
)

type PullRequestDTO struct {
	PullRequestID     string     `json:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            string     `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         *time.Time `json:"createdAt,omitempty"`
	MergedAt          *time.Time `json:"mergedAt,omitempty"`
}

type PullRequestShortDTO struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}

func PullRequestToDTO(pr domain.PullRequest) PullRequestDTO {
	dto := PullRequestDTO{
		PullRequestID:     pr.ID,
		PullRequestName:   pr.Name,
		AuthorID:          pr.AuthorID,
		Status:            string(pr.Status),
		AssignedReviewers: append([]string(nil), pr.AssignedReviewers...),
	}
	if !pr.CreatedAt.IsZero() {
		t := pr.CreatedAt
		dto.CreatedAt = &t
	}
	if pr.MergedAt != nil {
		t := *pr.MergedAt
		dto.MergedAt = &t
	}
	return dto
}

func PullRequestShortToDTO(pr domain.PullRequest) PullRequestShortDTO {
	return PullRequestShortDTO{
		PullRequestID:   pr.ID,
		PullRequestName: pr.Name,
		AuthorID:        pr.AuthorID,
		Status:          string(pr.Status),
	}
}

func PullRequestsShortToDTO(prs []domain.PullRequest) []PullRequestShortDTO {
	res := make([]PullRequestShortDTO, 0, len(prs))
	for _, pr := range prs {
		res = append(res, PullRequestShortToDTO(pr))
	}
	return res
}

type CreatePullRequestRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

type MergePullRequestRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

type ReassignPullRequestRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_user_id"`
}
