import { ChevronLeft, ChevronRight, FolderOpen, Pin } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { apiClient, ApiError } from "../api/client";
import type { BookshelfEntry, BookshelfGroup } from "../api/types";
import { BookCard } from "../components/BookCard";
import {
  BookshelfBookMenu,
  BookshelfGroupMenu,
  BookshelfMoveDialog,
  BookshelfRenameGroupDialog,
} from "../components/BookshelfMenus";
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

function timestampValue(value?: string) {
  if (!value) return 0;
  const parsed = Date.parse(value);
  return Number.isFinite(parsed) ? parsed : 0;
}

export function BookshelfPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [groups, setGroups] = useState<BookshelfGroup[]>([]);
  const [entries, setEntries] = useState<BookshelfEntry[]>([]);
  const [total, setTotal] = useState(0);
  const [isLoadingGroups, setIsLoadingGroups] = useState(true);
  const [isLoadingEntries, setIsLoadingEntries] = useState(true);
  const [error, setError] = useState("");
  const [pageSizeMode, setPageSizeMode] = useState("20");
  const [customPageSize, setCustomPageSize] = useState("20");
  const [movingEntry, setMovingEntry] = useState<BookshelfEntry | null>(null);
  const [renamingGroup, setRenamingGroup] = useState<BookshelfGroup | null>(null);

  const page = clampPage(parseIntParam(searchParams.get("page"), 1));
  const pageSize = clampPageSize(parseIntParam(searchParams.get("pageSize"), 20));
  const totalPages = Math.max(Math.ceil(total / pageSize), 1);
  const visibleGroups = useMemo(() => groups.filter((group) => (group.itemCount ?? 0) > 0), [groups]);
  const mixedItems = useMemo(() => {
    const items = [
      ...visibleGroups.map((group) => ({ kind: "group" as const, id: `group-${group.id}`, group })),
      ...entries.map((entry) => ({ kind: "entry" as const, id: `entry-${entry.id}`, entry })),
    ];
    items.sort((left, right) => {
      const leftPinned = left.kind === "group" ? left.group.isPinned : left.entry.isPinned;
      const rightPinned = right.kind === "group" ? right.group.isPinned : right.entry.isPinned;
      if (leftPinned !== rightPinned) {
        return leftPinned ? -1 : 1;
      }
      const leftTime = left.kind === "group"
        ? timestampValue(left.group.pinnedAt) || timestampValue(left.group.createdAt)
        : timestampValue(left.entry.pinnedAt) || timestampValue(left.entry.createdAt);
      const rightTime = right.kind === "group"
        ? timestampValue(right.group.pinnedAt) || timestampValue(right.group.createdAt)
        : timestampValue(right.entry.pinnedAt) || timestampValue(right.entry.createdAt);
      if (leftTime !== rightTime) {
        return rightTime - leftTime;
      }
      if (left.kind !== right.kind) {
        return left.kind === "group" ? -1 : 1;
      }
      return 0;
    });
    return items;
  }, [entries, visibleGroups]);

  const pageSizeCurrentMode = useMemo(() => {
    if (PAGE_SIZE_OPTIONS.includes(pageSize as (typeof PAGE_SIZE_OPTIONS)[number])) {
      return String(pageSize);
    }
    return "custom";
  }, [pageSize]);

  async function loadGroups() {
    try {
      setIsLoadingGroups(true);
      setError("");
      const response = await apiClient.bookshelfGroups();
      setGroups(response.items ?? []);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "书架分组加载失败");
    } finally {
      setIsLoadingGroups(false);
    }
  }

  async function loadEntries() {
    try {
      setIsLoadingEntries(true);
      setError("");
      const response = await apiClient.bookshelf({
        page,
        pageSize,
      });
      setEntries(response.items ?? []);
      setTotal(response.total ?? 0);
      setPageSizeMode(pageSizeCurrentMode);
      setCustomPageSize(String(pageSize));
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "书架加载失败");
      setEntries([]);
      setTotal(0);
    } finally {
      setIsLoadingEntries(false);
    }
  }

  useEffect(() => {
    void loadGroups();
  }, []);

  useEffect(() => {
    void loadEntries();
  }, [page, pageSize]);

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

  async function refreshAll() {
    await Promise.all([loadGroups(), loadEntries()]);
  }

  async function renameGroup(group: BookshelfGroup) {
    setRenamingGroup(group);
  }

  async function submitRenameGroup(group: BookshelfGroup, nextName: string) {
    try {
      await apiClient.updateBookshelfGroup(group.id, { name: nextName });
      setRenamingGroup(null);
      await loadGroups();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "重命名分组失败");
    }
  }

  async function deleteGroup(group: BookshelfGroup) {
    if (!globalThis.confirm(`确定删除分组“${group.name}”吗？组内书籍会回到书架首页。`)) {
      return;
    }
    try {
      await apiClient.deleteBookshelfGroup(group.id);
      await refreshAll();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "删除分组失败");
    }
  }

  async function toggleGroupPin(group: BookshelfGroup) {
    try {
      await apiClient.updateBookshelfGroup(group.id, { pinned: !group.isPinned });
      await loadGroups();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "分组置顶失败");
    }
  }

  async function togglePin(entry: BookshelfEntry) {
    try {
      await apiClient.updateBookshelfBook(entry.bookId, { pinned: !entry.isPinned });
      await loadEntries();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "置顶操作失败");
    }
  }

  async function moveEntry(bookId: number, groupId: number) {
    try {
      await apiClient.updateBookshelfBook(bookId, { groupId });
      await refreshAll();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "移动书籍失败");
    }
  }

  async function createGroupAndMove(bookId: number, groupName: string) {
    try {
      const created = await apiClient.createBookshelfGroup({ name: groupName });
      await apiClient.updateBookshelfBook(bookId, { groupId: created.id });
      await refreshAll();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "创建分组失败");
    }
  }

  async function removeEntry(bookId: number) {
    try {
      await apiClient.removeFromBookshelf(bookId);
      await refreshAll();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "移出书架失败");
    }
  }

  const hasAnyContent = visibleGroups.length > 0 || entries.length > 0;

  return (
    <main className="page shell">
      <section className="page-banner">
        <div>
          <p className="eyebrow">我的书架</p>
          <h1>书和分组放在一起看</h1>
          <p className="muted">未分组的书直接躺在书架里，收进分组的书会折叠成同级入口，点进去再继续整理。</p>
        </div>
      </section>

      {error && hasAnyContent ? <p className="form-error">{error}</p> : null}

      <section className="section-heading">
        <div>
          <p className="eyebrow">书架首页</p>
          <h2>未分组书籍与分组入口</h2>
        </div>
      </section>

      {isLoadingGroups || isLoadingEntries ? <LoadingState label="正在加载书架..." /> : null}
      {!isLoadingGroups && !isLoadingEntries && error && !hasAnyContent ? (
        <ErrorState message={error} onRetry={refreshAll} />
      ) : null}
      {!isLoadingGroups && !isLoadingEntries && !error && !hasAnyContent ? (
        <EmptyState
          title="书架还是空的"
          description="去首页发现更多小说，或者先把想收纳的书加入书架。"
        />
      ) : null}

      {!isLoadingGroups && !isLoadingEntries && !error && hasAnyContent ? (
        <>
          <div className="book-grid bookshelf-mixed-grid">
            {mixedItems.map((item) => (
              item.kind === "group" ? (
                <article key={item.id} className={`bookshelf-group-shell ${item.group.isPinned ? "pinned" : ""}`}>
                  <BookshelfGroupMenu
                    group={item.group}
                    onTogglePin={toggleGroupPin}
                    onRename={renameGroup}
                    onDelete={deleteGroup}
                  />
                  <Link to={`/bookshelf/groups/${item.group.id}`} className="bookshelf-group-card">
                    <div className="bookshelf-group-art" aria-hidden="true">
                      <span />
                      <span />
                      <span />
                    </div>
                    <div className="bookshelf-group-body">
                      <span className="book-card-count book-card-count-floating">{item.group.itemCount ?? 0} 本</span>
                      <div className="book-title-row">
                        {item.group.isPinned ? (
                          <span className="book-pin-inline" aria-label="已置顶" title="已置顶">
                            <Pin size={14} aria-hidden="true" />
                          </span>
                        ) : null}
                        <h3>{item.group.name}</h3>
                      </div>
                      <p className="description">把同一类书收进这里，书架首页就会更清爽。</p>
                      <div className="book-card-tag-row">
                        <span className="book-chip">分组</span>
                      </div>
                      <div className="bookshelf-group-footer">
                        <span>进入分组</span>
                        <ChevronRight size={16} aria-hidden="true" />
                      </div>
                    </div>
                  </Link>
                </article>
              ) : (
                <article key={item.id} className={`bookshelf-card ${item.entry.isPinned ? "pinned" : ""}`}>
                  <BookshelfBookMenu
                    entry={item.entry}
                    onTogglePin={togglePin}
                    onRequestMove={setMovingEntry}
                    onRemove={removeEntry}
                  />
                  <div className="bookshelf-card-book">
                    <BookCard book={item.entry.book} isPinned={item.entry.isPinned} />
                  </div>
                </article>
              )
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
      ) : null}

      {!isLoadingGroups && !isLoadingEntries && !error && visibleGroups.length === 0 && entries.length > 0 ? (
        <section className="panel ambient-panel bookshelf-tip-panel">
          <FolderOpen size={18} aria-hidden="true" />
          <p className="muted">现在所有书都直接放在书架首页。通过卡片右上角菜单把书移进新分组，这里就会折叠出新的分组入口。</p>
        </section>
      ) : null}
      <BookshelfMoveDialog
        entry={movingEntry}
        groups={groups}
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
