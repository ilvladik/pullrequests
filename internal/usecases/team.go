package usecases

import (
	"context"
	"pullrequests/internal/domain"
	"pullrequests/internal/dtos"
)

type TeamUsecase struct {
	repo domain.TeamRepo
	trm  domain.TransactionManager
}

func NewTeamUsecase(repo domain.TeamRepo, trm domain.TransactionManager) *TeamUsecase {
	return &TeamUsecase{repo: repo, trm: trm}
}

func (u *TeamUsecase) AddTeam(ctx context.Context, in dtos.TeamRequest) (*dtos.TeamResponse, error) {
	err := u.trm.Do(ctx, func(ctx context.Context) error {
		team, err := u.repo.GetTeamByTeamName(ctx, in.TeamName)
		if err != nil {
			return err
		}
		if team != nil {
			return domain.NewDomainError(domain.ErrTeamExistsCode)
		}
		if err := u.repo.Add(ctx, &domain.Team{Name: in.TeamName}); err != nil {
			return err
		}
		for _, m := range in.Members {
			if err := u.repo.AddTeamMember(ctx, in.TeamName, &domain.TeamMember{
				UserID:   m.UserID,
				Username: m.Username,
				IsActive: m.IsActive}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	response := &dtos.TeamResponse{
		Team: dtos.Team{
			TeamName: in.TeamName,
			Members:  in.Members,
		},
	}
	return response, nil
}

func (u *TeamUsecase) GetTeam(ctx context.Context, teamName string) (*dtos.Team, error) {
	var out *dtos.Team
	err := u.trm.Do(ctx, func(ctx context.Context) error {
		team, err := u.repo.GetTeamByTeamName(ctx, teamName)
		if err != nil {
			return err
		}
		if team == nil {
			return domain.NewDomainError(domain.ErrNotFoundCode)
		}
		members, err := u.repo.GetTeamMembersByTeamName(ctx, teamName)
		if err != nil {
			return err
		}
		out = &dtos.Team{
			TeamName: team.Name,
			Members:  make([]dtos.TeamMember, 0, len(members)),
		}
		for _, m := range members {
			out.Members = append(out.Members, dtos.TeamMember{
				UserID:   m.UserID,
				Username: m.Username,
				IsActive: m.IsActive,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
