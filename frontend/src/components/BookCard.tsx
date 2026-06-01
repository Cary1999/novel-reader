import { BookOpen, ChevronRight } from "lucide-react";
import { Link } from "react-router-dom";
import type { BookSummary } from "../api/types";

export function BookCard({ book }: { book: BookSummary }) {
  return (
    <article className="book-card">
      <div className="book-cover" aria-hidden="true">
        <div className="book-cover-spine" />
        <BookOpen size={28} />
      </div>
      <div className="book-card-body">
        <div className="book-card-head">
          <span className="book-chip">{book.category || "未分类"}</span>
          <span className="book-card-count">{book.chapterCount} 章</span>
        </div>
        <h3>{book.title}</h3>
        <p className="author">作者：{book.author}</p>
        <p className="description">{book.description || "暂无简介"}</p>
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
