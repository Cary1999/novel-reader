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
	siteapp "novel-reader/backend/internal/application/site"
	sitecommand "novel-reader/backend/internal/application/site/command"
	uploadapp "novel-reader/backend/internal/application/upload"
	bookentity "novel-reader/backend/internal/domain/book/entity"
	categoryentity "novel-reader/backend/internal/domain/category/entity"
	identityentity "novel-reader/backend/internal/domain/identity/entity"
	"novel-reader/backend/internal/domain/shared"
	siteentity "novel-reader/backend/internal/domain/site/entity"
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

func TestChapterDetailAllowsAuthenticatedReader(t *testing.T) {
	handler, tokens := testHandler(t)
	token, err := tokens.Issue(identityentity.User{ID: 7, Username: "reader", Role: shared.RoleReader})
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
}

func TestMyBooksRejectsReader(t *testing.T) {
	handler, tokens := testHandler(t)
	token, err := tokens.Issue(identityentity.User{ID: 7, Username: "reader", Role: shared.RoleReader})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/me/books", nil)
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
	token, err := tokens.Issue(identityentity.User{ID: 1, Username: "author", Role: shared.RoleAuthor})
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

func TestRecommendScoreUpdateRejectsAuthor(t *testing.T) {
	handler, tokens := testHandler(t)
	token, err := tokens.Issue(identityentity.User{ID: 1, Username: "author", Role: shared.RoleAuthor})
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

func TestAdminSiteSettingsRejectsFrontToken(t *testing.T) {
	handler, tokens := testHandler(t)
	token, err := tokens.Issue(identityentity.User{ID: 7, Username: "reader", Role: shared.RoleReader})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPatch, "/api/admin/site-settings", strings.NewReader(`{"brandName":"新站名"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertErrorCode(t, rec.Body.String(), "FORBIDDEN")
}

func TestReviewerCanUpdateRecommendScore(t *testing.T) {
	handler, tokens := testHandler(t)
	token, err := tokens.IssueOperator(identityentity.Operator{ID: 99, Username: "reviewer", Role: shared.RoleReviewer})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPatch, "/api/admin/books/1/recommend-score", strings.NewReader(`{"recommendScore":18}`))
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
	if item.RecommendScore != 18 {
		t.Fatalf("RecommendScore = %d, want 18", item.RecommendScore)
	}
}

func TestAdminBooksRejectsFrontToken(t *testing.T) {
	handler, tokens := testHandler(t)
	token, err := tokens.Issue(identityentity.User{ID: 1, Username: "author", Role: shared.RoleAuthor})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/admin/books?page=1&pageSize=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertErrorCode(t, rec.Body.String(), "FORBIDDEN")
}

func TestReviewerCanPromoteReaderToAuthor(t *testing.T) {
	handler, tokens := testHandler(t)
	token, err := tokens.IssueOperator(identityentity.Operator{ID: 99, Username: "reviewer", Role: shared.RoleReviewer})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPatch, "/api/admin/users/7/author-role", strings.NewReader(`{"action":"promote_to_author"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var user map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &user); err != nil {
		t.Fatal(err)
	}
	if user["role"] != string(shared.RoleAuthor) {
		t.Fatalf("role = %v, want %q", user["role"], shared.RoleAuthor)
	}
}

func TestOperatorsRequiresSuperAdmin(t *testing.T) {
	handler, tokens := testHandler(t)
	token, err := tokens.IssueOperator(identityentity.Operator{ID: 99, Username: "reviewer", Role: shared.RoleReviewer})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/admin/operators", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertErrorCode(t, rec.Body.String(), "FORBIDDEN")
}

func TestSuperAdminCanCreateOperator(t *testing.T) {
	handler, tokens := testHandler(t)
	token, err := tokens.IssueOperator(identityentity.Operator{ID: 100, Username: "root", Role: shared.RoleSuperAdmin})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/admin/operators", strings.NewReader(`{"username":"auditor","password":"password123","role":"reviewer"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var operator map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &operator); err != nil {
		t.Fatal(err)
	}
	if operator["role"] != string(shared.RoleReviewer) {
		t.Fatalf("role = %v, want %q", operator["role"], shared.RoleReviewer)
	}
}

func testHandler(t *testing.T) (http.Handler, *jwt.Manager) {
	t.Helper()
	tokens := jwt.NewManager([]byte("test-secret"), time.Hour)
	repo := fakeIdentityRepo{}
	identityQueries := identityapp.NewQueries(repo, repo, repo)
	identityCommands := identityapp.NewCommands(repo, repo, repo, tokens, tokens)
	bookQueries := bookapp.NewQueries(fakeBookRepo{})
	bookCommands := bookapp.NewCommands(fakeBookRepo{}, fakeCategoryRepo{})
	categoryQueries := categoryapp.NewQueries(fakeCategoryRepo{})
	categoryCommands := categoryapp.NewCommands(fakeCategoryRepo{})
	siteQueries := siteapp.NewQueries(fakeSiteRepo{})
	siteCommands := siteapp.NewCommands(fakeSiteRepo{}, fakeFileStore{})
	uploadSvc := uploadapp.NewService(fakeBookRepo{}, fakeCategoryRepo{}, fakeUploadRepo{}, fakeFileStore{}, fakeFileStore{}, fakeParser{})
	h := New(identityQueries, identityCommands, bookQueries, bookCommands, categoryQueries, categoryCommands, siteQueries, siteCommands, uploadSvc, tokens, 1024, 1024, t.TempDir(), t.TempDir(), "")
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

func (fakeIdentityRepo) FindUserByUsername(_ context.Context, username string) (identityentity.User, error) {
	switch username {
	case "author":
		return identityentity.User{ID: 1, Username: "author", Nickname: "author", PasswordHash: "$2a$10$7EqJtq98hPqEX7fNZaFWoO5NL0E7g8iIrM3P6KDAdAm8YGtSNYGG6", Role: shared.RoleAuthor}, nil
	default:
		return identityentity.User{}, shared.ErrNotFound
	}
}

func (fakeIdentityRepo) FindUserByID(_ context.Context, id int64) (identityentity.User, error) {
	role := shared.RoleReader
	if id == 1 {
		role = shared.RoleAuthor
	}
	return identityentity.User{ID: id, Username: "reader", Nickname: "reader", Role: role}, nil
}

func (fakeIdentityRepo) ListUsers(context.Context) ([]identityentity.FrontUserSummary, error) {
	return []identityentity.FrontUserSummary{{ID: 7, Username: "reader", Nickname: "reader", Role: string(shared.RoleReader)}}, nil
}

func (fakeIdentityRepo) UpdateUserNickname(_ context.Context, id int64, nickname string) (identityentity.User, error) {
	return identityentity.User{ID: id, Username: "reader", Nickname: nickname, Role: shared.RoleReader}, nil
}

func (fakeIdentityRepo) UpdateUserPassword(context.Context, int64, string) error {
	return nil
}

func (fakeIdentityRepo) PromoteUserToAuthor(_ context.Context, id int64) (identityentity.User, error) {
	return identityentity.User{ID: id, Username: "reader", Nickname: "reader", Role: shared.RoleAuthor}, nil
}

func (fakeIdentityRepo) CreateOperator(_ context.Context, username, passwordHash string, role shared.Role) (identityentity.Operator, error) {
	return identityentity.Operator{ID: 11, Username: username, PasswordHash: passwordHash, Role: role}, nil
}

func (fakeIdentityRepo) FindOperatorByUsername(_ context.Context, username string) (identityentity.Operator, error) {
	switch username {
	case "reviewer":
		return identityentity.Operator{ID: 99, Username: "reviewer", PasswordHash: "$2a$10$7EqJtq98hPqEX7fNZaFWoO5NL0E7g8iIrM3P6KDAdAm8YGtSNYGG6", Role: shared.RoleReviewer}, nil
	case "root":
		return identityentity.Operator{ID: 100, Username: "root", PasswordHash: "$2a$10$7EqJtq98hPqEX7fNZaFWoO5NL0E7g8iIrM3P6KDAdAm8YGtSNYGG6", Role: shared.RoleSuperAdmin}, nil
	default:
		return identityentity.Operator{}, shared.ErrNotFound
	}
}

func (fakeIdentityRepo) FindOperatorByID(_ context.Context, id int64) (identityentity.Operator, error) {
	role := shared.RoleReviewer
	if id == 100 {
		role = shared.RoleSuperAdmin
	}
	return identityentity.Operator{ID: id, Username: "operator", Role: role}, nil
}

func (fakeIdentityRepo) ListOperators(context.Context) ([]identityentity.Operator, error) {
	return []identityentity.Operator{{ID: 99, Username: "reviewer", Role: shared.RoleReviewer}}, nil
}

func (fakeIdentityRepo) UpdateOperatorRole(_ context.Context, id int64, role shared.Role) (identityentity.Operator, error) {
	return identityentity.Operator{ID: id, Username: "operator", Role: role}, nil
}

func (fakeIdentityRepo) UpdateOperatorPassword(context.Context, int64, string) error {
	return nil
}

func (fakeIdentityRepo) CreateAuthorApplication(_ context.Context, item identityentity.AuthorApplication) (identityentity.AuthorApplication, error) {
	item.ID = 1
	item.Username = "reader"
	item.Nickname = "reader"
	return item, nil
}

func (fakeIdentityRepo) FindLatestAuthorApplicationByUserID(context.Context, int64) (identityentity.AuthorApplication, error) {
	return identityentity.AuthorApplication{}, shared.ErrNotFound
}

func (fakeIdentityRepo) ListAuthorApplications(context.Context) ([]identityentity.AuthorApplication, error) {
	return []identityentity.AuthorApplication{{ID: 1, UserID: 7, Username: "reader", Nickname: "reader", PenName: "青石", Reason: "希望开始连载自己的小说", Status: "pending"}}, nil
}

func (fakeIdentityRepo) ReviewAuthorApplication(_ context.Context, applicationID int64, status, reviewNote string, reviewedByOperatorID int64) (identityentity.AuthorApplication, error) {
	return identityentity.AuthorApplication{ID: applicationID, UserID: 7, Username: "reader", Nickname: "reader", PenName: "青石", Reason: "希望开始连载自己的小说", Status: status, ReviewNote: reviewNote, ReviewedByOperatorID: &reviewedByOperatorID}, nil
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

func (fakeBookRepo) UpdateRecommendScore(_ context.Context, bookID int64, recommendScore int) (bookentity.Book, error) {
	owner := int64(1)
	return bookentity.Book{ID: bookID, Title: "测试书", Author: "作者", OwnerUserID: &owner, Category: "玄幻", RecommendScore: recommendScore}, nil
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

func (fakeFileStore) SaveSiteIcon(string, io.Reader) (sitecommand.SavedIcon, error) {
	return sitecommand.SavedIcon{RelativePath: "site-icon.svg"}, nil
}

type fakeParser struct{}

func (fakeParser) ParseChapters(string) ([]bookentity.ChapterDraft, error) {
	return []bookentity.ChapterDraft{{Index: 1, Title: "第一章", Content: "内容"}}, nil
}

type fakeSiteRepo struct{}

func (fakeSiteRepo) GetSettings(context.Context) (siteentity.Settings, error) {
	return siteentity.Settings{
		ID:              1,
		BrandName:       "阅卷书屋",
		BrandSubtitle:   "Local Reading Archive",
		HeroEyebrow:     "发现好故事",
		HeroTitle:       "一站式书屋",
		HeroDescription: "描述",
	}, nil
}

func (fakeSiteRepo) UpsertSettings(_ context.Context, item siteentity.Settings) (siteentity.Settings, error) {
	return item, nil
}
