import type {
  BookDetail,
  CreateBookInput,
  BookMetadataInput,
  Category,
  ChapterDetail,
  ChapterInput,
  ChapterSummary,
  CurrentUser,
  LoginResponse,
  PagedBooks,
  RegisterResponse,
  SearchBooksParams,
  UploadBookInput,
  UploadSummary,
} from "./types";

const TOKEN_KEY = "novel_reader_token";
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

export const tokenStore = {
  get(): string | null {
    try {
      return globalThis.localStorage?.getItem(TOKEN_KEY) ?? null;
    } catch {
      return null;
    }
  },
  set(token: string): void {
    globalThis.localStorage?.setItem(TOKEN_KEY, token);
  },
  clear(): void {
    globalThis.localStorage?.removeItem(TOKEN_KEY);
  },
};

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

function bookJson<T extends { categoryId: string | number }>(input: T) {
  return {
    ...input,
    categoryId: Number(input.categoryId),
  };
}

export const apiClient = {
  register(username: string, password: string) {
    return request<RegisterResponse>("/api/auth/register", {
      method: "POST",
      body: JSON.stringify({ username, password }),
    });
  },

  login(username: string, password: string) {
    return request<LoginResponse>("/api/auth/login", {
      method: "POST",
      body: JSON.stringify({ username, password }),
    });
  },

  me() {
    return request<CurrentUser>("/api/auth/me");
  },

  updateMe(nickname: string) {
    return request<CurrentUser>("/api/auth/me", {
      method: "PATCH",
      body: JSON.stringify({ nickname }),
    });
  },

  changePassword(oldPassword: string, newPassword: string) {
    return request<{ changed: boolean }>("/api/auth/password", {
      method: "PATCH",
      body: JSON.stringify({ oldPassword, newPassword }),
    });
  },

  categories() {
    return request<{ items: Category[] }>("/api/categories");
  },

  searchBooks(params: SearchBooksParams = {}) {
    return request<PagedBooks>("/api/books/search", {}, {
      q: params.q,
      category: params.category,
      page: params.page,
      pageSize: params.pageSize,
    });
  },

  recommendations(params: { page?: number; pageSize?: number } = {}) {
    return request<PagedBooks>("/api/books/recommendations", {}, {
      page: params.page,
      pageSize: params.pageSize,
    });
  },

  book(bookId: string | number) {
    return request<BookDetail>(`/api/books/${bookId}`);
  },

  chapters(bookId: string | number) {
    return request<{ items: ChapterSummary[] }>(`/api/books/${bookId}/chapters`);
  },

  chapter(bookId: string | number, chapterId: string | number) {
    return request<ChapterDetail>(`/api/books/${bookId}/chapters/${chapterId}`);
  },

  uploadBook(input: UploadBookInput) {
    const formData = new FormData();
    formData.set("title", input.title);
    formData.set("categoryId", String(input.categoryId));
    formData.set("description", input.description ?? "");
    formData.set("file", input.file);

    return request<UploadSummary>("/api/me/books/upload", {
      method: "POST",
      body: formData,
    });
  },

  adminUploadBook(input: UploadBookInput) {
    const formData = new FormData();
    formData.set("title", input.title);
    formData.set("categoryId", String(input.categoryId));
    formData.set("description", input.description ?? "");
    formData.set("file", input.file);

    return request<UploadSummary>("/api/admin/books/upload", {
      method: "POST",
      body: formData,
    });
  },

  adminBooks(params: SearchBooksParams = {}) {
    return request<PagedBooks>("/api/admin/books", {}, {
      q: params.q,
      category: params.category,
      page: params.page,
      pageSize: params.pageSize,
    });
  },

  myBooks(params: SearchBooksParams = {}) {
    return request<PagedBooks>("/api/me/books", {}, {
      q: params.q,
      categoryId: params.categoryId,
      page: params.page,
      pageSize: params.pageSize,
    });
  },

  createBook(input: CreateBookInput, scope: "me" | "admin" = "me") {
    return request<BookDetail>(scope === "admin" ? "/api/admin/books" : "/api/me/books", {
      method: "POST",
      body: JSON.stringify(bookJson(input)),
    });
  },

  updateBook(bookId: string | number, input: BookMetadataInput) {
    return request<BookDetail>(`/api/books/${bookId}`, {
      method: "PATCH",
      body: JSON.stringify(bookJson(input)),
    });
  },

  uploadBookCover(bookId: string | number, file: File) {
    const formData = new FormData();
    formData.set("file", file);
    return request<{ coverUrl: string }>(`/api/books/${bookId}/cover`, {
      method: "POST",
      body: formData,
    });
  },

  deleteBook(bookId: string | number) {
    return request<{ deleted: boolean }>(`/api/books/${bookId}`, {
      method: "DELETE",
    });
  },

  addChapter(bookId: string | number, input: ChapterInput) {
    return request<ChapterDetail>(`/api/books/${bookId}/chapters`, {
      method: "POST",
      body: JSON.stringify(input),
    });
  },

  updateChapter(bookId: string | number, chapterId: string | number, input: ChapterInput) {
    return request<ChapterDetail>(`/api/books/${bookId}/chapters/${chapterId}`, {
      method: "PATCH",
      body: JSON.stringify(input),
    });
  },

  deleteChapter(bookId: string | number, chapterId: string | number) {
    return request<{ deleted: boolean }>(`/api/books/${bookId}/chapters/${chapterId}`, {
      method: "DELETE",
    });
  },

  createCategory(name: string) {
    return request<Category>("/api/admin/categories", {
      method: "POST",
      body: JSON.stringify({ name }),
    });
  },

  updateCategory(categoryId: string | number, name: string) {
    return request<Category>(`/api/admin/categories/${categoryId}`, {
      method: "PATCH",
      body: JSON.stringify({ name }),
    });
  },

  deleteCategory(categoryId: string | number) {
    return request<{ deleted: boolean }>(`/api/admin/categories/${categoryId}`, {
      method: "DELETE",
    });
  },
};
