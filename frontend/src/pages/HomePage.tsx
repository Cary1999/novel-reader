import { ChevronLeft, ChevronRight, Flame, Sparkles } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { apiClient, ApiError } from "../api/client";
import type { BookSummary, Category } from "../api/types";
import { BookCard } from "../components/BookCard";
import { EmptyState, ErrorState, LoadingState } from "../components/StateViews";
import { SearchForm } from "../components/SearchForm";
import { useSiteSettings } from "../site/SiteSettingsContext";

const HOMEPAGE_LIMIT = 10;
const SEARCH_PAGE_SIZE_OPTIONS = [10, 20, 50] as const;

function clampPageSize(value: number) {
  if (value < 1) return 1;
  if (value > 100) return 100;
  return value;
}

function clampPage(value: number) {
  if (value < 1) return 1;
  return value;
}

function parseIntParam(value: string | null, fallback: number) {
  if (!value) return fallback;
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallback;
}

export function HomePage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { settings } = useSiteSettings();
  const [categories, setCategories] = useState<Category[]>([]);
  const [newBooks, setNewBooks] = useState<BookSummary[]>([]);
  const [recommendationBooks, setRecommendationBooks] = useState<BookSummary[]>([]);
  const [isLoadingDiscovery, setIsLoadingDiscovery] = useState(true);
  const [discoveryError, setDiscoveryError] = useState("");
  const [searchBooks, setSearchBooks] = useState<BookSummary[]>([]);
  const [searchTotal, setSearchTotal] = useState(0);
  const [isSearching, setIsSearching] = useState(false);
  const [searchError, setSearchError] = useState("");
  const [pageSizeMode, setPageSizeMode] = useState<string>("10");
  const [customPageSize, setCustomPageSize] = useState("10");

  const query = searchParams.get("q") ?? "";
  const category = searchParams.get("category") ?? "";
  const page = clampPage(parseIntParam(searchParams.get("page"), 1));
  const pageSize = clampPageSize(parseIntParam(searchParams.get("pageSize"), 10));
  const activeSearch = Boolean(query.trim() || category.trim());

  const totalPages = Math.max(Math.ceil(searchTotal / pageSize), 1);
  const searchMode = useMemo(() => {
    if (SEARCH_PAGE_SIZE_OPTIONS.includes(pageSize as (typeof SEARCH_PAGE_SIZE_OPTIONS)[number])) {
      return String(pageSize);
    }
    return "custom";
  }, [pageSize]);

  async function loadDiscovery() {
    try {
      setIsLoadingDiscovery(true);
      setDiscoveryError("");
      const [categoryResponse, bookResponse, recommendationResponse] = await Promise.all([
        apiClient.categories(),
        apiClient.searchBooks({ page: 1, pageSize: HOMEPAGE_LIMIT }),
        apiClient.recommendations({ page: 1, pageSize: HOMEPAGE_LIMIT }),
      ]);
      setCategories(categoryResponse.items ?? []);
      setNewBooks(bookResponse.items ?? []);
      setRecommendationBooks(recommendationResponse.items ?? []);
    } catch (err) {
      setDiscoveryError(err instanceof ApiError ? err.message : "首页内容加载失败");
    } finally {
      setIsLoadingDiscovery(false);
    }
  }

  async function loadSearchResults() {
    if (!activeSearch) {
      setSearchBooks([]);
      setSearchTotal(0);
      setSearchError("");
      setIsSearching(false);
      return;
    }

    try {
      setIsSearching(true);
      setSearchError("");
      const response = await apiClient.searchBooks({
        q: query || undefined,
        category: category || undefined,
        page,
        pageSize,
      });
      setSearchBooks(response.items ?? []);
      setSearchTotal(response.total ?? 0);
      setPageSizeMode(searchMode);
      setCustomPageSize(searchMode === "custom" ? String(pageSize) : String(pageSize));
    } catch (err) {
      setSearchError(err instanceof ApiError ? err.message : "搜索失败，请稍后再试");
      setSearchBooks([]);
      setSearchTotal(0);
    } finally {
      setIsSearching(false);
    }
  }

  useEffect(() => {
    void loadDiscovery();
  }, []);

  useEffect(() => {
    void loadSearchResults();
  }, [activeSearch, category, page, pageSize, query]);

  useEffect(() => {
    setPageSizeMode(searchMode);
    setCustomPageSize(String(pageSize));
  }, [pageSize, searchMode]);

  function goToSearch(nextQuery: string, nextCategory: string, nextPage = 1, nextPageSize = pageSize) {
    const params = new URLSearchParams();
    if (nextQuery) params.set("q", nextQuery);
    if (nextCategory) params.set("category", nextCategory);
    params.set("page", String(nextPage));
    params.set("pageSize", String(clampPageSize(nextPageSize)));
    navigate(`/${params.size ? `?${params}` : ""}`);
  }

  function applyPageSize(nextPageSize: number) {
    goToSearch(query, category, 1, nextPageSize);
  }

  function gotoPage(nextPage: number) {
    goToSearch(query, category, nextPage, pageSize);
  }

  const rankedBooks = useMemo(() => recommendationBooks.slice(0, HOMEPAGE_LIMIT), [recommendationBooks]);
  const freshBooks = useMemo(() => newBooks.slice(0, HOMEPAGE_LIMIT), [newBooks]);
  const showSearchResults = activeSearch;

  return (
    <main className="page">
      <section className="discover-hero">
        <div className="shell hero-inner">
          <div className="hero-copy-block">
            <p className="eyebrow">{settings.heroEyebrow}</p>
            <h1>{settings.heroTitle}</h1>
            <p className="hero-copy">{settings.heroDescription}</p>
          </div>
        </div>
      </section>

      <section className="shell home-search-row" aria-label="搜索书库">
        <SearchForm
          initialQuery={query}
          initialCategory={category}
          categories={categories}
          onSubmit={(nextQuery, nextCategory) => goToSearch(nextQuery, nextCategory, 1, pageSize)}
        />
      </section>

      <section className="shell content-grid">
        <div className="main-column">
          <div className="section-heading">
            <div>
              <h2>{showSearchResults ? "为你找到" : "为你推荐"}</h2>
            </div>
          </div>

          {showSearchResults ? (
            <>
              {isSearching ? <LoadingState label="正在搜索..." /> : null}
              {searchError ? <ErrorState message={searchError} onRetry={loadSearchResults} /> : null}
              {!isSearching && !searchError && searchBooks.length === 0 ? (
                <EmptyState title="没有找到匹配小说" description="换一个关键词、作者名或分类再试试。" />
              ) : null}
              {!isSearching && !searchError && searchBooks.length > 0 ? (
                <>
                  <div className="result-summary">共 {searchTotal} 本，当前第 {page} / {totalPages} 页</div>
                  <div className="book-grid search-results">
                    {searchBooks.map((book) => <BookCard key={book.id} book={book} />)}
                  </div>
                  <div className="pagination bookshelf-pagination">
                    <button className="ghost-button" type="button" onClick={() => gotoPage(page - 1)} disabled={page <= 1}>
                      <ChevronLeft size={16} aria-hidden="true" />
                      上一页
                    </button>
                    <span>{page} / {totalPages}</span>
                    <button className="ghost-button" type="button" onClick={() => gotoPage(page + 1)} disabled={page >= totalPages}>
                      下一页
                      <ChevronRight size={16} aria-hidden="true" />
                    </button>
                    <div className="page-size-control pagination-page-size inline">
                      <select
                        value={pageSizeMode}
                        onChange={(event) => {
                          const value = event.target.value;
                          setPageSizeMode(value);
                          if (value === "custom") {
                            return;
                          }
                          applyPageSize(Number(value));
                        }}
                      >
                        {SEARCH_PAGE_SIZE_OPTIONS.map((value) => (
                          <option key={value} value={value}>{value}</option>
                        ))}
                        <option value="custom">自定义</option>
                      </select>
                      {pageSizeMode === "custom" ? (
                        <>
                          <input
                            type="number"
                            min={1}
                            max={100}
                            value={customPageSize}
                            onChange={(event) => setCustomPageSize(event.target.value)}
                          />
                          <button
                            className="ghost-button compact"
                            type="button"
                            onClick={() => {
                              const value = clampPageSize(Number(customPageSize) || 10);
                              applyPageSize(value);
                            }}
                          >
                            应用
                          </button>
                        </>
                      ) : null}
                    </div>
                  </div>
                </>
              ) : null}
            </>
          ) : (
            <>
              {isLoadingDiscovery ? <LoadingState label="正在加载书库..." /> : null}
              {discoveryError ? <ErrorState message={discoveryError} onRetry={loadDiscovery} /> : null}
              {!isLoadingDiscovery && !discoveryError && rankedBooks.length === 0 ? (
                <EmptyState title="还没有推荐内容" description="等管理员调整推荐值后，这里会展示榜单。" />
              ) : null}
              {!isLoadingDiscovery && !discoveryError && rankedBooks.length > 0 ? (
                <div className="book-grid">
                  {rankedBooks.map((book) => <BookCard key={book.id} book={book} />)}
                </div>
              ) : null}
            </>
          )}
        </div>

        <aside className="side-column">
          <section className="panel">
            <div className="panel-title">
              <Sparkles size={18} aria-hidden="true" />
              <h2>快速分类</h2>
            </div>
            <div className="category-list">
              {categories.length ? categories.map((item) => (
                <button
                  key={item.id}
                  type="button"
                  className="category-pill"
                  onClick={() => goToSearch("", item.name, 1, pageSize)}
                >
                  {item.name}
                </button>
              )) : <span className="muted">暂无分类</span>}
            </div>
          </section>

          <section className="panel">
            <div className="panel-title">
              <Sparkles size={18} aria-hidden="true" />
              <h2>新书速递</h2>
            </div>
            <ol className="rank-list new-book-list">
              {freshBooks.map((book) => (
                <li key={book.id}>
                  <Link to={`/books/${book.id}`}>{book.title}</Link>
                  <span>{book.author}</span>
                </li>
              ))}
            </ol>
            {!freshBooks.length ? <p className="muted">暂无新书</p> : null}
          </section>

          <section className="panel ambient-panel">
            <div className="panel-title">
              <Flame size={18} aria-hidden="true" />
              <h2>阅读提示</h2>
            </div>
            <p className="muted">游客可搜索与查看详情，登录后可直接进入章节阅读；上传后的小说也会在书库中展示。</p>
          </section>
        </aside>
      </section>
    </main>
  );
}
