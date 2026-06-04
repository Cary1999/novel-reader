package main

import (
	"context"
	"database/sql"
	"log"
	"path/filepath"
	"time"

	"github.com/go-kratos/kratos/v2"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	_ "github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"

	bookapp "novel-reader/backend/internal/application/book"
	categoryapp "novel-reader/backend/internal/application/category"
	identityapp "novel-reader/backend/internal/application/identity"
	siteapp "novel-reader/backend/internal/application/site"
	uploadapp "novel-reader/backend/internal/application/upload"
	"novel-reader/backend/internal/infrastructure/auth/jwt"
	"novel-reader/backend/internal/infrastructure/config"
	"novel-reader/backend/internal/infrastructure/parser/txt"
	"novel-reader/backend/internal/infrastructure/persistence/mysql"
	"novel-reader/backend/internal/infrastructure/storage/local"
	httpapi "novel-reader/backend/internal/interface/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if cfg.GeneratedJWTSecret {
		log.Printf("JWT_SECRET was not set; generated an in-memory development secret")
	}
	if cfg.GeneratedSuperAdminPassword {
		log.Printf("SUPER_ADMIN_PASSWORD was not set; generated development super admin password for username %q: %s", cfg.SuperAdminUsername, cfg.SuperAdminPassword)
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

	gdb, err := gorm.Open(gormmysql.New(gormmysql.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		log.Fatalf("open gorm mysql: %v", err)
	}

	store := mysql.NewStore(db, gdb)
	if err := store.Migrate(ctx); err != nil {
		log.Fatalf("migrate mysql: %v", err)
	}
	if err := store.Seed(ctx, mysql.SeedOptions{
		SuperAdminUsername: cfg.SuperAdminUsername,
		SuperAdminPassword: cfg.SuperAdminPassword,
	}); err != nil {
		log.Fatalf("seed mysql: %v", err)
	}

	tokenManager := jwt.NewManager([]byte(cfg.JWTSecret), cfg.TokenTTL)
	uploadStore := local.NewStore(cfg.UploadDir, cfg.MaxUploadBytes)
	coverStore := local.NewStore(cfg.CoverDir, cfg.MaxCoverBytes)
	siteIconStore := local.NewStore(filepath.Join(cfg.CoverDir, "site"), cfg.MaxCoverBytes)
	parser := txt.NewParser()

	identityQueries := identityapp.NewQueries(store, store, store)
	identityCommands := identityapp.NewCommands(store, store, store, tokenManager, tokenManager)
	bookQueries := bookapp.NewQueries(store)
	bookCommands := bookapp.NewCommands(store, store)
	categoryQueries := categoryapp.NewQueries(store)
	categoryCommands := categoryapp.NewCommands(store)
	siteQueries := siteapp.NewQueries(store)
	siteCommands := siteapp.NewCommands(store, siteIconStore)
	uploadSvc := uploadapp.NewService(store, store, store, uploadStore, coverStore, parser)

	handler := httpapi.New(identityQueries, identityCommands, bookQueries, bookCommands, categoryQueries, categoryCommands, siteQueries, siteCommands, uploadSvc, tokenManager, cfg.MaxUploadBytes, cfg.MaxCoverBytes, cfg.CoverDir, filepath.Join(cfg.CoverDir, "site"), cfg.DefaultCoverFile)

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
