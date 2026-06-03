package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	bookapp "novel-reader/backend/internal/application/book"
	categoryapp "novel-reader/backend/internal/application/category"
	identityapp "novel-reader/backend/internal/application/identity"
	uploadapp "novel-reader/backend/internal/application/upload"
	bookentity "novel-reader/backend/internal/domain/book/entity"
	categoryentity "novel-reader/backend/internal/domain/category/entity"
	identityentity "novel-reader/backend/internal/domain/identity/entity"
	"novel-reader/backend/internal/domain/shared"
	uploadentity "novel-reader/backend/internal/domain/upload/entity"
	"novel-reader/backend/internal/infrastructure/auth/jwt"
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
	token, err := tokens.Issue(identityentity.User{ID: 7, Username: "reader", Role: shared.RoleUser})
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
	var chapterItem bookentity.Chapter
	if err := json.Unmarshal(rec.Body.Bytes(), &chapterItem); err != nil {
		t.Fatal(err)
	}
	if chapterItem.ID != 2 || chapterItem.BookID != 1 || chapterItem.Content == "" {
		t.Fatalf("unexpected chapter response: %#v", chapterItem)
	}
}

func TestAdminUploadRejectsNonAdminToken(t *testing.T) {
	handler, tokens := testHandler(t)
	token, err := tokens.Issue(identityentity.User{ID: 7, Username: "reader", Role: shared.RoleUser})
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
	token, err := tokens.Issue(identityentity.User{ID: 7, Username: "reader", Role: shared.RoleUser})
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
	token, err := tokens.Issue(identityentity.User{ID: 1, Username: "admin", Role: shared.RoleAdmin})
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
	var chapterItem bookentity.Chapter
	if err := json.Unmarshal(rec.Body.Bytes(), &chapterItem); err != nil {
		t.Fatal(err)
	}
	if chapterItem.BookID != 1 || chapterItem.Title != "新章" || chapterItem.Content != "正文" {
		t.Fatalf("unexpected chapter response: %#v", chapterItem)
	}
}

func TestAdminUploadParsesOnlyForAdmin(t *testing.T) {
	handler, tokens := testHandler(t)
	token, err := tokens.Issue(identityentity.User{ID: 1, Username: "admin", Role: shared.RoleAdmin})
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

func TestRecommendScoreUpdateRejectsAuthor(t *testing.T) {
	handler, tokens := testHandler(t)
	token, err := tokens.Issue(identityentity.User{ID: 1, Username: "reader", Role: shared.RoleUser})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPatch, "/api/books/1", strings.NewReader(`{"title":"书名","categoryId":1,"description":"简介","recommendScore":9}`))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertErrorCode(t, rec.Body.String(), "FORBIDDEN")
}

