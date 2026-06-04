import type {
  AdminLoginResponse,
  AuthorApplication,
  BookDetail,
  BookshelfBatchInput,
  BookshelfBookInput,
  BookshelfEntry,
  BookshelfEntryListResponse,
  BookshelfEntryQueryParams,
  BookshelfGroup,
  BookshelfGroupInput,
  BookshelfGroupListResponse,
  BookMetadataInput,
  Category,
  ChapterDetail,
  ChapterInput,
  ChapterSummary,
  CreateBookInput,
  CurrentOperator,
  CurrentUser,
  FrontUserSummary,
  LoginResponse,
  PagedBooks,
  RegisterResponse,
  SearchBooksParams,
  SiteSettings,
  SiteSettingsInput,
  UploadBookInput,
  UploadSummary,
} from "./types";

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? "";

export class ApiError extends Error {
  code: string;
  status: number;

  constructor(message: string, code = "UNKNOWN", status = 0) {
    super(message);
    this.name = "ApiError";
    this.code = code;
    this.status = status;
  }
}

function createTokenStore(key: string) {
  return {
    get(): string | null {
      try {
        return globalThis.localStorage?.getItem(key) ?? null;
      } catch {
        return null;
      }
    },
    set(token: string): void {
      globalThis.localStorage?.setItem(key, token);
    },
    clear(): void {
      globalThis.localStorage?.removeItem(key);
    },
  };
}

export const frontTokenStore = createTokenStore("novel_reader_front_token");
export const adminTokenStore = createTokenStore("novel_reader_admin_token");

function buildUrl(path: string, query?: Record<string, string | number | undefined>) {
  const params = new URLSearchParams();
  Object.entries(query ?? {}).forEach(([key, value]) => {
    if (value !== undefined && value !== "") {
      params.set(key, String(value));
    }
  });
  return `${API_BASE}${path}${params.size ? `?${params}` : ""}`;
}

async function parseJson<T>(response: Response): Promise<T> {
  const text = await response.text();
  const body = text ? JSON.parse(text) : {};

  if (!response.ok) {
    throw new ApiError(
      body.message ?? "请求失败，请稍后再试",
      body.code ?? response.statusText,
      response.status,
    );
  }

  return body as T;
}

async function request<T>(
  path: string,
  tokenStore: ReturnType<typeof createTokenStore>,
  options: RequestInit = {},
  query?: Record<string, string | number | undefined>,
): Promise<T> {
  const headers = new Headers(options.headers);
  const token = tokenStore.get();

  if (!(options.body instanceof FormData)) {
    headers.set("Content-Type", "application/json");
  }
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const response = await fetch(buildUrl(path, query), {
    ...options,
    headers,
  });

  return parseJson<T>(response);
}

function frontRequest<T>(path: string, options: RequestInit = {}, query?: Record<string, string | number | undefined>) {
  return request<T>(path, frontTokenStore, options, query);
}

function adminRequest<T>(path: string, options: RequestInit = {}, query?: Record<string, string | number | undefined>) {
  return request<T>(path, adminTokenStore, options, query);
}

function bookJson<T extends { categoryId: string | number }>(input: T) {
  return {
    ...input,
    categoryId: Number(input.categoryId),
  };
}

