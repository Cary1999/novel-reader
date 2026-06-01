package httphandler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"novel-reader/backend/internal/auth"
	"novel-reader/backend/internal/domain"
)

type AuthUseCase interface {
	Register(ctx context.Context, username, password string) (domain.User, error)
	Login(ctx context.Context, username, password string) (string, domain.User, error)
	CurrentUser(ctx context.Context, claims auth.Claims) (domain.User, error)
	UpdateNickname(ctx context.Context, claims auth.Claims, nickname string) (domain.User, error)
	ChangePassword(ctx context.Context, claims auth.Claims, oldPassword, newPassword string) error
}

type BookUseCase interface {
	ListCategories(ctx context.Context) ([]domain.Category, error)
	SearchBooks(ctx context.Context, q, category string, page, pageSize int) ([]domain.Book, int, error)
	ListRecommendedBooks(ctx context.Context, page, pageSize int) ([]domain.Book, int, error)
	GetBook(ctx context.Context, id int64) (domain.Book, error)
	ListChapters(ctx context.Context, bookID int64) ([]domain.Chapter, error)
	GetChapter(ctx context.Context, bookID, chapterID int64) (domain.Chapter, error)
}

type AdminUseCase interface {
	ListBooks(ctx context.Context, q, category string, page, pageSize int) ([]domain.Book, int, error)
	ListMyBooks(ctx context.Context, ownerID int64, q string, categoryID int64, page, pageSize int) ([]domain.Book, int, error)
	CreateBook(ctx context.Context, ownerID int64, input domain.UploadBookInput) (domain.Book, error)
	UploadBook(ctx context.Context, ownerID int64, input domain.UploadBookInput, originalName string, reader io.Reader) (domain.UploadResult, error)
	UpdateBook(ctx context.Context, actorID int64, role domain.Role, bookID int64, input domain.BookMetadataInput) (domain.Book, error)
	UploadCover(ctx context.Context, actorID int64, role domain.Role, bookID int64, originalName string, reader io.Reader) (string, error)
	DeleteBook(ctx context.Context, actorID int64, role domain.Role, bookID int64) error
	AddChapter(ctx context.Context, actorID int64, role domain.Role, bookID int64, input domain.ChapterInput) (domain.Chapter, error)
	UpdateChapter(ctx context.Context, actorID int64, role domain.Role, bookID, chapterID int64, input domain.ChapterInput) (domain.Chapter, error)
	DeleteChapter(ctx context.Context, actorID int64, role domain.Role, bookID, chapterID int64) error
	CreateCategory(ctx context.Context, name string) (domain.Category, error)
	UpdateCategory(ctx context.Context, id int64, name string) (domain.Category, error)
	DeleteCategory(ctx context.Context, id int64) error
}

type Handler struct {
	authSvc        AuthUseCase
	bookSvc        BookUseCase
	adminSvc       AdminUseCase
	tokens         *auth.Manager
	maxUploadBytes int64
	coverDir       string
	defaultCover   string
}

type contextKey string

const claimsKey contextKey = "claims"

