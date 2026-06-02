package categoryapp

import (
	categorycommand "novel-reader/backend/internal/application/category/command"
	categoryquery "novel-reader/backend/internal/application/category/query"
	categoryrepository "novel-reader/backend/internal/domain/category/repository"
	categoryservice "novel-reader/backend/internal/domain/category/service"
)

type Commands struct {
	CreateCategory *categorycommand.CreateCategoryHandler
	UpdateCategory *categorycommand.UpdateCategoryHandler
	DeleteCategory *categorycommand.DeleteCategoryHandler
}

type Queries struct {
	ListCategories *categoryquery.ListCategoriesHandler
}

func NewCommands(repo categoryrepository.CategoryRepository) *Commands {
	service := categoryservice.NewCategoryService()
	return &Commands{
		CreateCategory: categorycommand.NewCreateCategoryHandler(repo, service),
		UpdateCategory: categorycommand.NewUpdateCategoryHandler(repo, service),
		DeleteCategory: categorycommand.NewDeleteCategoryHandler(repo),
	}
}

func NewQueries(repo categoryrepository.CategoryRepository) *Queries {
	return &Queries{
		ListCategories: categoryquery.NewListCategoriesHandler(repo),
	}
}
