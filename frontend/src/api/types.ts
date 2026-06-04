export type FrontRole = "reader" | "author";
export type AdminRole = "reviewer" | "super_admin";

export interface ApiErrorBody {
  code?: string;
  message?: string;
}

export interface CurrentUser {
  id: number;
  username: string;
  nickname: string;
  role: FrontRole;
  avatarUrl?: string;
}

export interface CurrentOperator {
  id: number;
  username: string;
  role: AdminRole;
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

export interface AdminLoginResponse {
  token: string;
  operator: CurrentOperator;
}

export interface Category {
  id: number;
  name: string;
}

export interface SiteSettings {
  id?: number;
  brandName: string;
  brandSubtitle: string;
  brandIconUrl: string;
  heroEyebrow: string;
  heroTitle: string;
  heroDescription: string;
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
  recommendScore?: number;
  coverUrl?: string;
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
  recommendScore?: number;
  coverUrl?: string;
}

export interface BookshelfGroup {
  id: number;
  userId: number;
  name: string;
  sortOrder: number;
  itemCount?: number;
  isPinned: boolean;
  pinnedAt?: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface BookshelfEntry {
  id: number;
  userId: number;
  bookId: number;
  groupId?: number;
  groupName: string;
  isPinned: boolean;
  pinnedAt?: string;
  createdAt?: string;
  updatedAt?: string;
  book: BookSummary;
}

export interface BookshelfGroupListResponse {
  items: BookshelfGroup[];
}

export interface BookshelfEntryListResponse {
  items: BookshelfEntry[];
  total: number;
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

export interface BookshelfEntryQueryParams {
  groupId?: number | string;
  page?: number;
  pageSize?: number;
}

export interface BookshelfGroupInput {
  name?: string;
  sortOrder?: number;
  pinned?: boolean;
}

export interface BookshelfBookInput {
  groupId?: number | string;
  pinned?: boolean;
}

export interface BookshelfBatchInput {
  action: "move" | "pin" | "unpin" | "remove";
  groupId?: number | string;
  bookIds: number[];
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
  recommendScore?: number;
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

export interface SiteSettingsInput {
  brandName: string;
  brandSubtitle: string;
  heroEyebrow: string;
  heroTitle: string;
  heroDescription: string;
}

export interface AuthorApplication {
  id: number;
  userId: number;
  username?: string;
  nickname?: string;
  penName: string;
  reason: string;
  status: "pending" | "approved" | "rejected";
  reviewNote?: string;
  reviewedByOperatorId?: number;
  reviewedAt?: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface FrontUserSummary {
  id: number;
  username: string;
  nickname: string;
  role: FrontRole;
  latestApplicationId?: number;
  latestApplicationStatus?: string;
  latestApplicationCreatedAt?: string;
  createdAt?: string;
  updatedAt?: string;
}
