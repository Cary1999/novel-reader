package httphandler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"novel-reader/backend/internal/auth"
	"novel-reader/backend/internal/domain"
)

func TestChapterDetailRequiresLogin(t *testing.T) {
	handler, _ := testHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/books/1/chapters/2", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertErrorCode(t, rec.Body.String(), "UNAUTHORIZED")
}

func TestChapterDetailAllowsAuthenticatedUser(t *testing.T) {
	handler, tokens := testHandler(t)
	token, err := tokens.Issue(domain.User{ID: 7, Username: "reader", Role: domain.RoleUser})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/books/1/chapters/2", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var chapter domain.Chapter
	if err := json.Unmarshal(rec.Body.Bytes(), &chapter); err != nil {
		t.Fatal(err)
	}
	if chapter.ID != 2 || chapter.BookID != 1 || chapter.Content == "" {
		t.Fatalf("unexpected chapter response: %#v", chapter)
	}
}

func TestAdminUploadRejectsNonAdminToken(t *testing.T) {
	handler, tokens := testHandler(t)
	token, err := tokens.Issue(domain.User{ID: 7, Username: "reader", Role: domain.RoleUser})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/admin/books/upload", strings.NewReader(""))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertErrorCode(t, rec.Body.String(), "FORBIDDEN")
}

func TestAdminBookManagementRejectsNonAdminToken(t *testing.T) {
	handler, tokens := testHandler(t)
	token, err := tokens.Issue(domain.User{ID: 7, Username: "reader", Role: domain.RoleUser})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodDelete, "/api/admin/books/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertErrorCode(t, rec.Body.String(), "FORBIDDEN")
}

func TestAdminAddChapterAllowsAdmin(t *testing.T) {
	handler, tokens := testHandler(t)
	token, err := tokens.Issue(domain.User{ID: 1, Username: "admin", Role: domain.RoleAdmin})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/admin/books/1/chapters", strings.NewReader(`{"title":"新章","content":"正文"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var chapter domain.Chapter
	if err := json.Unmarshal(rec.Body.Bytes(), &chapter); err != nil {
		t.Fatal(err)
	}
	if chapter.BookID != 1 || chapter.Title != "新章" || chapter.Content != "正文" {
		t.Fatalf("unexpected chapter response: %#v", chapter)
	}
}

func TestAdminUploadParsesOnlyForAdmin(t *testing.T) {
	handler, tokens := testHandler(t)
	token, err := tokens.Issue(domain.User{ID: 1, Username: "admin", Role: domain.RoleAdmin})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/admin/books/upload", strings.NewReader(""))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected multipart validation after admin auth, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertErrorCode(t, rec.Body.String(), "BAD_REQUEST")
}

func testHandler(t *testing.T) (http.Handler, *auth.Manager) {
	t.Helper()
	tokens := auth.NewManager([]byte("test-secret"), time.Hour)
	h := New(fakeAuth{}, fakeBooks{}, fakeAdmin{}, tokens, 1024, t.TempDir(), "")
	return h.Routes(), tokens
}

func assertErrorCode(t *testing.T, body, want string) {
	t.Helper()
	var got struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatal(err)
	}
	if got.Code != want {
		t.Fatalf("expected code %q, got %q", want, got.Code)
	}
}

type fakeAuth struct{}

func (fakeAuth) Register(context.Context, string, string) (domain.User, error) {
	return domain.User{}, nil
}

func (fakeAuth) Login(context.Context, string, string) (string, domain.User, error) {
	return "", domain.User{}, nil
}

func (fakeAuth) CurrentUser(_ context.Context, claims auth.Claims) (domain.User, error) {
	return domain.User{ID: claims.UserID, Username: claims.Username, Nickname: claims.Username, Role: claims.Role}, nil
}

func (fakeAuth) UpdateNickname(_ context.Context, claims auth.Claims, nickname string) (domain.User, error) {
	return domain.User{ID: claims.UserID, Username: claims.Username, Nickname: nickname, Role: claims.Role}, nil
}

func (fakeAuth) ChangePassword(context.Context, auth.Claims, string, string) error {
	return nil
}

type fakeBooks struct{}

func (fakeBooks) ListCategories(context.Context) ([]domain.Category, error) {
	return nil, nil
}

func (fakeBooks) SearchBooks(context.Context, string, string, int, int) ([]domain.Book, int, error) {
	return nil, 0, nil
}

func (fakeBooks) ListRecommendedBooks(context.Context, int, int) ([]domain.Book, int, error) {
	return nil, 0, nil
}

func (fakeBooks) GetBook(context.Context, int64) (domain.Book, error) {
	return domain.Book{}, nil
}

func (fakeBooks) ListChapters(context.Context, int64) ([]domain.Chapter, error) {
	return nil, nil
}

func (fakeBooks) GetChapter(_ context.Context, bookID, chapterID int64) (domain.Chapter, error) {
	return domain.Chapter{ID: chapterID, BookID: bookID, Index: 1, Title: "Chapter 1", Content: "content"}, nil
}

type fakeAdmin struct{}

func (fakeAdmin) ListBooks(context.Context, string, string, int, int) ([]domain.Book, int, error) {
	return nil, 0, nil
}

func (fakeAdmin) ListMyBooks(context.Context, int64, string, int64, int, int) ([]domain.Book, int, error) {
	return nil, 0, nil
}

func (fakeAdmin) CreateBook(_ context.Context, ownerID int64, input domain.UploadBookInput) (domain.Book, error) {
	return domain.Book{ID: 10, Title: input.Title, Author: "作者", OwnerUserID: &ownerID, Description: input.Description}, nil
}

func (fakeAdmin) UploadBook(context.Context, int64, domain.UploadBookInput, string, io.Reader) (domain.UploadResult, error) {
	return domain.UploadResult{}, nil
}

func (fakeAdmin) UpdateBook(_ context.Context, _ int64, _ domain.Role, bookID int64, input domain.BookMetadataInput) (domain.Book, error) {
	return domain.Book{ID: bookID, Title: input.Title, Author: "作者", Description: input.Description}, nil
}

func (fakeAdmin) UploadCover(context.Context, int64, domain.Role, int64, string, io.Reader) (string, error) {
	return "/api/books/1/cover", nil
}

func (fakeAdmin) DeleteBook(context.Context, int64, domain.Role, int64) error {
	return nil
}

func (fakeAdmin) AddChapter(_ context.Context, _ int64, _ domain.Role, bookID int64, input domain.ChapterInput) (domain.Chapter, error) {
	return domain.Chapter{ID: 11, BookID: bookID, Index: 3, Title: input.Title, Content: input.Content}, nil
}

func (fakeAdmin) UpdateChapter(_ context.Context, _ int64, _ domain.Role, bookID, chapterID int64, input domain.ChapterInput) (domain.Chapter, error) {
	return domain.Chapter{ID: chapterID, BookID: bookID, Index: 1, Title: input.Title, Content: input.Content}, nil
}

func (fakeAdmin) DeleteChapter(context.Context, int64, domain.Role, int64, int64) error {
	return nil
}

func (fakeAdmin) CreateCategory(_ context.Context, name string) (domain.Category, error) {
	return domain.Category{ID: 1, Name: name}, nil
}

func (fakeAdmin) UpdateCategory(_ context.Context, id int64, name string) (domain.Category, error) {
	return domain.Category{ID: id, Name: name}, nil
}

func (fakeAdmin) DeleteCategory(context.Context, int64) error {
	return nil
}
