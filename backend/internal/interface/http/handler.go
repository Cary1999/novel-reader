package httpapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	bookapp "novel-reader/backend/internal/application/book"
	bookcommand "novel-reader/backend/internal/application/book/command"
	bookquery "novel-reader/backend/internal/application/book/query"
	categoryapp "novel-reader/backend/internal/application/category"
	categorycommand "novel-reader/backend/internal/application/category/command"
	identityapp "novel-reader/backend/internal/application/identity"
	identitycommand "novel-reader/backend/internal/application/identity/command"
	siteapp "novel-reader/backend/internal/application/site"
	sitecommand "novel-reader/backend/internal/application/site/command"
	uploadapp "novel-reader/backend/internal/application/upload"
	identityentity "novel-reader/backend/internal/domain/identity/entity"
	"novel-reader/backend/internal/domain/shared"
	siteentity "novel-reader/backend/internal/domain/site/entity"
	"novel-reader/backend/internal/infrastructure/auth/jwt"
)

type Handler struct {
	identityQueries  *identityapp.Queries
	identityCommands *identityapp.Commands
	bookQueries      *bookapp.Queries
	bookCommands     *bookapp.Commands
	categoryQueries  *categoryapp.Queries
	categoryCommands *categoryapp.Commands
	siteQueries      *siteapp.Queries
	siteCommands     *siteapp.Commands
	uploadSvc        *uploadapp.Service
	tokens           *jwt.Manager
	maxUploadBytes   int64
	maxIconBytes     int64
	coverDir         string
	siteIconDir      string
	defaultCover     string
}

type contextKey string

const actorKey contextKey = "actor"

func New(identityQueries *identityapp.Queries, identityCommands *identityapp.Commands, bookQueries *bookapp.Queries, bookCommands *bookapp.Commands, categoryQueries *categoryapp.Queries, categoryCommands *categoryapp.Commands, siteQueries *siteapp.Queries, siteCommands *siteapp.Commands, uploadSvc *uploadapp.Service, tokens *jwt.Manager, maxUploadBytes int64, maxIconBytes int64, coverDir string, siteIconDir string, defaultCover string) *Handler {
	return &Handler{
		identityQueries:  identityQueries,
		identityCommands: identityCommands,
		bookQueries:      bookQueries,
		bookCommands:     bookCommands,
		categoryQueries:  categoryQueries,
		categoryCommands: categoryCommands,
		siteQueries:      siteQueries,
		siteCommands:     siteCommands,
		uploadSvc:        uploadSvc,
		tokens:           tokens,
		maxUploadBytes:   maxUploadBytes,
		maxIconBytes:     maxIconBytes,
		coverDir:         coverDir,
		siteIconDir:      siteIconDir,
		defaultCover:     strings.TrimSpace(defaultCover),
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
	mux.HandleFunc("GET /api/site-settings", h.SiteSettings)
	mux.HandleFunc("GET /api/site-settings/icon", h.SiteIcon)
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
	mux.HandleFunc("PATCH /api/admin/site-settings", h.requireAuth(h.requireAdmin(h.UpdateSiteSettings)))
	mux.HandleFunc("POST /api/admin/site-settings/icon", h.requireAuth(h.requireAdmin(h.UploadSiteIcon)))
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
	user, err := h.identityCommands.Register.Handle(r.Context(), identitycommand.Register{
		Username: req.Username,
		Password: req.Password,
	})
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
	result, err := h.identityCommands.Login.Handle(r.Context(), identitycommand.Login{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"token": result.Token,
		"user":  publicUser(result.User),
	})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user, err := h.identityQueries.CurrentUser.Handle(r.Context(), mustActor(r))
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
	user, err := h.identityCommands.UpdateNickname.Handle(r.Context(), mustActor(r), identitycommand.UpdateNickname{
		Nickname: req.Nickname,
	})
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
	if err := h.identityCommands.ChangePassword.Handle(r.Context(), mustActor(r), identitycommand.ChangePassword{
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	}); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"changed": true})
}

