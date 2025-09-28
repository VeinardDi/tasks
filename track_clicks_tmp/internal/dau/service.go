package dau

import (
	"context"

	"example.com/dau/internal/date"
)

type Repo interface {
	GetUniqueUsersCountForAuthors(ctx context.Context, authorIDs []int) ([]int, error)
	Set(ctx context.Context, userID int, authorID int) error
}

type Service struct {
	date date.Service
	repo Repo
}

type EventRequest struct {
	UserID   int `json:"user_id"`
	AuthorID int `json:"author_id"`
}

func (s *Service) Event(ctx context.Context, request *EventRequest) error {
	err := s.repo.Set(ctx, request.UserID, request.AuthorID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Dau(ctx context.Context, authorsList []int) ([]int, error) {
	return s.repo.GetUniqueUsersCountForAuthors(ctx, authorsList)
}

func NewService(date date.Service, repo Repo) *Service {
	return &Service{
		date: date,
		repo: repo,
	}
}
