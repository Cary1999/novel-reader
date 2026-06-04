import { BookOpen, ChevronRight, Pin } from "lucide-react";
import { Link } from "react-router-dom";
import type { BookSummary } from "../api/types";

export function BookCard({
  book,
  isPinned = false,
  isInBookshelf = false,
  showBookshelfStatus = false,
}: {
  book: BookSummary;
  isPinned?: boolean;
  isInBookshelf?: boolean;
  showBookshelfStatus?: boolean;
}) {
  return (
    <article className="book-card">
      <div className="book-cover" aria-hidden="true">
        {book.coverUrl ? (
          <img className="book-cover-img" src={book.coverUrl} alt="" loading="lazy" />
        ) : (
          <>
            <div className="book-cover-spine" />
            <BookOpen size={28} />
          </>
        )}
      </div>
      <div className="book-card-body">
        <span className="book-card-count book-card-count-floating">{book.chapterCount} 章</span>
        <div className="book-title-row">
          {isPinned ? (
            <span className="book-pin-inline" aria-label="已置顶" title="已置顶">
              <Pin size={14} aria-hidden="true" />
            </span>
          ) : null}
          <h3>{book.title}</h3>
        </div>
        <p className="author">作者：{book.author}</p>
        <p className="description">{book.description || "暂无简介"}</p>
        <div className="book-card-tag-row">
          <span className="book-chip">{book.category || "未分类"}</span>
          {showBookshelfStatus && isInBookshelf ? <span className="book-chip book-chip-state">已加入书架</span> : null}
        </div>
        <div className="book-card-footer">
          <span>最新：{book.latestChapterTitle || "暂无章节"}</span>
          <Link to={`/books/${book.id}`} aria-label={`查看《${book.title}》详情`}>
            详情
            <ChevronRight size={16} aria-hidden="true" />
          </Link>
        </div>
      </div>
    </article>
  );
}
