import { ChevronLeft, ChevronRight } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { Link, useNavigate, useParams, useSearchParams } from "react-router-dom";
import { apiClient, ApiError } from "../api/client";
import type { BookshelfEntry, BookshelfGroup } from "../api/types";
import { BookCard } from "../components/BookCard";
import { BookshelfBookMenu, BookshelfMoveDialog, BookshelfRenameGroupDialog } from "../components/BookshelfMenus";
import { EmptyState, ErrorState, LoadingState } from "../components/StateViews";

const PAGE_SIZE_OPTIONS = [10, 20, 50] as const;

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

export function BookshelfGroupPage() {
  const navigate = useNavigate();
  const { groupId = "" } = useParams();
  const [searchParams, setSearchParams] = useSearchParams();
  const [group, setGroup] = useState<BookshelfGroup | null>(null);
  const [groups, setGroups] = useState<BookshelfGroup[]>([]);
  const [entries, setEntries] = useState<BookshelfEntry[]>([]);
  const [total, setTotal] = useState(0);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [pageSizeMode, setPageSizeMode] = useState("20");
  const [customPageSize, setCustomPageSize] = useState("20");
  const [movingEntry, setMovingEntry] = useState<BookshelfEntry | null>(null);
  const [renamingGroup, setRenamingGroup] = useState<BookshelfGroup | null>(null);

  const page = clampPage(parseIntParam(searchParams.get("page"), 1));
  const pageSize = clampPageSize(parseIntParam(searchParams.get("pageSize"), 20));
  const totalPages = Math.max(Math.ceil(total / pageSize), 1);
  const currentGroupId = Number(groupId);

  const pageSizeCurrentMode = useMemo(() => {
    if (PAGE_SIZE_OPTIONS.includes(pageSize as (typeof PAGE_SIZE_OPTIONS)[number])) {
      return String(pageSize);
    }
    return "custom";
  }, [pageSize]);

  async function load() {
    try {
      setIsLoading(true);
      setError("");
      const [groupResponse, groupsResponse, entriesResponse] = await Promise.all([
        apiClient.bookshelfGroup(currentGroupId),
        apiClient.bookshelfGroups(),
        apiClient.bookshelf({ groupId: currentGroupId, page, pageSize }),
      ]);
      setGroup(groupResponse);
      setGroups(groupsResponse.items ?? []);
      setEntries(entriesResponse.items ?? []);
      setTotal(entriesResponse.total ?? 0);
      setPageSizeMode(pageSizeCurrentMode);
      setCustomPageSize(String(pageSize));
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "分组详情加载失败");
      setEntries([]);
      setTotal(0);
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, [currentGroupId, page, pageSize]);

  useEffect(() => {
    setPageSizeMode(pageSizeCurrentMode);
    setCustomPageSize(String(pageSize));
  }, [pageSize, pageSizeCurrentMode]);

  function updatePage(nextPage: number, nextPageSize = pageSize) {
    const params = new URLSearchParams();
    if (nextPage > 1) params.set("page", String(nextPage));
    if (nextPageSize !== 20) params.set("pageSize", String(clampPageSize(nextPageSize)));
    setSearchParams(params);
  }

  async function renameGroup() {
    if (!group) return;
    setRenamingGroup(group);
  }

  async function submitRenameGroup(targetGroup: BookshelfGroup, nextName: string) {
    try {
      const updated = await apiClient.updateBookshelfGroup(targetGroup.id, { name: nextName });
      setGroup(updated);
      setRenamingGroup(null);
      await load();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "重命名分组失败");
    }
  }

  async function deleteGroup() {
    if (!group) return;
    if (!globalThis.confirm(`确定删除分组“${group.name}”吗？组内书籍会回到书架首页。`)) {
      return;
    }
    try {
      await apiClient.deleteBookshelfGroup(group.id);
      navigate("/bookshelf");
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "删除分组失败");
    }
  }

  async function togglePin(entry: BookshelfEntry) {
    try {
      await apiClient.updateBookshelfBook(entry.bookId, { pinned: !entry.isPinned });
      await load();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "置顶操作失败");
    }
  }

  async function moveEntry(bookId: number, targetGroupId: number) {
    try {
      await apiClient.updateBookshelfBook(bookId, { groupId: targetGroupId });
      await load();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "移动书籍失败");
    }
  }

  async function createGroupAndMove(bookId: number, groupName: string) {
    try {
      const created = await apiClient.createBookshelfGroup({ name: groupName });
      await apiClient.updateBookshelfBook(bookId, { groupId: created.id });
      await load();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "创建分组失败");
    }
  }

  async function removeEntry(bookId: number) {
    try {
      await apiClient.removeFromBookshelf(bookId);
      await load();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "移出书架失败");
    }
  }

  if (isLoading) {
    return <main className="page shell"><LoadingState label="正在加载分组..." /></main>;
  }

  if (error && !group) {
    return <main className="page shell"><ErrorState message={error} onRetry={load} /></main>;
  }

  if (!group) {
    return <main className="page shell"><EmptyState title="分组不存在" description="它可能已经被删除，或者你没有权限访问。" /></main>;
  }

  return (
    <main className="page shell">
      <section className="page-banner">
        <div>
          <h1>{group.name}</h1>
          <p className="muted">这个分组里现在有 {total} 本书。你可以继续移动，整理完再回到书架首页。</p>
        </div>
        <div className="bookshelf-detail-actions">
          <button className="ghost-button compact" type="button" onClick={() => void renameGroup()}>
            重命名
          </button>
          <button className="ghost-button compact danger" type="button" onClick={() => void deleteGroup()}>
            删除分组
          </button>
          <Link className="ghost-button compact" to="/bookshelf">
            返回书架
          </Link>
        </div>
      </section>

      {error ? <p className="form-error">{error}</p> : null}
      
      {!entries.length ? (
        <EmptyState title="这个分组还是空的" description="可以从书架首页的书籍菜单里，把想看的书移动进来。" />
      ) : (
        <>
          <div className="result-summary">共 {total} 本，当前第 {page} / {totalPages} 页</div>
          <div className="book-grid bookshelf-grid">
            {entries.map((entry) => (
              <article key={entry.id} className={`bookshelf-card ${entry.isPinned ? "pinned" : ""}`}>
                <BookshelfBookMenu
                  entry={entry}
                  onTogglePin={togglePin}
                  onRequestMove={setMovingEntry}
                  onRemove={removeEntry}
                />
                <div className="bookshelf-card-book">
                  <BookCard book={entry.book} isPinned={entry.isPinned} />
                </div>
              </article>
            ))}
          </div>
          <div className="pagination bookshelf-pagination">
            <button className="ghost-button" type="button" onClick={() => updatePage(page - 1, pageSize)} disabled={page <= 1}>
              <ChevronLeft size={16} aria-hidden="true" />
              上一页
            </button>
            <span>{page} / {totalPages}</span>
            <button className="ghost-button" type="button" onClick={() => updatePage(page + 1, pageSize)} disabled={page >= totalPages}>
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
                  updatePage(1, Number(value));
                }}
              >
                {PAGE_SIZE_OPTIONS.map((value) => (
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
                    onClick={() => updatePage(1, clampPageSize(Number(customPageSize) || 20))}
                  >
                    应用
                  </button>
                </>
              ) : null}
            </div>
          </div>
        </>
      )}
      <BookshelfMoveDialog
        entry={movingEntry}
        groups={groups}
        currentGroupId={group.id}
        allowMoveToShelf
        onClose={() => setMovingEntry(null)}
        onMove={moveEntry}
        onCreateGroupAndMove={createGroupAndMove}
      />
      <BookshelfRenameGroupDialog
        group={renamingGroup}
        onClose={() => setRenamingGroup(null)}
        onSubmit={submitRenameGroup}
      />
    </main>
  );
}
