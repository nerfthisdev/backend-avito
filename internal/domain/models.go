package domain

import "time"

type Team struct {
	Name    string
	Members []User
}

type User struct {
	ID       string
	Username string
	TeamName string
	IsActive bool
}

type PRStatus string

const (
	PRStatusOpen    PRStatus = "OPEN"
	PRStatusMergerd PRStatus = "MERGED"
)

type PullRequest struct {
	ID                string
	Name              string
	AuthorID          string
	Status            PRStatus
	AssignedReviewers []string
	CreatedAt         time.Time
	MergedAt          *time.Time
}

var (
	ErrPullRequestExists = ErrWithCode{"PR_EXISTS"}
	ErrPullRequestMerged = ErrWithCode{"PR_MERGED"}
	ErrNotAssigned       = ErrWithCode{"NOT_ASSIGNED"}
	ErrNoCandidate       = ErrWithCode{"NO_CANDIDATE"}
)

type ErrWithCode struct {
	Code string
}

func (e ErrWithCode) Error() string { return e.Code }
