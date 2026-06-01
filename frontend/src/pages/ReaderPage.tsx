import { ChevronLeft, ChevronRight, List, LockKeyhole, Moon, Sun, Type } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { Link, useLocation, useNavigate, useParams } from "react-router-dom";
import { apiClient, ApiError } from "../api/client";
import type { BookDetail, ChapterDetail, ChapterSummary } from "../api/types";
import { useAuth } from "../auth/AuthContext";
import { EmptyState, ErrorState, LoadingState } from "../components/StateViews";

type ReaderTheme = "light" | "dark";

export function ReaderPage() {
  const { bookId = "", chapterId = "" } = useParams();
  const navigate = useNavigate();
  const location = useLocation();
  const { isAuthenticated, isLoading: authLoading } = useAuth();
  const [book, setBook] = useState<BookDetail | null>(null);
  const [chapter, setChapter] = useState<ChapterDetail | null>(null);
  const [chapters, setChapters] = useState<ChapterSummary[]>([]);
  const [fontSize, setFontSize] = useState(18);
  const [theme, setTheme] = useState<ReaderTheme>("light");
  const [showToc, setShowToc] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");

  const currentIndex = useMemo(
    () => chapters.findIndex((item) => String(item.id) === String(chapterId)),
    [chapterId, chapters],
  );
  const previousChapter = currentIndex > 0 ? chapters[currentIndex - 1] : null;
  const nextChapter = currentIndex >= 0 && currentIndex < chapters.length - 1 ? chapters[currentIndex + 1] : null;

  async function loadPublicData() {
    const [bookResponse, chapterResponse] = await Promise.all([
      apiClient.book(bookId),
      apiClient.chapters(bookId),
    ]);
    setBook(bookResponse);
    setChapters(chapterResponse.items ?? []);
  }

  async function load() {
    try {
      setIsLoading(true);
      setError("");
      await loadPublicData();
      if (isAuthenticated) {
        setChapter(await apiClient.chapter(bookId, chapterId));
      }
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        setError("");
      } else {
        setError(err instanceof ApiError ? err.message : "章节加载失败");
      }
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    if (!authLoading) {
      void load();
    }
  }, [authLoading, isAuthenticated, bookId, chapterId]);

  function openChapter(nextId: number) {
    setShowToc(false);
    navigate(`/books/${bookId}/chapters/${nextId}`);
  }

  if (authLoading || isLoading) {
    return <main className="page shell"><LoadingState label="正在打开章节..." /></main>;
  }

  if (!isAuthenticated) {
    return (
      <main className="page shell">
        <section className="panel permission-panel">
          <LockKeyhole size={28} aria-hidden="true" />
          <p className="eyebrow">需要登录</p>
          <h1>章节正文仅对登录用户开放</h1>
          <p className="muted">你仍然可以查看《{book?.title ?? "当前小说"}》的详情和章节目录。</p>
          <div className="detail-actions">
            <Link className="primary-button compact" to="/login" state={{ from: location }}>登录阅读</Link>
            <Link className="ghost-button" to={`/books/${bookId}`}>查看详情</Link>
          </div>
        </section>
      </main>
    );
  }

  if (error) {
    return <main className="page shell"><ErrorState message={error} onRetry={load} /></main>;
  }

  if (!chapter) {
    return <main className="page shell"><EmptyState title="章节不存在" description="请从目录选择其他章节。" /></main>;
  }

  const paragraphs = chapter.content
    .split(/\n+/)
    .map((line) => line.trim())
    .filter(Boolean);

  return (
    <main className={`reader-page reader-${theme}`}>
      <div className="reader-toolbar">
        <Link className="ghost-button compact" to={`/books/${bookId}`}>返回详情</Link>
        <div className="reader-controls">
          <button className="ghost-button icon-button" type="button" onClick={() => setShowToc((value) => !value)} title="章节目录">
            <List size={18} aria-hidden="true" />
            <span>目录</span>
          </button>
          <label className="reader-size-control">
            <Type size={18} aria-hidden="true" />
            <span className="sr-only">字号</span>
            <input
              type="range"
              min="16"
              max="24"
              value={fontSize}
              onChange={(event) => setFontSize(Number(event.target.value))}
            />
            <span>{fontSize}px</span>
          </label>
          <button className="ghost-button icon-button" type="button" onClick={() => setTheme(theme === "light" ? "dark" : "light")} title="切换主题">
            {theme === "light" ? <Moon size={18} aria-hidden="true" /> : <Sun size={18} aria-hidden="true" />}
            <span>{theme === "light" ? "深色" : "浅色"}</span>
          </button>
        </div>
      </div>

      <div className="reader-layout">
        <aside className={`toc-panel ${showToc ? "is-open" : ""}`} aria-label="章节目录">
          <div className="toc-heading">
            <strong>{book?.title ?? "目录"}</strong>
            <button className="ghost-button compact" type="button" onClick={() => setShowToc(false)}>收起</button>
          </div>
          <div className="toc-list">
            {chapters.map((item) => (
              <button
                key={item.id}
                className={item.id === chapter.id ? "active" : ""}
                type="button"
                onClick={() => openChapter(item.id)}
              >
                <span>{item.index}</span>
                {item.title}
              </button>
            ))}
          </div>
        </aside>

        <article className="reader-article" style={{ fontSize }}>
          <div className="reader-article-head">
            <p className="eyebrow">{book?.title}</p>
            <span className="reader-progress">第 {chapter.index} 章 / 共 {chapters.length} 章</span>
          </div>
          <h1>{chapter.title}</h1>
          {paragraphs.map((paragraph, index) => (
            <p key={`${chapter.id}-${index}`}>{paragraph}</p>
          ))}
          {!paragraphs.length ? <p>本章暂无正文。</p> : null}
          <nav className="reader-chapter-nav" aria-label="章节切换">
            <button className="ghost-button" type="button" disabled={!previousChapter} onClick={() => previousChapter && openChapter(previousChapter.id)}>
              <ChevronLeft size={16} aria-hidden="true" />
              上一章
            </button>
            <button className="ghost-button" type="button" onClick={() => setShowToc(true)}>
              <List size={16} aria-hidden="true" />
              目录
            </button>
            <button className="ghost-button" type="button" disabled={!nextChapter} onClick={() => nextChapter && openChapter(nextChapter.id)}>
              下一章
              <ChevronRight size={16} aria-hidden="true" />
            </button>
          </nav>
        </article>
      </div>
    </main>
  );
}