func (h *Handler) Categories(w http.ResponseWriter, r *http.Request) {
	items, err := h.categoryQueries.ListCategories.Handle(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) SiteSettings(w http.ResponseWriter, r *http.Request) {
	item, err := h.siteQueries.GetSettings.Handle(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, publicSiteSettings(item))
}

func (h *Handler) SearchBooks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page := parseInt(query.Get("page"), 1)
	pageSize := parseInt(query.Get("pageSize"), 20)
	items, total, err := h.bookQueries.SearchBooks.Handle(r.Context(), bookquery.SearchBooks{
		Keyword:      query.Get("q"),
		CategoryName: query.Get("category"),
		Page:         page,
		PageSize:     pageSize,
	})
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
	items, total, err := h.bookQueries.ListRecommended.Handle(r.Context(), page, pageSize)
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
	item, err := h.bookQueries.GetBook.Handle(r.Context(), bookID)
	if err != nil {
		writeError(w, err)
		return
	}
	item.CoverURL = "/api/books/" + strconv.FormatInt(item.ID, 10) + "/cover"
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) ChapterList(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathInt(w, r, "bookId")
	if !ok {
		return
	}
	items, err := h.bookQueries.ListChapters.Handle(r.Context(), bookID)
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
	item, err := h.bookQueries.GetChapter.Handle(r.Context(), bookID, chapterID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) UploadBookCover(w http.ResponseWriter, r *http.Request) {
	const maxCoverBytes = 10 * 1024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, maxCoverBytes+1024*1024)
	if err := r.ParseMultipartForm(maxCoverBytes + 1024*1024); err != nil {
		writeError(w, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "invalid multipart upload"))
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "file is required"))
		return
	}
	defer file.Close()

	bookID, ok := pathInt(w, r, "bookId")
	if !ok {
		return
	}
	coverURL, err := h.uploadSvc.UploadCover(r.Context(), mustActor(r), bookID, header.Filename, file)
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
	item, err := h.bookQueries.GetBook.Handle(r.Context(), bookID)
	if err != nil {
		writeError(w, err)
		return
	}
	if item.CoverPath == nil || strings.TrimSpace(*item.CoverPath) == "" {
		h.writeDefaultCover(w)
		return
	}

	fullPath := filepath.Join(h.coverDir, *item.CoverPath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		h.writeDefaultCover(w)
		return
	}

	w.Header().Set("Content-Type", imageContentTypeByExt(*item.CoverPath))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (h *Handler) AdminBooks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page := parseInt(query.Get("page"), 1)
	pageSize := parseInt(query.Get("pageSize"), 20)
	items, total, err := h.bookQueries.SearchBooks.Handle(r.Context(), bookquery.SearchBooks{
		Keyword:      query.Get("q"),
		CategoryName: query.Get("category"),
		Page:         page,
		PageSize:     pageSize,
	})
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
	items, total, err := h.bookQueries.ListOwnedBooks.Handle(r.Context(), bookquery.ListOwnedBooks{
		OwnerUserID: mustActor(r).UserID,
		Keyword:     query.Get("q"),
		CategoryID:  categoryID,
		Page:        page,
		PageSize:    pageSize,
	})
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
	item, err := h.bookCommands.CreateBook.Handle(r.Context(), mustActor(r), bookcommand.CreateBook{
		Title:       req.Title,
		CategoryID:  req.CategoryID,
		Description: req.Description,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) UploadBook(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.maxUploadBytes+1024*1024)
	if err := r.ParseMultipartForm(h.maxUploadBytes + 1024*1024); err != nil {
		writeError(w, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "invalid multipart upload"))
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "file is required"))
		return
	}
	defer file.Close()

	result, err := h.uploadSvc.UploadBook(r.Context(), mustActor(r), bookcommand.CreateBook{
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
	var req bookcommand.UpdateBookMetadata
	if !decodeJSON(w, r, &req) {
		return
	}
	item, err := h.bookCommands.UpdateBook.Handle(r.Context(), mustActor(r), bookID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathInt(w, r, "bookId")
	if !ok {
		return
	}
	if err := h.bookCommands.DeleteBook.Handle(r.Context(), mustActor(r), bookID); err != nil {
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
	var req bookcommand.SaveChapter
	if !decodeJSON(w, r, &req) {
		return
	}
	item, err := h.bookCommands.AddChapter.Handle(r.Context(), mustActor(r), bookID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
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
	var req bookcommand.SaveChapter
	if !decodeJSON(w, r, &req) {
		return
	}
	item, err := h.bookCommands.UpdateChapter.Handle(r.Context(), mustActor(r), bookID, chapterID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
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
	if err := h.bookCommands.DeleteChapter.Handle(r.Context(), mustActor(r), bookID, chapterID); err != nil {
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
	item, err := h.categoryCommands.CreateCategory.Handle(r.Context(), categorycommand.CreateCategory{
		Name: req.Name,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
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
	item, err := h.categoryCommands.UpdateCategory.Handle(r.Context(), categoryID, categorycommand.UpdateCategory{
		Name: req.Name,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	categoryID, ok := pathInt(w, r, "categoryId")
	if !ok {
		return
	}
	if err := h.categoryCommands.DeleteCategory.Handle(r.Context(), categoryID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

func (h *Handler) UpdateSiteSettings(w http.ResponseWriter, r *http.Request) {
	var req sitecommand.UpdateSettings
	if !decodeJSON(w, r, &req) {
		return
	}
	item, err := h.siteCommands.UpdateSettings.Handle(r.Context(), mustActor(r), req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, publicSiteSettings(item))
}

func (h *Handler) UploadSiteIcon(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.maxIconBytes+1024*1024)
	if err := r.ParseMultipartForm(h.maxIconBytes + 1024*1024); err != nil {
		writeError(w, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "invalid multipart upload"))
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "file is required"))
		return
	}
	defer file.Close()

	item, err := h.siteCommands.UpdateIcon.Handle(r.Context(), mustActor(r), header.Filename, file)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, publicSiteSettings(item))
}

func (h *Handler) SiteIcon(w http.ResponseWriter, r *http.Request) {
	item, err := h.siteQueries.GetSettings.Handle(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	if item.BrandIconPath == nil || strings.TrimSpace(*item.BrandIconPath) == "" {
		h.writeDefaultSiteIcon(w)
		return
	}

	fullPath := filepath.Join(h.siteIconDir, *item.BrandIconPath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		h.writeDefaultSiteIcon(w)
		return
	}
	w.Header().Set("Content-Type", imageContentTypeByExt(*item.BrandIconPath))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (h *Handler) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raw, err := jwt.BearerToken(r.Header.Get("Authorization"))
		if err != nil {
			writeError(w, shared.NewError(http.StatusUnauthorized, "UNAUTHORIZED", "login required"))
			return
		}
		claims, err := h.tokens.Parse(raw)
		if err != nil {
			writeError(w, shared.NewError(http.StatusUnauthorized, "UNAUTHORIZED", "login required"))
			return
		}
		actor := shared.Actor{
			UserID:   claims.UserID,
			Username: claims.Username,
			Role:     claims.Role,
		}
		next(w, r.WithContext(context.WithValue(r.Context(), actorKey, actor)))
	}
}

func (h *Handler) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if mustActor(r).Role != shared.RoleAdmin {
			writeError(w, shared.NewError(http.StatusForbidden, "FORBIDDEN", "admin required"))
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
		writeError(w, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "invalid json body"))
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
	var appErr *shared.AppError
	if errors.As(err, &appErr) {
		writeJSON(w, appErr.Status, map[string]string{"code": appErr.Code, "message": appErr.Message})
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]string{"code": "INTERNAL", "message": "internal server error"})
}

func publicUser(user identityentity.User) map[string]any {
	return map[string]any{"id": user.ID, "username": user.Username, "nickname": user.Nickname, "role": user.Role}
}

func publicSiteSettings(item siteentity.Settings) siteentity.Settings {
	item.BrandIconURL = "/api/site-settings/icon"
	if !item.UpdatedAt.IsZero() {
		item.BrandIconURL += "?v=" + strconv.FormatInt(item.UpdatedAt.UnixMilli(), 10)
	}
	return item
}

func mustActor(r *http.Request) shared.Actor {
	actor, _ := r.Context().Value(actorKey).(shared.Actor)
	return actor
}

func pathInt(w http.ResponseWriter, r *http.Request, key string) (int64, bool) {
	value, err := strconv.ParseInt(r.PathValue(key), 10, 64)
	if err != nil || value <= 0 {
		writeError(w, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "invalid path id"))
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

func imageContentTypeByExt(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
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
			w.Header().Set("Content-Type", imageContentTypeByExt(path))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(data)
			return
		}
	}
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

func (h *Handler) writeDefaultSiteIcon(w http.ResponseWriter) {
	const iconSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64" role="img" aria-label="site icon"><defs><linearGradient id="g" x1="0%" y1="0%" x2="100%" y2="100%"><stop offset="0%" stop-color="#f07a3a"/><stop offset="100%" stop-color="#b94a1a"/></linearGradient></defs><rect width="64" height="64" rx="18" fill="url(#g)"/><path d="M22 18h8c2.7 0 4.9 1 6 2.3 1.1-1.4 3.3-2.3 6-2.3h8a2 2 0 0 1 2 2v24a2 2 0 0 1-2 2h-8c-2.7 0-4.9 1-6 2.3-1.1-1.4-3.3-2.3-6-2.3h-8a2 2 0 0 1-2-2V20a2 2 0 0 1 2-2Zm2 4v20h6c2.3 0 4.3.6 6 1.8V23.8c-1.7-1.2-3.7-1.8-6-1.8Zm16 0c-2.3 0-4.3.6-6 1.8v20c1.7-1.2 3.7-1.8 6-1.8h6V22Z" fill="#fff"/></svg>`
	w.Header().Set("Content-Type", "image/svg+xml")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(iconSVG))
}
