package service

import (
	"context"
	"errors"

	"aisearch/internal/model"
	"aisearch/internal/repository"
)

var ErrWikiNotFound = errors.New("wiki not found")

type WikiService struct {
	repo *repository.WikiRepository
}

func NewWikiService(repo *repository.WikiRepository) *WikiService {
	return &WikiService{repo: repo}
}

func (s *WikiService) List(ctx context.Context) []model.Wiki {
	return s.repo.List(ctx)
}

func (s *WikiService) GetByID(ctx context.Context, id string) (model.Wiki, error) {
	wiki, ok := s.repo.GetByID(ctx, id)
	if !ok {
		return model.Wiki{}, ErrWikiNotFound
	}

	return wiki, nil
}
