import { ChevronLeft, ChevronRight } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { apiClient, ApiError } from "../api/client";
import type { BookSummary, Category } from "../api/types";
import { BookCard } from "../components/BookCard";
import { EmptyState, ErrorState, LoadingState } from "../components/StateViews";
import { SearchForm } from "../components/SearchForm";

const PAGE_SIZE = 10;

export function SearchPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const query = searchParams.get("q") ?? "";
  const category = searchParams.get("category") ?? "";
  const page = Math.max(Number(searchParams.get("page") ?? "1"), 1);
  const [categories, setCategories] = useState<Category[]>([]);
  const [books, setBooks] = useState<BookSummary[]>([]);
  const [total, setTotal] = useState(0);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");

  const totalPages = Math.max(Math.ceil(total / PAGE_SIZE), 1);

  const pageTitle = useMemo(() => {
    if (query && category) return `搜索“${query}” · ${category}`;
    if (query) return `搜索“${query}”`;
    if (category) return `${category} 分类`;
    return "全部小说";
  }, [category, query]);

  async function load() {
    try {
      setIsLoading(true);
      setError("");
      const [categoryResponse, bookResponse] = await Promise.all([
        apiClient.categories(),
        apiClient.searchBooks({ q: query, category, page, pageSize: PAGE_SIZE }),
      ]);
      setCategories(categoryResponse.items ?? []);
      setBooks(bookResponse.items ?? []);
      setTotal(bookResponse.total ?? 0);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "搜索结果加载失败");
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, [query, category, page]);

  function goToPage(nextPage: number) {
    const params = new URLSearchParams();
    if (query) params.set("q", query);
    if (category) params.set("category", category);
    if (nextPage > 1) params.set("page", String(nextPage));
    navigate(`/search${params.size ? `?${params}` : ""}`);
  }

  return (
    <main className="page shell">
      <section className="search-toolbar">
        <div>
          <p className="eyebrow">搜索</p>
          <h1>{pageTitle}</h1>
        </div>
        <SearchForm
          initialQuery={query}
          initialCategory={category}
          categories={categories}
          onSubmit={(nextQuery, nextCategory) => {
            const params = new URLSearchParams();
            if (nextQuery) params.set("q", nextQuery);
            if (nextCategory) params.set("category", nextCategory);
            navigate(`/search${params.size ? `?${params}` : ""}`);
          }}
        />
      </section>

      {isLoading ? <LoadingState label="正在搜索..." /> : null}
      {error ? <ErrorState message={error} onRetry={load} /> : null}
      {!isLoading && !error && books.length === 0 ? (
        <EmptyState title="没有找到匹配小说" description="换一个关键词或分类再试试。" />
      ) : null}
      {!isLoading && !error && books.length > 0 ? (
        <>
          <div className="result-summary">共 {total} 本，当前第 {page} / {totalPages} 页</div>
          <div className="book-grid search-results">
            {books.map((book) => <BookCard key={book.id} book={book} />)}
          </div>
          <div className="pagination">
            <button className="ghost-button" type="button" onClick={() => goToPage(page - 1)} disabled={page <= 1}>
              <ChevronLeft size={16} aria-hidden="true" />
              上一页
            </button>
            <span>{page} / {totalPages}</span>
            <button className="ghost-button" type="button" onClick={() => goToPage(page + 1)} disabled={page >= totalPages}>
              下一页
              <ChevronRight size={16} aria-hidden="true" />
            </button>
          </div>
        </>
      ) : null}
    </main>
  );
}
