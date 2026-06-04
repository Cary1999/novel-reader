import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { adminTokenStore, apiClient, ApiError, frontTokenStore } from "../api/client";

const storage = new Map<string, string>();

beforeEach(() => {
  storage.clear();
  vi.stubGlobal("localStorage", {
    getItem: vi.fn((key: string) => storage.get(key) ?? null),
    setItem: vi.fn((key: string, value: string) => storage.set(key, value)),
    removeItem: vi.fn((key: string) => storage.delete(key)),
  });
});

afterEach(() => {
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});

describe("apiClient", () => {
  it("adds bearer token to authenticated requests", async () => {
    frontTokenStore.set("abc123");
    const fetchMock = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) => (
      new Response(JSON.stringify({ id: 1, username: "reader", nickname: "reader", role: "reader" }))
    ));
    vi.stubGlobal("fetch", fetchMock);

    await apiClient.me();

    const [, options] = fetchMock.mock.calls[0];
    const headers = options?.headers as Headers;
    expect(headers.get("Authorization")).toBe("Bearer abc123");
  });

  it("turns backend error bodies into ApiError", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => new Response(
      JSON.stringify({ code: "UNAUTHORIZED", message: "login required" }),
      { status: 401 },
    )));

    await expect(apiClient.chapter(1, 2)).rejects.toMatchObject({
      code: "UNAUTHORIZED",
      message: "login required",
      status: 401,
    });
  });

  it("sends upload forms without forcing json content type", async () => {
    frontTokenStore.set("author-token");
    const file = new File(["第一章 开始\n正文"], "novel.txt", { type: "text/plain" });
    const fetchMock = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) => new Response(JSON.stringify({
      bookId: 1,
      chapterCount: 1,
      uploadId: 9,
      firstChapterTitle: "第一章 开始",
      lastChapterTitle: "第一章 开始",
    })));
    vi.stubGlobal("fetch", fetchMock);

    await apiClient.uploadBook({ title: "测试书", categoryId: 1, file });

    const [, options] = fetchMock.mock.calls[0];
    expect(options).toBeDefined();
    const headers = options!.headers as Headers;
    expect(options!.body).toBeInstanceOf(FormData);
    expect(headers.get("Content-Type")).toBeNull();
    expect(headers.get("Authorization")).toBe("Bearer author-token");
  });

  it("uses book management endpoints with front bearer token", async () => {
    frontTokenStore.set("author-token");
    const fetchMock = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) => new Response(JSON.stringify({ deleted: true })));
    vi.stubGlobal("fetch", fetchMock);

    await apiClient.deleteBook(42);

    const [input, options] = fetchMock.mock.calls[0];
    const headers = options?.headers as Headers;
    expect(input).toBe("/api/books/42");
    expect(options?.method).toBe("DELETE");
    expect(headers.get("Authorization")).toBe("Bearer author-token");
  });

  it("uploads cover files as multipart form data", async () => {
    frontTokenStore.set("user-token");
    const file = new File(["cover"], "cover.png", { type: "image/png" });
    const fetchMock = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) => new Response(JSON.stringify({
      coverUrl: "/api/books/1/cover",
    })));
    vi.stubGlobal("fetch", fetchMock);

    await apiClient.uploadBookCover(1, file);

    const [input, options] = fetchMock.mock.calls[0];
    const headers = options?.headers as Headers;
    expect(input).toBe("/api/books/1/cover");
    expect(options?.method).toBe("POST");
    expect(options?.body).toBeInstanceOf(FormData);
    expect(headers.get("Content-Type")).toBeNull();
    expect(headers.get("Authorization")).toBe("Bearer user-token");
  });

  it("loads public site settings without auth", async () => {
    const fetchMock = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) => new Response(JSON.stringify({
      brandName: "阅卷书屋",
      brandSubtitle: "Local Reading Archive",
      brandIconUrl: "/api/site-settings/icon",
      heroEyebrow: "发现好故事",
      heroTitle: "一站式书屋",
      heroDescription: "说明",
    })));
    vi.stubGlobal("fetch", fetchMock);

    await apiClient.siteSettings();

    const [input, options] = fetchMock.mock.calls[0];
    const headers = options?.headers as Headers;
    expect(input).toBe("/api/site-settings");
    expect(headers.get("Authorization")).toBeNull();
  });

  it("uploads site icon as multipart form data", async () => {
    adminTokenStore.set("admin-token");
    const file = new File(["icon"], "icon.svg", { type: "image/svg+xml" });
    const fetchMock = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) => new Response(JSON.stringify({
      brandName: "阅卷书屋",
      brandSubtitle: "Local Reading Archive",
      brandIconUrl: "/api/site-settings/icon",
      heroEyebrow: "发现好故事",
      heroTitle: "一站式书屋",
      heroDescription: "说明",
    })));
    vi.stubGlobal("fetch", fetchMock);

    await apiClient.uploadSiteIcon(file);

    const [input, options] = fetchMock.mock.calls[0];
    const headers = options?.headers as Headers;
    expect(input).toBe("/api/admin/site-settings/icon");
    expect(options?.body).toBeInstanceOf(FormData);
    expect(headers.get("Content-Type")).toBeNull();
    expect(headers.get("Authorization")).toBe("Bearer admin-token");
  });

  it("serializes book categoryId as a number for json create requests", async () => {
    frontTokenStore.set("user-token");
    const fetchMock = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) => new Response(JSON.stringify({
      id: 1,
      title: "测试书",
      author: "reader",
      categoryId: 6,
      category: "测试",
      description: "",
      chapterCount: 0,
    })));
    vi.stubGlobal("fetch", fetchMock);

    await apiClient.createBook({ title: "测试书", categoryId: "6", description: "" });

    const [, options] = fetchMock.mock.calls[0];
    expect(JSON.parse(options?.body as string)).toMatchObject({ categoryId: 6 });
  });

  it("uses admin bearer token for operator endpoints", async () => {
    adminTokenStore.set("reviewer-token");
    const fetchMock = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) => new Response(JSON.stringify({ items: [] })));
    vi.stubGlobal("fetch", fetchMock);

    await apiClient.listAuthorApplications();

    const [, options] = fetchMock.mock.calls[0] as [RequestInfo | URL, RequestInit];
    const headers = options?.headers as Headers;
    expect(headers.get("Authorization")).toBe("Bearer reviewer-token");
  });

  it("uses admin bearer token for recommend score updates", async () => {
    adminTokenStore.set("reviewer-token");
    const fetchMock = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) => new Response(JSON.stringify({
      id: 1,
      title: "测试书",
      author: "作者",
      category: "玄幻",
      description: "",
      chapterCount: 10,
      recommendScore: 12,
    })));
    vi.stubGlobal("fetch", fetchMock);

    await apiClient.updateRecommendScore(1, 12);

    const [input, options] = fetchMock.mock.calls[0] as [RequestInfo | URL, RequestInit];
    const headers = options?.headers as Headers;
    expect(input).toBe("/api/admin/books/1/recommend-score");
    expect(options?.method).toBe("PATCH");
    expect(headers.get("Authorization")).toBe("Bearer reviewer-token");
    expect(JSON.parse(options?.body as string)).toMatchObject({ recommendScore: 12 });
  });
});