export const apiClient = {
  register(username: string, password: string) {
    return frontRequest<RegisterResponse>("/api/auth/register", {
      method: "POST",
      body: JSON.stringify({ username, password }),
    });
  },

  login(username: string, password: string) {
    return frontRequest<LoginResponse>("/api/auth/login", {
      method: "POST",
      body: JSON.stringify({ username, password }),
    });
  },

  adminLogin(username: string, password: string) {
    return adminRequest<AdminLoginResponse>("/api/admin/auth/login", {
      method: "POST",
      body: JSON.stringify({ username, password }),
    });
  },

  me() {
    return frontRequest<CurrentUser>("/api/auth/me");
  },

  adminMe() {
    return adminRequest<CurrentOperator>("/api/admin/auth/me");
  },

  updateMe(nickname: string) {
    return frontRequest<CurrentUser>("/api/auth/me", {
      method: "PATCH",
      body: JSON.stringify({ nickname }),
    });
  },

  uploadAvatar(file: File) {
    const formData = new FormData();
    formData.set("file", file);
    return frontRequest<CurrentUser>("/api/auth/me/avatar", {
      method: "POST",
      body: formData,
    });
  },

  changePassword(oldPassword: string, newPassword: string) {
    return frontRequest<{ changed: boolean }>("/api/auth/password", {
      method: "PATCH",
      body: JSON.stringify({ oldPassword, newPassword }),
    });
  },

  submitAuthorApplication(penName: string, reason: string) {
    return frontRequest<AuthorApplication>("/api/author-applications", {
      method: "POST",
      body: JSON.stringify({ penName, reason }),
    });
  },

  myAuthorApplication() {
    return frontRequest<AuthorApplication>("/api/author-applications/me");
  },

  categories() {
    return frontRequest<{ items: Category[] }>("/api/categories");
  },

  siteSettings() {
    return frontRequest<SiteSettings>("/api/site-settings", {
      cache: "no-store",
    });
  },

  updateSiteSettings(input: SiteSettingsInput) {
    return adminRequest<SiteSettings>("/api/admin/site-settings", {
      method: "PATCH",
      body: JSON.stringify(input),
    });
  },

  uploadSiteIcon(file: File) {
    const formData = new FormData();
    formData.set("file", file);
    return adminRequest<SiteSettings>("/api/admin/site-settings/icon", {
      method: "POST",
      body: formData,
    });
  },

  searchBooks(params: SearchBooksParams = {}) {
    return frontRequest<PagedBooks>("/api/books/search", {}, {
      q: params.q,
      category: params.category,
      page: params.page,
      pageSize: params.pageSize,
    });
  },

  recommendations(params: { page?: number; pageSize?: number } = {}) {
    return frontRequest<PagedBooks>("/api/books/recommendations", {}, {
      page: params.page,
      pageSize: params.pageSize,
    });
  },

  book(bookId: string | number) {
    return frontRequest<BookDetail>(`/api/books/${bookId}`);
  },

  chapters(bookId: string | number) {
    return frontRequest<{ items: ChapterSummary[] }>(`/api/books/${bookId}/chapters`);
  },

  chapter(bookId: string | number, chapterId: string | number) {
    return frontRequest<ChapterDetail>(`/api/books/${bookId}/chapters/${chapterId}`);
  },

  bookshelfGroups() {
    return frontRequest<BookshelfGroupListResponse>("/api/me/bookshelf/groups");
  },

  bookshelfGroup(groupId: string | number) {
    return frontRequest<BookshelfGroup>(`/api/me/bookshelf/groups/${groupId}`);
  },

  createBookshelfGroup(input: BookshelfGroupInput) {
    return frontRequest<BookshelfGroup>("/api/me/bookshelf/groups", {
      method: "POST",
      body: JSON.stringify(input),
    });
  },

  updateBookshelfGroup(groupId: string | number, input: BookshelfGroupInput) {
    return frontRequest<BookshelfGroup>(`/api/me/bookshelf/groups/${groupId}`, {
      method: "PATCH",
      body: JSON.stringify(input),
    });
  },

  deleteBookshelfGroup(groupId: string | number) {
    return frontRequest<{ deleted: boolean }>(`/api/me/bookshelf/groups/${groupId}`, {
      method: "DELETE",
    });
  },

  bookshelf(params: BookshelfEntryQueryParams = {}) {
    return frontRequest<BookshelfEntryListResponse>("/api/me/bookshelf", {}, {
      groupId: params.groupId,
      page: params.page,
      pageSize: params.pageSize,
    });
  },

  bookshelfEntry(bookId: string | number) {
    return frontRequest<BookshelfEntry>(`/api/me/bookshelf/${bookId}`);
  },

  addToBookshelf(bookId: string | number, input: BookshelfBookInput = {}) {
    return frontRequest<BookshelfEntry>(`/api/me/bookshelf/${bookId}`, {
      method: "POST",
      body: JSON.stringify(input),
    });
  },

  updateBookshelfBook(bookId: string | number, input: BookshelfBookInput = {}) {
    return frontRequest<BookshelfEntry>(`/api/me/bookshelf/${bookId}`, {
      method: "PATCH",
      body: JSON.stringify(input),
    });
  },

  removeFromBookshelf(bookId: string | number) {
    return frontRequest<{ deleted: boolean }>(`/api/me/bookshelf/${bookId}`, {
      method: "DELETE",
    });
  },

  batchManageBookshelf(input: BookshelfBatchInput) {
    return frontRequest<{ updated: boolean }>("/api/me/bookshelf/batch", {
      method: "POST",
      body: JSON.stringify(input),
    });
  },

  uploadBook(input: UploadBookInput) {
    const formData = new FormData();
    formData.set("title", input.title);
    formData.set("categoryId", String(input.categoryId));
    formData.set("description", input.description ?? "");
    formData.set("file", input.file);

    return frontRequest<UploadSummary>("/api/me/books/upload", {
      method: "POST",
      body: formData,
    });
  },

  myBooks(params: SearchBooksParams = {}) {
    return frontRequest<PagedBooks>("/api/me/books", {}, {
      q: params.q,
      categoryId: params.categoryId,
      page: params.page,
      pageSize: params.pageSize,
    });
  },

  createBook(input: CreateBookInput) {
    return frontRequest<BookDetail>("/api/me/books", {
      method: "POST",
      body: JSON.stringify(bookJson(input)),
    });
  },

  updateBook(bookId: string | number, input: BookMetadataInput) {
    return frontRequest<BookDetail>(`/api/books/${bookId}`, {
      method: "PATCH",
      body: JSON.stringify(bookJson(input)),
    });
  },

  uploadBookCover(bookId: string | number, file: File) {
    const formData = new FormData();
    formData.set("file", file);
    return frontRequest<{ coverUrl: string }>(`/api/books/${bookId}/cover`, {
      method: "POST",
      body: formData,
    });
  },

  deleteBook(bookId: string | number) {
    return frontRequest<{ deleted: boolean }>(`/api/books/${bookId}`, {
      method: "DELETE",
    });
  },

  addChapter(bookId: string | number, input: ChapterInput) {
    return frontRequest<ChapterDetail>(`/api/books/${bookId}/chapters`, {
      method: "POST",
      body: JSON.stringify(input),
    });
  },

  updateChapter(bookId: string | number, chapterId: string | number, input: ChapterInput) {
    return frontRequest<ChapterDetail>(`/api/books/${bookId}/chapters/${chapterId}`, {
      method: "PATCH",
      body: JSON.stringify(input),
    });
  },

  deleteChapter(bookId: string | number, chapterId: string | number) {
    return frontRequest<{ deleted: boolean }>(`/api/books/${bookId}/chapters/${chapterId}`, {
      method: "DELETE",
    });
  },

  listAuthorApplications() {
    return adminRequest<{ items: AuthorApplication[] }>("/api/admin/author-applications");
  },

  adminBooks(params: SearchBooksParams = {}) {
    return adminRequest<PagedBooks>("/api/admin/books", {}, {
      q: params.q,
      category: params.category,
      page: params.page,
      pageSize: params.pageSize,
    });
  },

  updateRecommendScore(bookId: string | number, recommendScore: number) {
    return adminRequest<BookDetail>(`/api/admin/books/${bookId}/recommend-score`, {
      method: "PATCH",
      body: JSON.stringify({ recommendScore }),
    });
  },

  reviewAuthorApplication(applicationId: string | number, decision: "approved" | "rejected", reviewNote = "") {
    return adminRequest<AuthorApplication>(`/api/admin/author-applications/${applicationId}`, {
      method: "PATCH",
      body: JSON.stringify({ decision, reviewNote }),
    });
  },

  listUsers() {
    return adminRequest<{ items: FrontUserSummary[] }>("/api/admin/users");
  },

  promoteUserToAuthor(userId: string | number) {
    return adminRequest<CurrentUser>(`/api/admin/users/${userId}/author-role`, {
      method: "PATCH",
      body: JSON.stringify({ action: "promote_to_author" }),
    });
  },

  listOperators() {
    return adminRequest<{ items: CurrentOperator[] }>("/api/admin/operators");
  },

  createOperator(username: string, password: string, role: "reviewer" | "super_admin") {
    return adminRequest<CurrentOperator>("/api/admin/operators", {
      method: "POST",
      body: JSON.stringify({ username, password, role }),
    });
  },

  updateOperator(operatorId: string | number, role: "reviewer" | "super_admin") {
    return adminRequest<CurrentOperator>(`/api/admin/operators/${operatorId}`, {
      method: "PATCH",
      body: JSON.stringify({ role }),
    });
  },

  createCategory(name: string) {
    return adminRequest<Category>("/api/admin/categories", {
      method: "POST",
      body: JSON.stringify({ name }),
    });
  },

  updateCategory(categoryId: string | number, name: string) {
    return adminRequest<Category>(`/api/admin/categories/${categoryId}`, {
      method: "PATCH",
      body: JSON.stringify({ name }),
    });
  },

  deleteCategory(categoryId: string | number) {
    return adminRequest<{ deleted: boolean }>(`/api/admin/categories/${categoryId}`, {
      method: "DELETE",
    });
  },
};