func TestBookUpdatePreservesRecommendScoreWhenAuthorEditsMetadata(t *testing.T) {
	handler, tokens := testHandler(t)
	token, err := tokens.Issue(identityentity.User{ID: 1, Username: "reader", Role: shared.RoleUser})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPatch, "/api/books/1", strings.NewReader(`{"title":"新书名","categoryId":1,"description":"新简介"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var item bookentity.Book
	if err := json.Unmarshal(rec.Body.Bytes(), &item); err != nil {
		t.Fatal(err)
	}
	if item.RecommendScore != 7 {
		t.Fatalf("RecommendScore = %d, want 7", item.RecommendScore)
	}
}

func testHandler(t *testing.T) (http.Handler, *jwt.Manager) {
	t.Helper()
	tokens := jwt.NewManager([]byte("test-secret"), time.Hour)
	identityQueries := identityapp.NewQueries(fakeIdentityRepo{})
	identityCommands := identityapp.NewCommands(fakeIdentityRepo{}, tokens)
	bookQueries := bookapp.NewQueries(fakeBookRepo{})
	bookCommands := bookapp.NewCommands(fakeBookRepo{}, fakeCategoryRepo{})
	categoryQueries := categoryapp.NewQueries(fakeCategoryRepo{})
	categoryCommands := categoryapp.NewCommands(fakeCategoryRepo{})
	uploadSvc := uploadapp.NewService(fakeBookRepo{}, fakeCategoryRepo{}, fakeUploadRepo{}, fakeFileStore{}, fakeFileStore{}, fakeParser{})
	h := New(identityQueries, identityCommands, bookQueries, bookCommands, categoryQueries, categoryCommands, uploadSvc, tokens, 1024, t.TempDir(), "")
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

type fakeIdentityRepo struct{}

func (fakeIdentityRepo) CreateUser(context.Context, string, string, shared.Role) (identityentity.User, error) {
	return identityentity.User{}, nil
}

func (fakeIdentityRepo) FindUserByUsername(context.Context, string) (identityentity.User, error) {
	return identityentity.User{}, shared.ErrNotFound
}

func (fakeIdentityRepo) FindUserByID(_ context.Context, id int64) (identityentity.User, error) {
	return identityentity.User{ID: id, Username: "reader", Nickname: "reader", Role: shared.RoleUser}, nil
}

func (fakeIdentityRepo) UpdateUserNickname(_ context.Context, id int64, nickname string) (identityentity.User, error) {
	return identityentity.User{ID: id, Username: "reader", Nickname: nickname, Role: shared.RoleUser}, nil
}

func (fakeIdentityRepo) UpdateUserPassword(context.Context, int64, string) error {
	return nil
}

type fakeBookRepo struct{}

func (fakeBookRepo) SearchBooks(context.Context, string, string, int, int) ([]bookentity.Book, int, error) {
	return nil, 0, nil
}

func (fakeBookRepo) ListRecommendedBooks(context.Context, int, int) ([]bookentity.Book, int, error) {
	return nil, 0, nil
}

func (fakeBookRepo) ListBooksByOwner(context.Context, int64, string, int64, int, int) ([]bookentity.Book, int, error) {
	return nil, 0, nil
}

func (fakeBookRepo) FindBook(_ context.Context, id int64) (bookentity.Book, error) {
	owner := int64(1)
	return bookentity.Book{ID: id, OwnerUserID: &owner, RecommendScore: 7}, nil
}

func (fakeBookRepo) UpdateBookCoverPath(context.Context, int64, *string) error {
	return nil
}

func (fakeBookRepo) ListChapters(context.Context, int64) ([]bookentity.Chapter, error) {
	return nil, nil
}

func (fakeBookRepo) FindChapter(_ context.Context, bookID, chapterID int64) (bookentity.Chapter, error) {
	return bookentity.Chapter{ID: chapterID, BookID: bookID, Index: 1, Title: "Chapter 1", Content: "content"}, nil
}

func (fakeBookRepo) CreateBookWithChapters(context.Context, bookentity.Book, *int64, int64, []bookentity.ChapterDraft) (int64, error) {
	return 1, nil
}

func (fakeBookRepo) CreateBook(_ context.Context, input bookentity.Book, _ *int64) (bookentity.Book, error) {
	return bookentity.Book{ID: 10, Title: input.Title, Author: "作者", OwnerUserID: input.OwnerUserID, Description: input.Description}, nil
}

func (fakeBookRepo) UpdateBookMetadata(_ context.Context, bookID int64, input bookentity.Book, _ *int64) (bookentity.Book, error) {
	return bookentity.Book{ID: bookID, Title: input.Title, Author: "作者", Description: input.Description, RecommendScore: input.RecommendScore}, nil
}

func (fakeBookRepo) DeleteBook(context.Context, int64) error {
	return nil
}

func (fakeBookRepo) AddChapter(_ context.Context, bookID int64, input bookentity.Chapter) (bookentity.Chapter, error) {
	return bookentity.Chapter{ID: 11, BookID: bookID, Index: 3, Title: input.Title, Content: input.Content}, nil
}

func (fakeBookRepo) UpdateChapter(_ context.Context, bookID, chapterID int64, input bookentity.Chapter) (bookentity.Chapter, error) {
	return bookentity.Chapter{ID: chapterID, BookID: bookID, Index: 1, Title: input.Title, Content: input.Content}, nil
}

func (fakeBookRepo) DeleteChapter(context.Context, int64, int64) error {
	return nil
}

type fakeCategoryRepo struct{}

func (fakeCategoryRepo) ListCategories(context.Context) ([]categoryentity.Category, error) {
	return nil, nil
}

func (fakeCategoryRepo) FindCategoryByID(context.Context, int64) (categoryentity.Category, error) {
	return categoryentity.Category{ID: 1, Name: "分类"}, nil
}

func (fakeCategoryRepo) CreateCategory(_ context.Context, name string) (categoryentity.Category, error) {
	return categoryentity.Category{ID: 1, Name: name}, nil
}

func (fakeCategoryRepo) UpdateCategory(_ context.Context, id int64, name string) (categoryentity.Category, error) {
	return categoryentity.Category{ID: id, Name: name}, nil
}

func (fakeCategoryRepo) DeleteCategory(context.Context, int64) error {
	return nil
}

type fakeUploadRepo struct{}

func (fakeUploadRepo) CreateUpload(context.Context, uploadentity.Upload) (int64, error) {
	return 1, nil
}

func (fakeUploadRepo) MarkUpload(context.Context, int64, string, string) error {
	return nil
}

type fakeFileStore struct{}

func (fakeFileStore) SaveTXT(string, io.Reader) (uploadapp.SavedFile, error) {
	return uploadapp.SavedFile{}, nil
}

func (fakeFileStore) SaveCover(string, io.Reader) (uploadapp.SavedFile, error) {
	return uploadapp.SavedFile{RelativePath: "cover.png"}, nil
}

type fakeParser struct{}

func (fakeParser) ParseChapters(string) ([]bookentity.ChapterDraft, error) {
	return []bookentity.ChapterDraft{{Index: 1, Title: "第一章", Content: "内容"}}, nil
}
