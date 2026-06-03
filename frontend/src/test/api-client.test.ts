import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { apiClient, ApiError, tokenStore } from "../api/client";

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
    tokenStore.set("abc123");
    const fetchMock = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) => (
      new Response(JSON.stringify({ id: 1, username: "admin", role: "admin" }))
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
    tokenStore.set("admin-token");
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
    expect(headers.get("Authorization")).toBe("Bearer admin-token");
  });

  it("uses admin management endpoints with bearer token", async () => {
    tokenStore.set("admin-token");
    const fetchMock = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) => new Response(JSON.stringify({ deleted: true })));
    vi.stubGlobal("fetch", fetchMock);

    await apiClient.deleteBook(42);

    const [input, options] = fetchMock.mock.calls[0];
    const headers = options?.headers as Headers;
    expect(input).toBe("/api/books/42");
    expect(options?.method).toBe("DELETE");
    expect(headers.get("Authorization")).toBe("Bearer admin-token");
  });

  it("uploads cover files as multipart form data", async () => {
    tokenStore.set("user-token");
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

  it("serializes book categoryId as a number for json create requests", async () => {
    tokenStore.set("user-token");
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
});
