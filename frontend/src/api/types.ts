export type Role = "user" | "admin";

export interface ApiErrorBody {
  code?: string;
  message?: string;
}

export interface CurrentUser {
  id: number;
  username: string;
  nickname: string;
  role: Role;
}

export interface RegisterResponse {
  userId: number;
  username: string;
  nickname: string;
}

export interface LoginResponse {
  token: string;
  user: CurrentUser;
}

export interface Category {
  id: number;
  name: string;
}

export interface BookSummary {
  id: number;
  title: string;
  author: string;
  categoryId?: number;
  category: string;
  description: string;
  chapterCount: number;
  latestChapterTitle?: string;
  createdAt?: string;
}

export interface BookDetail {
  id: number;
  title: string;
  author: string;
  categoryId?: number;
  category: string;
  description: string;
  chapterCount: number;
}

export interface ChapterSummary {
  id: number;
  index: number;
  title: string;
}

export interface ChapterDetail extends ChapterSummary {
  bookId: number;
  content: string;
}

export interface PagedBooks {
  items: BookSummary[];
  total: number;
}

export interface UploadSummary {
  bookId: number;
  chapterCount: number;
  uploadId: number;
  firstChapterTitle: string;
  lastChapterTitle: string;
}

export interface SearchBooksParams {
  q?: string;
  category?: string;
  categoryId?: number | string;
  page?: number;
  pageSize?: number;
}

export interface UploadBookInput {
  title: string;
  categoryId: number | string;
  description?: string;
  file: File;
}

export interface BookMetadataInput {
  title: string;
  categoryId: number | string;
  description?: string;
}

export interface CreateBookInput {
  title: string;
  categoryId: number | string;
  description?: string;
}

export interface ChapterInput {
  title: string;
  content: string;
}