func New(authSvc AuthUseCase, bookSvc BookUseCase, adminSvc AdminUseCase, tokens *auth.Manager, maxUploadBytes int64, coverDir string, defaultCover string) *Handler {
	return &Handler{
		authSvc:        authSvc,
		bookSvc:        bookSvc,
		adminSvc:       adminSvc,
		tokens:         tokens,
		maxUploadBytes: maxUploadBytes,
		coverDir:       coverDir,
		defaultCover:   strings.TrimSpace(defaultCover),
	}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("GET /healthz", h.Health)
	mux.HandleFunc("GET /api/health", h.Health)
	mux.HandleFunc("POST /api/auth/register", h.Register)
	mux.HandleFunc("POST /api/auth/login", h.Login)
	mux.HandleFunc("GET /api/auth/me", h.requireAuth(h.Me))
	mux.HandleFunc("PATCH /api/auth/me", h.requireAuth(h.UpdateMe))
	mux.HandleFunc("PATCH /api/auth/password", h.requireAuth(h.ChangePassword))
	mux.HandleFunc("GET /api/categories", h.Categories)
	mux.HandleFunc("GET /api/books/search", h.SearchBooks)
	mux.HandleFunc("GET /api/books/recommendations", h.Recommendations)
	mux.HandleFunc("GET /api/books/{bookId}", h.BookDetail)
	mux.HandleFunc("GET /api/books/{bookId}/chapters", h.ChapterList)
	mux.HandleFunc("GET /api/books/{bookId}/chapters/{chapterId}", h.requireAuth(h.ChapterDetail))
	mux.HandleFunc("GET /api/books/{bookId}/cover", h.BookCover)
	mux.HandleFunc("POST /api/books/{bookId}/cover", h.requireAuth(h.UploadBookCover))
	mux.HandleFunc("GET /api/me/books", h.requireAuth(h.MyBooks))
	mux.HandleFunc("POST /api/me/books", h.requireAuth(h.CreateMyBook))
	mux.HandleFunc("POST /api/me/books/upload", h.requireAuth(h.UploadBook))
	mux.HandleFunc("PATCH /api/books/{bookId}", h.requireAuth(h.UpdateBook))
	mux.HandleFunc("DELETE /api/books/{bookId}", h.requireAuth(h.DeleteBook))
	mux.HandleFunc("POST /api/books/{bookId}/chapters", h.requireAuth(h.AddChapter))
	mux.HandleFunc("PATCH /api/books/{bookId}/chapters/{chapterId}", h.requireAuth(h.UpdateChapter))
	mux.HandleFunc("DELETE /api/books/{bookId}/chapters/{chapterId}", h.requireAuth(h.DeleteChapter))
	mux.HandleFunc("POST /api/admin/categories", h.requireAuth(h.requireAdmin(h.CreateCategory)))
	mux.HandleFunc("PATCH /api/admin/categories/{categoryId}", h.requireAuth(h.requireAdmin(h.UpdateCategory)))
	mux.HandleFunc("DELETE /api/admin/categories/{categoryId}", h.requireAuth(h.requireAdmin(h.DeleteCategory)))
	mux.HandleFunc("GET /api/admin/books", h.requireAuth(h.requireAdmin(h.AdminBooks)))
	mux.HandleFunc("POST /api/admin/books", h.requireAuth(h.requireAdmin(h.CreateMyBook)))
	mux.HandleFunc("POST /api/admin/books/upload", h.requireAuth(h.requireAdmin(h.UploadBook)))
	mux.HandleFunc("PATCH /api/admin/books/{bookId}", h.requireAuth(h.requireAdmin(h.UpdateBook)))
	mux.HandleFunc("DELETE /api/admin/books/{bookId}", h.requireAuth(h.requireAdmin(h.DeleteBook)))
	mux.HandleFunc("POST /api/admin/books/{bookId}/chapters", h.requireAuth(h.requireAdmin(h.AddChapter)))
	mux.HandleFunc("PATCH /api/admin/books/{bookId}/chapters/{chapterId}", h.requireAuth(h.requireAdmin(h.UpdateChapter)))
	mux.HandleFunc("DELETE /api/admin/books/{bookId}/chapters/{chapterId}", h.requireAuth(h.requireAdmin(h.DeleteChapter)))
	return h.withCommonHeaders(mux)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	user, err := h.authSvc.Register(r.Context(), req.Username, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"userId": user.ID, "username": user.Username, "nickname": user.Nickname})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	token, user, err := h.authSvc.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"token": token,
		"user":  publicUser(user),
	})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user, err := h.authSvc.CurrentUser(r.Context(), mustClaims(r))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, publicUser(user))
}

