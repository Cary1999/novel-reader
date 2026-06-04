package bookapp

import (
	bookcommand "novel-reader/backend/internal/application/book/command"
	bookquery "novel-reader/backend/internal/application/book/query"
	bookrepository "novel-reader/backend/internal/domain/book/repository"
	bookservice "novel-reader/backend/internal/domain/book/service"
	categoryrepository "novel-reader/backend/internal/domain/category/repository"
)

type Commands struct {
	CreateBook           *bookcommand.CreateBookHandler
	UpdateBook           *bookcommand.UpdateBookHandler
	UpdateRecommendScore *bookcommand.UpdateRecommendScoreHandler
	DeleteBook           *bookcommand.DeleteBookHandler
	AddChapter           *bookcommand.AddChapterHandler
	UpdateChapter        *bookcommand.UpdateChapterHandler
	DeleteChapter        *bookcommand.DeleteChapterHandler
}

type Queries struct {
	SearchBooks     *bookquery.SearchBooksHandler
	ListRecommended *bookquery.ListRecommendedBooksHandler
	GetBook         *bookquery.GetBookHandler
	ListChapters    *bookquery.ListChaptersHandler
	GetChapter      *bookquery.GetChapterHandler
	ListOwnedBooks  *bookquery.ListOwnedBooksHandler
}

func NewCommands(books bookrepository.BookRepository, categories categoryrepository.CategoryRepository) *Commands {
	bookDomainService := bookservice.NewBookService(books)
	return &Commands{
		CreateBook:           bookcommand.NewCreateBookHandler(books, bookDomainService, categories),
		UpdateBook:           bookcommand.NewUpdateBookHandler(books, bookDomainService, categories),
		UpdateRecommendScore: bookcommand.NewUpdateRecommendScoreHandler(books),
		DeleteBook:           bookcommand.NewDeleteBookHandler(books, bookDomainService),
		AddChapter:           bookcommand.NewAddChapterHandler(books, bookDomainService),
		UpdateChapter:        bookcommand.NewUpdateChapterHandler(books, bookDomainService),
		DeleteChapter:        bookcommand.NewDeleteChapterHandler(books, bookDomainService),
	}
}

func NewQueries(books bookrepository.BookRepository) *Queries {
	return &Queries{
		SearchBooks:     bookquery.NewSearchBooksHandler(books),
		ListRecommended: bookquery.NewListRecommendedBooksHandler(books),
		GetBook:         bookquery.NewGetBookHandler(books),
		ListChapters:    bookquery.NewListChaptersHandler(books),
		GetChapter:      bookquery.NewGetChapterHandler(books),
		ListOwnedBooks:  bookquery.NewListOwnedBooksHandler(books),
	}
}
