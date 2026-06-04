package httpapi

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	bookshelfapp "novel-reader/backend/internal/application/bookshelf"
	"novel-reader/backend/internal/domain/shared"
)

func (h *Handler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.maxAvatarBytes+1024*1024)
	if err := r.ParseMultipartForm(h.maxAvatarBytes + 1024*1024); err != nil {
		writeError(w, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "invalid multipart upload"))
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "file is required"))
		return
	}
	defer file.Close()

	user, err := h.identityCommands.UpdateAvatar.Handle(r.Context(), mustActor(r), header.Filename, file)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, publicUser(user))
}

func (h *Handler) UserAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := pathInt(w, r, "userId")
	if !ok {
		return
	}
	user, err := h.identityQueries.GetUser.Handle(r.Context(), userID)
	if err != nil {
		h.writeDefaultAvatar(w)
		return
	}
	if user.AvatarPath == nil || strings.TrimSpace(*user.AvatarPath) == "" {
		h.writeDefaultAvatar(w)
		return
	}

	fullPath := filepath.Join(h.avatarDir, *user.AvatarPath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		h.writeDefaultAvatar(w)
		return
	}
	w.Header().Set("Content-Type", imageContentTypeByExt(*user.AvatarPath))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (h *Handler) ListMyBookshelfGroups(w http.ResponseWriter, r *http.Request) {
	items, err := h.bookshelfQueries.ListGroups.Handle(r.Context(), mustActor(r))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) CreateMyBookshelfGroup(w http.ResponseWriter, r *http.Request) {
	var req bookshelfapp.CreateGroup
	if !decodeJSON(w, r, &req) {
		return
	}
	item, err := h.bookshelfCommands.CreateGroup.Handle(r.Context(), mustActor(r), req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) UpdateMyBookshelfGroup(w http.ResponseWriter, r *http.Request) {
	groupID, ok := pathInt(w, r, "groupId")
	if !ok {
		return
	}
	var req struct {
		Name      string `json:"name"`
		SortOrder int    `json:"sortOrder"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.Name) != "" {
		item, err := h.bookshelfCommands.RenameGroup.Handle(r.Context(), mustActor(r), groupID, bookshelfapp.RenameGroup{Name: req.Name})
		if err != nil {
			writeError(w, err)
			return
		}
		if req.SortOrder != 0 {
			item, err = h.bookshelfCommands.ReorderGroup.Handle(r.Context(), mustActor(r), groupID, bookshelfapp.ReorderGroup{SortOrder: req.SortOrder})
			if err != nil {
				writeError(w, err)
				return
			}
		}
		writeJSON(w, http.StatusOK, item)
		return
	}
	item, err := h.bookshelfCommands.ReorderGroup.Handle(r.Context(), mustActor(r), groupID, bookshelfapp.ReorderGroup{SortOrder: req.SortOrder})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) DeleteMyBookshelfGroup(w http.ResponseWriter, r *http.Request) {
	groupID, ok := pathInt(w, r, "groupId")
	if !ok {
		return
	}
	if err := h.bookshelfCommands.DeleteGroup.Handle(r.Context(), mustActor(r), groupID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

func (h *Handler) ListMyBookshelf(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page := parseInt(query.Get("page"), 1)
	pageSize := parseInt(query.Get("pageSize"), 20)
	var groupID *int64
	if raw := strings.TrimSpace(query.Get("groupId")); raw != "" {
		if value, err := strconv.ParseInt(raw, 10, 64); err == nil && value > 0 {
			groupID = &value
		}
	}
	items, total, err := h.bookshelfQueries.ListEntries.Handle(r.Context(), mustActor(r), bookshelfapp.ListEntries{
		GroupID:  groupID,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	for i := range items {
		items[i].Book.CoverURL = "/api/books/" + strconv.FormatInt(items[i].Book.ID, 10) + "/cover"
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": total})
}

func (h *Handler) GetMyBookshelfBook(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathInt(w, r, "bookId")
	if !ok {
		return
	}
	item, err := h.bookshelfQueries.FindEntry.Handle(r.Context(), mustActor(r), bookID)
	if err != nil {
		writeError(w, err)
		return
	}
	item.Book.CoverURL = "/api/books/" + strconv.FormatInt(item.Book.ID, 10) + "/cover"
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) AddMyBookshelfBook(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathInt(w, r, "bookId")
	if !ok {
		return
	}
	var req bookshelfapp.AddBook
	if !decodeJSON(w, r, &req) {
		return
	}
	item, err := h.bookshelfCommands.AddBook.Handle(r.Context(), mustActor(r), bookID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	item.Book.CoverURL = "/api/books/" + strconv.FormatInt(item.Book.ID, 10) + "/cover"
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) UpdateMyBookshelfBook(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathInt(w, r, "bookId")
	if !ok {
		return
	}
	var req bookshelfapp.UpdateBook
	if !decodeJSON(w, r, &req) {
		return
	}
	item, err := h.bookshelfCommands.UpdateBook.Handle(r.Context(), mustActor(r), bookID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	item.Book.CoverURL = "/api/books/" + strconv.FormatInt(item.Book.ID, 10) + "/cover"
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) RemoveMyBookshelfBook(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathInt(w, r, "bookId")
	if !ok {
		return
	}
	if err := h.bookshelfCommands.RemoveBook.Handle(r.Context(), mustActor(r), bookID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

func (h *Handler) BatchManageMyBookshelf(w http.ResponseWriter, r *http.Request) {
	var req bookshelfapp.BatchManage
	if !decodeJSON(w, r, &req) {
		return
	}
	if err := h.bookshelfCommands.BatchManage.Handle(r.Context(), mustActor(r), req); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"updated": true})
}

func (h *Handler) writeDefaultAvatar(w http.ResponseWriter) {
	const avatarSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64" role="img" aria-label="avatar placeholder"><defs><linearGradient id="g" x1="0%" y1="0%" x2="100%" y2="100%"><stop offset="0%" stop-color="#f2a26b"/><stop offset="100%" stop-color="#d0672b"/></linearGradient></defs><rect width="64" height="64" rx="18" fill="url(#g)"/><circle cx="32" cy="25" r="11" fill="#fff" opacity="0.92"/><path d="M14 53c3.5-9.5 11-14 18-14s14.5 4.5 18 14" fill="#fff" opacity="0.92"/></svg>`
	w.Header().Set("Content-Type", "image/svg+xml")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(avatarSVG))
}