func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Nickname string `json:"nickname"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	user, err := h.authSvc.UpdateNickname(r.Context(), mustClaims(r), req.Nickname)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, publicUser(user))
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if err := h.authSvc.ChangePassword(r.Context(), mustClaims(r), req.OldPassword, req.NewPassword); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"changed": true})
}

func (h *Handler) Categories(w http.ResponseWriter, r *http.Request) {
	items, err := h.bookSvc.ListCategories(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) SearchBooks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page := parseInt(query.Get("page"), 1)
	pageSize := parseInt(query.Get("pageSize"), 20)
	items, total, err := h.bookSvc.SearchBooks(r.Context(), query.Get("q"), query.Get("category"), page, pageSize)
	if err != nil {
		writeError(w, err)
		return
	}
	for i := range items {
		items[i].CoverURL = "/api/books/" + strconv.FormatInt(items[i].ID, 10) + "/cover"
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": total})
}

func (h *Handler) Recommendations(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page := parseInt(query.Get("page"), 1)
	pageSize := parseInt(query.Get("pageSize"), 20)
	items, total, err := h.bookSvc.ListRecommendedBooks(r.Context(), page, pageSize)
	if err != nil {
		writeError(w, err)
		return
	}
	for i := range items {
		items[i].CoverURL = "/api/books/" + strconv.FormatInt(items[i].ID, 10) + "/cover"
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": total})
}

func (h *Handler) BookDetail(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathInt(w, r, "bookId")
	if !ok {
		return
	}
	book, err := h.bookSvc.GetBook(r.Context(), bookID)
	if err != nil {
		writeError(w, err)
		return
	}
	book.CoverURL = "/api/books/" + strconv.FormatInt(book.ID, 10) + "/cover"
	writeJSON(w, http.StatusOK, book)
}

func (h *Handler) ChapterList(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathInt(w, r, "bookId")
	if !ok {
		return
	}
	items, err := h.bookSvc.ListChapters(r.Context(), bookID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) ChapterDetail(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathInt(w, r, "bookId")
	if !ok {
		return
	}
	chapterID, ok := pathInt(w, r, "chapterId")
	if !ok {
		return
	}
	chapter, err := h.bookSvc.GetChapter(r.Context(), bookID, chapterID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, chapter)
}

func (h *Handler) UploadBookCover(w http.ResponseWriter, r *http.Request) {
	// Covers are limited separately from txt uploads; keep a small overhead for multipart.
	const maxCoverBytes = 10 * 1024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, maxCoverBytes+1024*1024)
	if err := r.ParseMultipartForm(maxCoverBytes + 1024*1024); err != nil {
		writeError(w, domain.NewError(http.StatusBadRequest, "BAD_REQUEST", "invalid multipart upload"))
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, domain.NewError(http.StatusBadRequest, "BAD_REQUEST", "file is required"))
		return
	}
	defer file.Close()

	bookID, ok := pathInt(w, r, "bookId")
	if !ok {
		return
	}
	claims := mustClaims(r)
	coverURL, err := h.adminSvc.UploadCover(r.Context(), claims.UserID, claims.Role, bookID, header.Filename, file)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"coverUrl": coverURL})
}

func (h *Handler) BookCover(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathInt(w, r, "bookId")
	if !ok {
		return
	}
	book, err := h.bookSvc.GetBook(r.Context(), bookID)
	if err != nil {
		writeError(w, err)
		return
	}

	// If no cover is set, return a built-in placeholder image (HTTP 200).
	if book.CoverPath == nil || strings.TrimSpace(*book.CoverPath) == "" {
		h.writeDefaultCover(w)
		return
	}

	fullPath := filepath.Join(h.coverDir, *book.CoverPath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		// If the file is missing on disk, behave like "no cover" rather than breaking UI.
		h.writeDefaultCover(w)
		return
	}

	w.Header().Set("Content-Type", coverContentTypeByExt(*book.CoverPath))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func coverContentTypeByExt(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}

func (h *Handler) writeDefaultCover(w http.ResponseWriter) {
	if h.defaultCover != "" {
		path := h.defaultCover
		if !filepath.IsAbs(path) {
			path = filepath.Join(h.coverDir, path)
		}
		data, err := os.ReadFile(path)
		if err == nil {
			w.Header().Set("Content-Type", coverContentTypeByExt(path))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(data)
			return
		}
	}
	// Fallback: 1x1 transparent PNG to guarantee HTTP 200 and avoid broken images.
	const placeholderBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO3Z9p0AAAAASUVORK5CYII="
	data, err := base64.StdEncoding.DecodeString(placeholderBase64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (h *Handler) AdminBooks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page := parseInt(query.Get("page"), 1)
	pageSize := parseInt(query.Get("pageSize"), 20)
	items, total, err := h.adminSvc.ListBooks(r.Context(), query.Get("q"), query.Get("category"), page, pageSize)
	if err != nil {
		writeError(w, err)
		return
	}
	for i := range items {
		items[i].CoverURL = "/api/books/" + strconv.FormatInt(items[i].ID, 10) + "/cover"
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": total})
}

func (h *Handler) MyBooks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page := parseInt(query.Get("page"), 1)
	pageSize := parseInt(query.Get("pageSize"), 20)
	categoryID := parseInt64(query.Get("categoryId"), 0)
	items, total, err := h.adminSvc.ListMyBooks(r.Context(), mustClaims(r).UserID, query.Get("q"), categoryID, page, pageSize)
	if err != nil {
		writeError(w, err)
		return
	}
	for i := range items {
		items[i].CoverURL = "/api/books/" + strconv.FormatInt(items[i].ID, 10) + "/cover"
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": total})
}

func (h *Handler) CreateMyBook(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title       string `json:"title"`
		CategoryID  int64  `json:"categoryId"`
		Description string `json:"description"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	book, err := h.adminSvc.CreateBook(r.Context(), mustClaims(r).UserID, domain.UploadBookInput{
		Title:       req.Title,
		CategoryID:  req.CategoryID,
		Description: req.Description,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, book)
}

