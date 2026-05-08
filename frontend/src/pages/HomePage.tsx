import { ArrowRight, ListOrdered, Sparkles } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { apiClient, ApiError } from "../api/client";
import type { BookSummary, Category } from "../api/types";
import { BookCard } from "../components/BookCard";
import { EmptyState, ErrorState, LoadingState } from "../components/StateViews";
import { SearchForm } from "../components/SearchForm";

export function HomePage() {
  const navigate = useNavigate();
  const [categories, setCategories] = useState<Category[]>([]);
  const [books, setBooks] = useState<BookSummary[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");

  async function load() {
    try {
      setIsLoading(true);
      setError("");
      const [categoryResponse, bookResponse] = await Promise.all([
        apiClient.categories(),
        apiClient.searchBooks({ page: 1, pageSize: 8 }),
      ]);
      setCategories(categoryResponse.items);
      setBooks(bookResponse.items);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "首页内容加载失败");
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, []);

  const rankedBooks = useMemo(
    () => [...books].sort((a, b) => b.chapterCount - a.chapterCount).slice(0, 5),
    [books],
  );

  return (
    <main className="page">
      <section className="discover-hero">
        <div className="shell hero-inner">
          <div>
            <p className="eyebrow">发现好故事</p>
            <h1>本地书库，安静阅读</h1>
            <p className="hero-copy">搜索、查看详情、登录后按章节阅读，也可以让管理员上传新的 txt 小说。</p>
          </div>
          <SearchForm
            categories={categories}
            onSubmit={(query, category) => {
              const params = new URLSearchParams();
              if (query) params.set("q", query);
              if (category) params.set("category", category);
              navigate(`/search${params.size ? `?${params}` : ""}`);
            }}
          />
        </div>
      </section>

      <section className="shell content-grid">
        <div className="main-column">
          <div className="section-heading">
            <div>
              <p className="eyebrow">新书</p>
              <h2>最近入库</h2>
            </div>
            <Link className="text-link" to="/search">
              查看更多
              <ArrowRight size={16} aria-hidden="true" />
            </Link>
          </div>

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
        </div>

        <aside className="side-column">
          <section className="panel">
            <div className="panel-title">
              <Sparkles size={18} aria-hidden="true" />
              <h2>分类入口</h2>
            </div>
            <div className="category-list">
              {categories.length ? categories.map((item) => (
                <Link key={item.id} to={`/search?category=${encodeURIComponent(item.name)}`}>
                  {item.name}
                </Link>
              )) : <span className="muted">暂无分类</span>}
            </div>
          </section>

          <section className="panel">
            <div className="panel-title">
              <ListOrdered size={18} aria-hidden="true" />
              <h2>章节榜</h2>
            </div>
            <ol className="rank-list">
              {rankedBooks.map((book) => (
                <li key={book.id}>
                  <Link to={`/books/${book.id}`}>{book.title}</Link>
                  <span>{book.chapterCount} 章</span>
                </li>
              ))}
            </ol>
            {!rankedBooks.length ? <p className="muted">暂无榜单数据</p> : null}
          </section>
        </aside>
      </section>
    </main>
  );
}
