package repository

import (
	"context"

	"aisearch/internal/model"
)

type WikiRepository struct {
	items []model.Wiki
}

func NewWikiRepository() *WikiRepository {
	return &WikiRepository{
		items: []model.Wiki{
			{
				ID:      "1",
				Title:   "Gin 项目骨架",
				Content: "这是一个示例 Wiki 条目，用来验证接口和分层结构。",
			},
		},
	}
}

func (r *WikiRepository) List(ctx context.Context) []model.Wiki {
	return r.items
}

func (r *WikiRepository) GetByID(ctx context.Context, id string) (model.Wiki, bool) {
	for _, item := range r.items {
		if item.ID == id {
			return item, true
		}
	}

	return model.Wiki{}, false
}