func (h *Handler) UploadBook(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.maxUploadBytes+1024*1024)
	if err := r.ParseMultipartForm(h.maxUploadBytes + 1024*1024); err != nil {
		writeError(w, domain.NewError(http.StatusBadRequest, "BAD_REQUEST", "invalid multipart upload"))
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, domain.NewError(http.StatusBadRequest, "BAD_REQUEST", "file is required"))
		return
	}
	defer file.Close()

	result, err := h.adminSvc.UploadBook(r.Context(), mustClaims(r).UserID, domain.UploadBookInput{
		Title:       r.FormValue("title"),
		CategoryID:  parseInt64(r.FormValue("categoryId"), 0),
		Description: r.FormValue("description"),
	}, header.Filename, file)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathInt(w, r, "bookId")
	if !ok {
		return
	}
	var req domain.BookMetadataInput
	if !decodeJSON(w, r, &req) {
		return
	}
	claims := mustClaims(r)
	book, err := h.adminSvc.UpdateBook(r.Context(), claims.UserID, claims.Role, bookID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, book)
}

func (h *Handler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathInt(w, r, "bookId")
	if !ok {
		return
	}
	claims := mustClaims(r)
	if err := h.adminSvc.DeleteBook(r.Context(), claims.UserID, claims.Role, bookID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

func (h *Handler) AddChapter(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathInt(w, r, "bookId")
	if !ok {
		return
	}
	var req domain.ChapterInput
	if !decodeJSON(w, r, &req) {
		return
	}
	claims := mustClaims(r)
	chapter, err := h.adminSvc.AddChapter(r.Context(), claims.UserID, claims.Role, bookID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, chapter)
}

func (h *Handler) UpdateChapter(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathInt(w, r, "bookId")
	if !ok {
		return
	}
	chapterID, ok := pathInt(w, r, "chapterId")
	if !ok {
		return
	}
	var req domain.ChapterInput
	if !decodeJSON(w, r, &req) {
		return
	}
	claims := mustClaims(r)
	chapter, err := h.adminSvc.UpdateChapter(r.Context(), claims.UserID, claims.Role, bookID, chapterID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, chapter)
}

func (h *Handler) DeleteChapter(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathInt(w, r, "bookId")
	if !ok {
		return
	}
	chapterID, ok := pathInt(w, r, "chapterId")
	if !ok {
		return
	}
	claims := mustClaims(r)
	if err := h.adminSvc.DeleteChapter(r.Context(), claims.UserID, claims.Role, bookID, chapterID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	category, err := h.adminSvc.CreateCategory(r.Context(), req.Name)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, category)
}

func (h *Handler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	categoryID, ok := pathInt(w, r, "categoryId")
	if !ok {
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	category, err := h.adminSvc.UpdateCategory(r.Context(), categoryID, req.Name)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, category)
}

func (h *Handler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	categoryID, ok := pathInt(w, r, "categoryId")
	if !ok {
		return
	}
	if err := h.adminSvc.DeleteCategory(r.Context(), categoryID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

func (h *Handler) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raw, err := auth.BearerToken(r.Header.Get("Authorization"))
		if err != nil {
			writeError(w, domain.NewError(http.StatusUnauthorized, "UNAUTHORIZED", "login required"))
			return
		}
		claims, err := h.tokens.Parse(raw)
		if err != nil {
			writeError(w, domain.NewError(http.StatusUnauthorized, "UNAUTHORIZED", "login required"))
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), claimsKey, claims)))
	}
}

func (h *Handler) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if mustClaims(r).Role != domain.RoleAdmin {
			writeError(w, domain.NewError(http.StatusForbidden, "FORBIDDEN", "admin required"))
			return
		}
		next(w, r)
	}
}

func (h *Handler) withCommonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		writeError(w, domain.NewError(http.StatusBadRequest, "BAD_REQUEST", "invalid json body"))
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, err error) {
	var appErr *domain.AppError
	if errors.As(err, &appErr) {
		writeJSON(w, appErr.Status, map[string]string{"code": appErr.Code, "message": appErr.Message})
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]string{"code": "INTERNAL", "message": "internal server error"})
}

func publicUser(user domain.User) map[string]any {
	return map[string]any{"id": user.ID, "username": user.Username, "nickname": user.Nickname, "role": user.Role}
}

func mustClaims(r *http.Request) auth.Claims {
	claims, _ := r.Context().Value(claimsKey).(auth.Claims)
	return claims
}

func pathInt(w http.ResponseWriter, r *http.Request, key string) (int64, bool) {
	value, err := strconv.ParseInt(r.PathValue(key), 10, 64)
	if err != nil || value <= 0 {
		writeError(w, domain.NewError(http.StatusBadRequest, "BAD_REQUEST", "invalid path id"))
		return 0, false
	}
	return value, true
}

func parseInt(raw string, fallback int) int {
	if strings.TrimSpace(raw) == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func parseInt64(raw string, fallback int64) int64 {
	if strings.TrimSpace(raw) == "" {
		return fallback
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return fallback
	}
	return value
}
