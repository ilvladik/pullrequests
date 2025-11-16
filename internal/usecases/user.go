package usecases

import (
	"context"
	"pullrequests/internal/domain"
	"pullrequests/internal/dtos"
)

type UserUsecase struct {
	userRepo domain.UserRepo
	trm      domain.TransactionManager
}

func NewUserUsecase(userRepo domain.UserRepo, trm domain.TransactionManager) *UserUsecase {
	return &UserUsecase{
		userRepo: userRepo,
		trm:      trm,
	}
}

func (u *UserUsecase) SetUserActive(ctx context.Context, req dtos.UserActiveRequest) (*dtos.UserResponse, error) {
	var user *domain.User
	err := u.trm.Do(ctx, func(ctx context.Context) error {
		existingUser, err := u.userRepo.GetUserByID(ctx, req.UserID)
		if err != nil {
			return err
		}
		if existingUser == nil {
			return domain.NewDomainError(domain.ErrNotFoundCode)
		}

		existingUser.IsActive = req.IsActive

		if err := u.userRepo.UpdateUser(ctx, existingUser); err != nil {
			return err
		}

		user = existingUser

		return nil
	})

	if err != nil {
		return nil, err
	}

	response := &dtos.UserResponse{
		User: dtos.User{
			UserID:   user.UserID,
			Username: user.Username,
			TeamName: user.TeamName,
			IsActive: user.IsActive,
		},
	}
	return response, nil
}
