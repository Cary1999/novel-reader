package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/go-kratos/kratos/v2"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	_ "github.com/go-sql-driver/mysql"

	"novel-reader/backend/internal/auth"
	"novel-reader/backend/internal/config"
	"novel-reader/backend/internal/httphandler"
	"novel-reader/backend/internal/repository"
	"novel-reader/backend/internal/service"
	"novel-reader/backend/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if cfg.GeneratedJWTSecret {
		log.Printf("JWT_SECRET was not set; generated an in-memory development secret")
	}
	if cfg.GeneratedAdminPassword {
		log.Printf("ADMIN_PASSWORD was not set; generated development admin password for username %q: %s", cfg.AdminUsername, cfg.AdminPassword)
	}

	db, err := sql.Open("mysql", cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("open mysql: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("ping mysql: %v", err)
	}

	store := repository.NewMySQLStore(db)
	if err := store.Migrate(ctx); err != nil {
		log.Fatalf("migrate mysql: %v", err)
	}
	if err := store.Seed(ctx, repository.SeedOptions{
		AdminUsername: cfg.AdminUsername,
		AdminPassword: cfg.AdminPassword,
	}); err != nil {
		log.Fatalf("seed mysql: %v", err)
	}

	tokenManager := auth.NewManager([]byte(cfg.JWTSecret), cfg.TokenTTL)
	uploadStore := storage.NewLocalStore(cfg.UploadDir, cfg.MaxUploadBytes)
	coverStore := storage.NewLocalStore(cfg.CoverDir, cfg.MaxCoverBytes)
	authSvc := service.NewAuthService(store, tokenManager)
	bookSvc := service.NewBookService(store)
	adminSvc := service.NewAdminService(store, uploadStore, coverStore)
	handler := httphandler.New(authSvc, bookSvc, adminSvc, tokenManager, cfg.MaxUploadBytes, cfg.CoverDir, cfg.DefaultCoverFile)

	httpSrv := khttp.NewServer(khttp.Address(cfg.HTTPAddr), khttp.Timeout(15*time.Second))
	httpSrv.HandlePrefix("/", handler.Routes())

	app := kratos.New(
		kratos.Name("novel-reader-backend"),
		kratos.Server(httpSrv),
	)
	if err := app.Run(); err != nil {
		log.Fatalf("run app: %v", err)
	}
}
