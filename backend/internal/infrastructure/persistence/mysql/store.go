package mysql

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
	mysqlrepo "novel-reader/backend/internal/infrastructure/data/mysql/repo"
)

type Store struct {
	db  *sql.DB
	gdb *gorm.DB

	mysqlrepo.IdentityRepository
	mysqlrepo.CategoryRepository
	mysqlrepo.BookRepository
	mysqlrepo.UploadRepository
}

func NewStore(db *sql.DB, gdb *gorm.DB) *Store {
	return &Store{
		db:                 db,
		gdb:                gdb,
		IdentityRepository: mysqlrepo.NewIdentityRepository(db),
		CategoryRepository: mysqlrepo.NewCategoryRepository(db),
		BookRepository:     mysqlrepo.NewBookRepository(gdb),
		UploadRepository:   mysqlrepo.NewUploadRepository(db),
	}
}

func (s *Store) Migrate(ctx context.Context) error {
	return migrate(ctx, s.db)
}

func (s *Store) Seed(ctx context.Context, opts SeedOptions) error {
	return seed(ctx, s.db, opts)
}
