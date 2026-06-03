import { ArrowRight, Compass, Flame, Sparkles, ThumbsUp } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { apiClient, ApiError } from "../api/client";
import type { BookSummary, Category } from "../api/types";
import { BookCard } from "../components/BookCard";
import { EmptyState, ErrorState, LoadingState } from "../components/StateViews";
import { SearchForm } from "../components/SearchForm";
import { useSiteSettings } from "../site/SiteSettingsContext";

const INLINE_LIMIT = 8;

export function HomePage() {
  const navigate = useNavigate();
  const { settings } = useSiteSettings();
  const [categories, setCategories] = useState<Category[]>([]);
  const [books, setBooks] = useState<BookSummary[]>([]);
  const [recommendations, setRecommendations] = useState<BookSummary[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [inlineQuery, setInlineQuery] = useState("");
  const [inlineCategory, setInlineCategory] = useState("");
  const [inlineResults, setInlineResults] = useState<BookSummary[] | null>(null);
  const [inlineTotal, setInlineTotal] = useState(0);
  const [isInlineSearching, setIsInlineSearching] = useState(false);
  const [inlineError, setInlineError] = useState("");

  async function load() {
    try {
      setIsLoading(true);
      setError("");
      const [categoryResponse, bookResponse, recommendationResponse] = await Promise.all([
        apiClient.categories(),
        apiClient.searchBooks({ page: 1, pageSize: 8 }),
        apiClient.recommendations({ page: 1, pageSize: 8 }),
      ]);
      setCategories(categoryResponse.items);
      setBooks(bookResponse.items);
      setRecommendations(recommendationResponse.items ?? []);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "首页内容加载失败");
    } finally {
      setIsLoading(false);
    }
  }

  async function runInlineSearch(nextQuery: string, nextCategory: string) {
    try {
      setIsInlineSearching(true);
      setInlineError("");
      setInlineQuery(nextQuery);
      setInlineCategory(nextCategory);
      const response = await apiClient.searchBooks({ q: nextQuery || undefined, category: nextCategory || undefined, page: 1, pageSize: INLINE_LIMIT });
      setInlineResults(response.items ?? []);
      setInlineTotal(response.total ?? 0);
    } catch (err) {
      setInlineError(err instanceof ApiError ? err.message : "搜索失败，请稍后再试");
      setInlineResults([]);
      setInlineTotal(0);
    } finally {
      setIsInlineSearching(false);
    }
  }

  useEffect(() => {
    void load();
  }, []);

  const rankedBooks = useMemo(() => recommendations.slice(0, 5), [recommendations]);

  return (
    <main className="page">
      <section className="discover-hero">
        <div className="shell hero-inner">
          <div className="hero-copy-block">
            <p className="eyebrow">{settings.heroEyebrow}</p>
            <h1>{settings.heroTitle}</h1>
            <p className="hero-copy">{settings.heroDescription}</p>
            <div className="hero-stats">
              <div>
                <strong>{categories.length}</strong>
                <span>已收录分类</span>
              </div>
              <div>
                <strong>{books.length}</strong>
                <span>首页展示新书</span>
              </div>
              <div>
                <strong>{rankedBooks.length}</strong>
                <span>当前榜单条目</span>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="shell home-search-row" aria-label="搜索书库">
        <SearchForm
          initialQuery={inlineQuery}
          initialCategory={inlineCategory}
          categories={categories}
          onSubmit={(query, category) => void runInlineSearch(query, category)}
        />
      </section>

      <section className="shell content-grid">
        <div className="main-column">
          <div className="section-heading">
            <div>
              <p className="eyebrow">{inlineResults ? "搜索结果" : "新书"}</p>
              <h2>{inlineResults ? "为你找到" : "最新入库"}</h2>
            </div>
            {inlineResults ? (
              inlineTotal > INLINE_LIMIT ? (
                <button
                  className="text-link"
                  type="button"
                  onClick={() => {
                    const params = new URLSearchParams();
                    if (inlineQuery) params.set("q", inlineQuery);
                    if (inlineCategory) params.set("category", inlineCategory);
                    navigate(`/search${params.size ? `?${params}` : ""}`);
                  }}
                >
                  查看更多
                  <ArrowRight size={16} aria-hidden="true" />
                </button>
              ) : (
                <button className="text-link" type="button" onClick={() => {
                  setInlineResults(null);
                  setInlineTotal(0);
                  setInlineError("");
                  setInlineQuery("");
                  setInlineCategory("");
                }}>
                  返回最新入库
                </button>
              )
            ) : (
              <Link className="text-link" to="/search">
                查看更多
                <ArrowRight size={16} aria-hidden="true" />
              </Link>
            )}
          </div>

          {inlineResults ? (
            <>
              {isInlineSearching ? <LoadingState label="正在搜索..." /> : null}
              {inlineError ? <ErrorState message={inlineError} onRetry={() => void runInlineSearch(inlineQuery, inlineCategory)} /> : null}
              {!isInlineSearching && !inlineError && inlineResults.length === 0 ? (
                <EmptyState title="没有找到匹配小说" description="换一个关键词或分类再试试。" />
              ) : null}
              {!isInlineSearching && !inlineError && inlineResults.length > 0 ? (
                <>
                  <div className="result-summary">共 {inlineTotal} 本，当前展示前 {Math.min(inlineTotal, INLINE_LIMIT)} 本</div>
                  <div className="book-grid">
                    {inlineResults.map((book) => <BookCard key={book.id} book={book} />)}
                  </div>
                  {inlineTotal > INLINE_LIMIT ? (
                    <div className="inline-more">
                      <button
                        className="ghost-button compact"
                        type="button"
                        onClick={() => {
                          const params = new URLSearchParams();
                          if (inlineQuery) params.set("q", inlineQuery);
                          if (inlineCategory) params.set("category", inlineCategory);
                          navigate(`/search${params.size ? `?${params}` : ""}`);
                        }}
                      >
                        查看更多
                      </button>
                    </div>
                  ) : null}
                </>
              ) : null}
            </>
          ) : (
            <>
              {isLoading ? <LoadingState label="正在加载书库..." /> : null}
              {error ? <ErrorState message={error} onRetry={load} /> : null}
              {!isLoading && !error && books.length === 0 ? (
                <EmptyState title="还没有小说" description="上传后会在这里展示新书。" />
              ) : null}
              {!isLoading && !error && books.length > 0 ? (
                <div className="book-grid">
                  {books.map((book) => <BookCard key={book.id} book={book} />)}
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
                  onClick={() => void runInlineSearch("", item.name)}
                >
                  {item.name}
                </button>
              )) : <span className="muted">暂无分类</span>}
            </div>
          </section>

          <section className="panel">
            <div className="panel-title">
              <ThumbsUp size={18} aria-hidden="true" />
              <h2>推荐榜单</h2>
            </div>
            <ol className="rank-list">
              {rankedBooks.map((book) => (
                <li key={book.id}>
                  <Link to={`/books/${book.id}`}>{book.title}</Link>
                  <span>{book.author}</span>
                </li>
              ))}
            </ol>
            {!rankedBooks.length ? <p className="muted">暂无榜单数据</p> : null}
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
