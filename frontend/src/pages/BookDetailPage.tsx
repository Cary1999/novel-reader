import { BookOpen, List, LockKeyhole, Plus, ScrollText } from "lucide-react";
import { useEffect, useState } from "react";
import { Link, useLocation, useNavigate, useParams } from "react-router-dom";
import { apiClient, ApiError } from "../api/client";
import type { BookDetail, ChapterSummary } from "../api/types";
import { useAuth } from "../auth/AuthContext";
import { EmptyState, ErrorState, LoadingState } from "../components/StateViews";

export function BookDetailPage() {
  const { bookId = "" } = useParams();
  const navigate = useNavigate();
  const location = useLocation();
  const { isAuthenticated } = useAuth();
  const [book, setBook] = useState<BookDetail | null>(null);
  const [chapters, setChapters] = useState<ChapterSummary[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");

  async function load() {
    try {
      setIsLoading(true);
      setError("");
      const [bookResponse, chapterResponse] = await Promise.all([
        apiClient.book(bookId),
        apiClient.chapters(bookId),
      ]);
      setBook(bookResponse);
      setChapters(chapterResponse.items ?? []);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "书籍详情加载失败");
    } finally {
      setIsLoading(false);
    }
  }

  async function addToBookshelf() {
    try {
      setMessage("");
      await apiClient.addToBookshelf(bookId);
      setMessage("已加入书架");
    } catch (err) {
      setMessage(err instanceof ApiError ? err.message : "加入书架失败");
    }
  }

  useEffect(() => {
    void load();
  }, [bookId]);

  function openChapter(chapterId: number) {
    if (!isAuthenticated) {
      navigate("/login", { state: { from: location } });
      return;
    }
    navigate(`/books/${bookId}/chapters/${chapterId}`);
  }

  if (isLoading) {
    return <main className="page shell"><LoadingState label="正在加载书籍详情..." /></main>;
  }

  if (error) {
    return <main className="page shell"><ErrorState message={error} onRetry={load} /></main>;
  }

  if (!book) {
    return <main className="page shell"><EmptyState title="书籍不存在" description="它可能已经被移除。" /></main>;
  }

  const latestChapter = chapters.at(-1);

  return (
    <main className="page shell">
      <section className="detail-header">
        <div className="detail-cover" aria-hidden="true">
          {book.coverUrl ? (
            <img className="book-cover-img" src={book.coverUrl} alt="" loading="lazy" />
          ) : (
            <>
              <div className="detail-cover-spine" />
              <BookOpen size={50} />
            </>
          )}
        </div>
        <div>
          <p className="eyebrow">{book.category || "未分类"}</p>
          <h1>{book.title}</h1>
          <div className="detail-meta">
            <span>作者：{book.author}</span>
            <span>{book.chapterCount} 章</span>
            <span>最新：{latestChapter?.title ?? "暂无章节"}</span>
          </div>
          <p className="detail-description">{book.description || "暂无简介"}</p>
          <div className="detail-actions">
            {chapters[0] ? (
              <button className="primary-button" type="button" onClick={() => openChapter(chapters[0].id)}>
                <BookOpen size={18} aria-hidden="true" />
                开始阅读
              </button>
            ) : null}
            {isAuthenticated ? (
              <button className="ghost-button" type="button" onClick={() => void addToBookshelf()}>
                <Plus size={18} aria-hidden="true" />
                加入书架
              </button>
            ) : null}
            <Link className="ghost-button" to="/">返回首页</Link>
          </div>
          {message ? <p className="success-banner">{message}</p> : null}
          <div className="detail-notes">
            <span><ScrollText size={16} aria-hidden="true" />目录随上传和编辑实时更新</span>
            <span><List size={16} aria-hidden="true" />支持按章节连续阅读</span>
          </div>
        </div>
      </section>

      {!isAuthenticated ? (
        <section className="panel permission-banner">
          <LockKeyhole size={20} aria-hidden="true" />
          <span>游客可以查看详情和目录，章节正文需要登录后阅读。</span>
          <Link className="text-link" to="/login" state={{ from: location }}>去登录</Link>
        </section>
      ) : null}

      <section className="chapter-section">
        <div className="section-heading">
          <div>
            <p className="eyebrow">目录</p>
            <h2>章节目录</h2>
          </div>
          <span className="muted">{chapters.length} 章</span>
        </div>
        {chapters.length ? (
          <div className="chapter-list">
            {chapters.map((chapter) => (
              <button key={chapter.id} type="button" onClick={() => openChapter(chapter.id)}>
                <List size={16} aria-hidden="true" />
                <span>第 {chapter.index} 章</span>
                <strong>{chapter.title}</strong>
                {!isAuthenticated ? <LockKeyhole size={15} aria-label="需登录" /> : null}
              </button>
            ))}
          </div>
        ) : (
          <EmptyState title="暂无章节" description="上传解析完成后目录会显示在这里。" />
        )}
      </section>
    </main>
  );
}
