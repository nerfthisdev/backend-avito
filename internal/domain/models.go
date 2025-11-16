package domain

import (
	"errors"
	"time"

	"github.com/nerfthisdev/backend-avito/internal/apperror"
)

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
	ErrPullRequestExists = apperror.New(apperror.CodePullReqExists, errors.New("pull request already exists"))
	ErrPullRequestMerged = apperror.New(apperror.CodePullReqMerged, errors.New("pull request already merged"))
	ErrNotAssigned       = apperror.New(apperror.CodeNotAssigned, errors.New("reviewer is not assigned to this PR"))
	ErrNoCandidate       = apperror.New(apperror.CodeNoCandidate, errors.New("no active replacement candidate in team"))
)
