package usecases

import (
	"context"
	"pullrequests/internal/domain"
	"pullrequests/internal/dtos"
	"time"
)

type PRUsecase struct {
	pullrequestRepo domain.PullRequestRepo
	userRepo        domain.UserRepo
	trm             domain.TransactionManager
}

func NewPRUsecase(
	pullrequestRepo domain.PullRequestRepo,
	userRepo domain.UserRepo,
	trm domain.TransactionManager) *PRUsecase {
	return &PRUsecase{
		pullrequestRepo: pullrequestRepo,
		userRepo:        userRepo,
		trm:             trm,
	}
}

func (u *PRUsecase) CreatePR(ctx context.Context, req dtos.CreatePRRequest) (*dtos.PRResponse, error) {
	var pullrequest *domain.PullRequest
	var reviewers []string

	err := u.trm.Do(ctx, func(ctx context.Context) error {
		existingPR, err := u.pullrequestRepo.GetPullRequestByID(ctx, req.PullRequestID)
		if err != nil {
			return err
		}
		if existingPR != nil {
			return domain.NewDomainError(domain.ErrPRExistsCode)
		}

		author, err := u.userRepo.GetUserByID(ctx, req.AuthorID)
		if err != nil {
			return err
		}
		if author == nil {
			return domain.NewDomainError(domain.ErrNotFoundCode)
		}

		pullrequest = &domain.PullRequest{
			PullRequestID:   req.PullRequestID,
			PullRequestName: req.PullRequestName,
			AuthorID:        req.AuthorID,
			Status:          domain.PRStatusOpen,
			CreatedAt:       time.Now(),
		}

		if err := u.pullrequestRepo.Add(ctx, pullrequest); err != nil {
			return err
		}
		activeUsers, err := u.userRepo.GetActiveUsersByTeamName(ctx, author.TeamName)
		if err != nil {
			return err
		}

		reviewersAssigned := 0
		for _, member := range activeUsers {
			if member.UserID != author.UserID && reviewersAssigned < 2 {
				if err := u.pullrequestRepo.AddReviewer(ctx, pullrequest.PullRequestID, member.UserID); err != nil {
					return err
				}
				reviewers = append(reviewers, member.UserID)
				reviewersAssigned++
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	response := &dtos.PRResponse{
		PR: dtos.PullRequest{
			PullRequestID:     pullrequest.PullRequestID,
			PullRequestName:   pullrequest.PullRequestName,
			AuthorID:          pullrequest.AuthorID,
			Status:            string(pullrequest.Status),
			AssignedReviewers: reviewers,
			CreatedAt:         pullrequest.CreatedAt.Format(time.RFC3339),
		},
	}

	if pullrequest.MergedAt != nil {
		response.PR.MergedAt = pullrequest.MergedAt.Format(time.RFC3339)
	}
	return response, nil
}

func (u *PRUsecase) MergePR(ctx context.Context, req dtos.MergePRRequest) (*dtos.PRResponse, error) {
	var pullrequest *domain.PullRequest
	var reviewers []string

	err := u.trm.Do(ctx, func(ctx context.Context) error {
		existingPR, err := u.pullrequestRepo.GetPullRequestByID(ctx, req.PullRequestID)
		if err != nil {
			return err
		}
		if existingPR == nil {
			return domain.NewDomainError(domain.ErrNotFoundCode)
		}

		pullrequest = existingPR
		reviewerEntities, err := u.pullrequestRepo.GetReviewers(ctx, req.PullRequestID)
		if err != nil {
			return err
		}
		reviewers = make([]string, 0, len(reviewerEntities))
		for _, reviewerEntity := range reviewerEntities {
			reviewers = append(reviewers, reviewerEntity.UserID)
		}

		if existingPR.Status == domain.PRStatusMerged {
			return nil
		}

		pullrequest.Status = domain.PRStatusMerged
		mergedAt := time.Now()
		pullrequest.MergedAt = &mergedAt

		if err := u.pullrequestRepo.UpdatePullRequest(ctx, pullrequest); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	response := &dtos.PRResponse{
		PR: dtos.PullRequest{
			PullRequestID:     pullrequest.PullRequestID,
			PullRequestName:   pullrequest.PullRequestName,
			AuthorID:          pullrequest.AuthorID,
			Status:            string(pullrequest.Status),
			AssignedReviewers: reviewers,
			CreatedAt:         pullrequest.CreatedAt.Format(time.RFC3339),
		},
	}

	if pullrequest.MergedAt != nil {
		response.PR.MergedAt = pullrequest.MergedAt.Format(time.RFC3339)
	}
	return response, nil
}

func (u *PRUsecase) ReassignReviewer(ctx context.Context, req dtos.ReassignPRRequest) (*dtos.ReassignResponse, error) {
	var pullrequest *domain.PullRequest
	var reviewers []string
	var newReviewerID string

	err := u.trm.Do(ctx, func(ctx context.Context) error {
		existingPR, err := u.pullrequestRepo.GetPullRequestByID(ctx, req.PullRequestID)
		if err != nil {
			return err
		}
		if existingPR == nil {
			return domain.NewDomainError(domain.ErrNotFoundCode)
		}
		pullrequest = existingPR
		if pullrequest.Status == domain.PRStatusMerged {
			return domain.NewDomainError(domain.ErrPRMergedCode)
		}

		reviewerEntities, err := u.pullrequestRepo.GetReviewers(ctx, req.PullRequestID)
		if err != nil {
			return err
		}

		oldReviewerAssigned := false
		for _, reviewer := range reviewerEntities {
			if reviewer.UserID == req.OldUserID {
				oldReviewerAssigned = true
				break
			}
		}

		if !oldReviewerAssigned {
			return domain.NewDomainError(domain.ErrNotAssignedCode)
		}

		author, err := u.userRepo.GetUserByID(ctx, pullrequest.AuthorID)
		if err != nil {
			return err
		}
		if author == nil {
			return domain.NewDomainError(domain.ErrNotFoundCode)
		}

		activeUsers, err := u.userRepo.GetActiveUsersByTeamName(ctx, author.TeamName)
		if err != nil {
			return err
		}

		var candidate *domain.User
		for _, m := range activeUsers {
			if m.UserID != author.UserID &&
				m.UserID != req.OldUserID &&
				!u.isUserAssigned(m.UserID, reviewerEntities) {
				candidate = &m
				break
			}
		}

		if candidate == nil {
			return domain.NewDomainError(domain.ErrNoCandidateCode)
		}

		if err := u.pullrequestRepo.RemoveReviewer(ctx, req.PullRequestID, req.OldUserID); err != nil {
			return err
		}

		if err := u.pullrequestRepo.AddReviewer(ctx, req.PullRequestID, candidate.UserID); err != nil {
			return err
		}

		updatedReviewerEntities, err := u.pullrequestRepo.GetReviewers(ctx, req.PullRequestID)
		if err != nil {
			return err
		}
		reviewers = make([]string, 0, len(updatedReviewerEntities))
		for _, reviewerEntity := range updatedReviewerEntities {
			reviewers = append(reviewers, reviewerEntity.UserID)
		}
		newReviewerID = candidate.UserID
		return nil
	})

	if err != nil {
		return nil, err
	}

	response := &dtos.ReassignResponse{
		PR: dtos.PullRequest{
			PullRequestID:     pullrequest.PullRequestID,
			PullRequestName:   pullrequest.PullRequestName,
			AuthorID:          pullrequest.AuthorID,
			Status:            string(pullrequest.Status),
			AssignedReviewers: reviewers,
			CreatedAt:         pullrequest.CreatedAt.Format(time.RFC3339),
		},
		ReplacedBy: newReviewerID,
	}

	if pullrequest.MergedAt != nil {
		response.PR.MergedAt = pullrequest.MergedAt.Format(time.RFC3339)
	}
	return response, nil
}

func (u *PRUsecase) GetUserReviewPRs(ctx context.Context, userID string) (*dtos.UserReviewResponse, error) {
	var pullrequestShorts []dtos.PullRequestShort

	err := u.trm.Do(ctx, func(ctx context.Context) error {
		user, err := u.userRepo.GetUserByID(ctx, userID)
		if err != nil {
			return err
		}
		if user == nil {
			return domain.NewDomainError(domain.ErrNotFoundCode)
		}

		prs, err := u.pullrequestRepo.GetUserAssignedPRs(ctx, userID)
		if err != nil {
			return err
		}

		pullrequestShorts = make([]dtos.PullRequestShort, 0, len(prs))
		for _, pr := range prs {
			pullrequestShorts = append(pullrequestShorts, dtos.PullRequestShort{
				PullRequestID:   pr.PullRequestID,
				PullRequestName: pr.PullRequestName,
				AuthorID:        pr.AuthorID,
				Status:          string(pr.Status),
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	response := &dtos.UserReviewResponse{
		UserID:       userID,
		PullRequests: pullrequestShorts,
	}
	return response, nil
}

func (u *PRUsecase) isUserAssigned(userID string, reviewers []domain.PullRequestReviewer) bool {
	for _, reviewer := range reviewers {
		if reviewer.UserID == userID {
			return true
		}
	}
	return false
}
