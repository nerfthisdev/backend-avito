package dto

import "github.com/nerfthisdev/backend-avito/internal/domain"

type PullRequestShortDTO struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
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
