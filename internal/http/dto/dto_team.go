package dto

import "github.com/nerfthisdev/backend-avito/internal/domain"

type TeamMemberDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type TeamDTO struct {
	TeamName string          `json:"team_name"`
	Members  []TeamMemberDTO `json:"members"`
}

func TeamToDTO(t domain.Team) TeamDTO {
	members := make([]TeamMemberDTO, 0, len(t.Members))
	for _, m := range t.Members {
		members = append(members, TeamMemberDTO{
			UserID:   m.ID,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}
	return TeamDTO{
		TeamName: t.Name,
		Members:  members,
	}
}

func TeamFromDTO(dto TeamDTO) domain.Team {
	members := make([]domain.User, 0, len(dto.Members))
	for _, m := range dto.Members {
		members = append(members, domain.User{
			ID:       m.UserID,
			Username: m.Username,
			TeamName: dto.TeamName,
			IsActive: m.IsActive,
		})
	}
	return domain.Team{
		Name:    dto.TeamName,
		Members: members,
	}
}
