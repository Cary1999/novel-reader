import { ChevronLeft, ChevronRight, Pin, PinOff, Plus, RotateCcw, Trash2 } from "lucide-react";
import { FormEvent, useEffect, useMemo, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { apiClient, ApiError } from "../api/client";
import type { BookshelfEntry, BookshelfGroup } from "../api/types";
import { BookCard } from "../components/BookCard";
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

export function BookshelfPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [groups, setGroups] = useState<BookshelfGroup[]>([]);
  const [entries, setEntries] = useState<BookshelfEntry[]>([]);
  const [total, setTotal] = useState(0);
  const [isLoadingGroups, setIsLoadingGroups] = useState(true);
  const [isLoadingEntries, setIsLoadingEntries] = useState(true);
  const [error, setError] = useState("");
  const [selectedIds, setSelectedIds] = useState<number[]>([]);
  const [newGroupName, setNewGroupName] = useState("");
  const [batchTargetGroupId, setBatchTargetGroupId] = useState<number | "">("");
  const [pageSizeMode, setPageSizeMode] = useState("20");
  const [customPageSize, setCustomPageSize] = useState("20");
  const [activeGroupId, setActiveGroupId] = useState<number | null>(null);

  const page = clampPage(parseIntParam(searchParams.get("page"), 1));
  const pageSize = clampPageSize(parseIntParam(searchParams.get("pageSize"), 20));
  const totalPages = Math.max(Math.ceil(total / pageSize), 1);

  const activeGroup = useMemo(
    () => groups.find((group) => group.id === activeGroupId) ?? groups.find((group) => group.isDefault) ?? groups[0] ?? null,
    [activeGroupId, groups],
  );

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
      const nextGroups = response.items ?? [];
      setGroups(nextGroups);
      const fallback = nextGroups.find((item) => item.isDefault) ?? nextGroups[0] ?? null;
      setActiveGroupId((current) => {
        if (current && nextGroups.some((item) => item.id === current)) {
          return current;
        }
        return fallback?.id ?? null;
      });
      if (fallback) {
        setBatchTargetGroupId((current) => (current === "" || nextGroups.some((item) => item.id === current) ? current : fallback.id));
      }
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "书架分组加载失败");
    } finally {
      setIsLoadingGroups(false);
    }
  }

  async function loadEntries() {
    if (!activeGroupId) {
      setEntries([]);
      setTotal(0);
      setIsLoadingEntries(false);
      return;
    }

    try {
      setIsLoadingEntries(true);
      setError("");
      const response = await apiClient.bookshelf({
        groupId: activeGroupId,
        page,
        pageSize,
      });
      setEntries(response.items ?? []);
      setTotal(response.total ?? 0);
      setPageSizeMode(pageSizeCurrentMode);
      setCustomPageSize(String(pageSize));
      setBatchTargetGroupId((current) => (current === "" && groups.length ? groups[0].id : current));
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
  }, [activeGroupId, page, pageSize]);

  useEffect(() => {
    setSelectedIds([]);
  }, [activeGroupId, page, pageSize]);

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
    await loadGroups();
    await loadEntries();
  }

  async function submitNewGroup(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    try {
      const name = newGroupName.trim();
      if (!name) return;
      await apiClient.createBookshelfGroup({ name });
      setNewGroupName("");
      await refreshAll();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "创建分组失败");
    }
  }

  async function renameGroup(group: BookshelfGroup) {
    const nextName = globalThis.prompt("请输入新的分组名称", group.name)?.trim();
    if (!nextName) return;
    try {
      await apiClient.updateBookshelfGroup(group.id, { name: nextName });
      await refreshAll();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "重命名分组失败");
    }
  }

  async function deleteGroup(group: BookshelfGroup) {
    if (!globalThis.confirm(`确定删除分组“${group.name}”吗？分组中的书籍会移回默认分组。`)) {
      return;
    }
    try {
      await apiClient.deleteBookshelfGroup(group.id);
      await refreshAll();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "删除分组失败");
    }
  }

  async function togglePin(entry: BookshelfEntry) {
    try {
      await apiClient.updateBookshelfBook(entry.bookId, { pinned: !entry.isPinned });
      await refreshAll();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "置顶操作失败");
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

  async function batchAction(action: "move" | "pin" | "unpin" | "remove") {
    try {
      const payload = {
        action,
        groupId: action === "move" ? batchTargetGroupId || undefined : undefined,
        bookIds: selectedIds,
      };
      await apiClient.batchManageBookshelf(payload);
      setSelectedIds([]);
      await refreshAll();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "批量管理失败");
    }
  }

  const selectedCount = selectedIds.length;

  return (
    <main className="page shell">
      <section className="page-banner">
        <div>
          <p className="eyebrow">我的书架</p>
          <h1>分组、置顶、批量管理</h1>
          <p className="muted">把常看的小说放进不同分组，重要的书可以置顶，批量操作更省心。</p>
        </div>
      </section>

      {error ? <p className="form-error">{error}</p> : null}

      <section className="panel form-stack">
        <div className="section-heading">
          <div>
            <p className="eyebrow">分组</p>
            <h2>书架分区</h2>
          </div>
          <form className="bookshelf-create-group" onSubmit={submitNewGroup}>
            <input
              value={newGroupName}
              placeholder="新建分组"
              maxLength={64}
              onChange={(event) => setNewGroupName(event.target.value)}
            />
            <button className="primary-button compact" type="submit">
              <Plus size={16} aria-hidden="true" />
              新建
            </button>
          </form>
        </div>

        {isLoadingGroups ? <LoadingState label="正在加载分组..." /> : null}
        {!isLoadingGroups && groups.length === 0 ? <EmptyState title="还没有分组" description="创建一个分组后就可以开始整理书架。" /> : null}

        <div className="bookshelf-group-list">
          {groups.map((group) => (
            <div key={group.id} className={`bookshelf-group-row ${group.id === activeGroup?.id ? "active" : ""}`}>
              <button
                className="bookshelf-group-tab"
                type="button"
                onClick={() => {
                  setActiveGroupId(group.id);
                  setSelectedIds([]);
                  updatePage(1, pageSize);
                }}
              >
                <strong>{group.name}</strong>
                <span>{group.itemCount ?? 0}</span>
              </button>
              <div className="bookshelf-group-actions">
                <button className="ghost-button compact" type="button" onClick={() => void renameGroup(group)}>
                  重命名
                </button>
                {!group.isDefault ? (
                  <button className="ghost-button compact" type="button" onClick={() => void deleteGroup(group)}>
                    删除
                  </button>
                ) : null}
              </div>
            </div>
          ))}
        </div>
      </section>

      <section className="section-heading">
        <div>
          <p className="eyebrow">书架内容</p>
          <h2>{activeGroup?.name ?? "默认分组"}</h2>
        </div>
        <div className="page-size-control">
          <label>
            每页
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
          </label>
          {pageSizeMode === "custom" ? (
            <>
              <label>
                数值
                <input
                  type="number"
                  min={1}
                  max={100}
                  value={customPageSize}
                  onChange={(event) => setCustomPageSize(event.target.value)}
                />
              </label>
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
      </section>

      {selectedCount > 0 ? (
        <section className="panel bookshelf-batch-bar">
          <div className="bookshelf-batch-copy">
            <strong>已选择 {selectedCount} 本书</strong>
            <span className="muted">可以批量移动、置顶或移出书架。</span>
          </div>
          <label>
            目标分组
            <select
              value={batchTargetGroupId}
              onChange={(event) => setBatchTargetGroupId(Number(event.target.value))}
            >
              {groups.map((group) => (
                <option key={group.id} value={group.id}>{group.name}</option>
              ))}
            </select>
          </label>
          <div className="bookshelf-batch-actions">
            <button className="primary-button compact" type="button" onClick={() => void batchAction("move")}>
              <RotateCcw size={16} aria-hidden="true" />
              移动
            </button>
            <button className="ghost-button compact" type="button" onClick={() => void batchAction("pin")}>
              <Pin size={16} aria-hidden="true" />
              置顶
            </button>
            <button className="ghost-button compact" type="button" onClick={() => void batchAction("unpin")}>
              <PinOff size={16} aria-hidden="true" />
              取消置顶
            </button>
            <button className="ghost-button compact danger" type="button" onClick={() => void batchAction("remove")}>
              <Trash2 size={16} aria-hidden="true" />
              移出
            </button>
          </div>
        </section>
      ) : null}

      {isLoadingEntries ? <LoadingState label="正在加载书架..." /> : null}
      {!isLoadingEntries && entries.length === 0 ? (
        <EmptyState
          title="书架还是空的"
          description="去首页发现更多小说，或者在详情页先加入书架。"
        />
      ) : null}

      {!isLoadingEntries && entries.length > 0 ? (
        <>
          <div className="result-summary">共 {total} 本，当前第 {page} / {totalPages} 页</div>
          <div className="book-grid bookshelf-grid">
            {entries.map((entry) => (
              <article key={entry.id} className={`bookshelf-card ${entry.isPinned ? "pinned" : ""}`}>
                <label className="bookshelf-card-check">
                  <input
                    type="checkbox"
                    checked={selectedIds.includes(entry.bookId)}
                    onChange={(event) => {
                      setSelectedIds((current) => (
                        event.target.checked
                          ? Array.from(new Set([...current, entry.bookId]))
                          : current.filter((id) => id !== entry.bookId)
                      ));
                    }}
                  />
                </label>
                <div className="bookshelf-card-book">
                  <BookCard book={entry.book} />
                </div>
                <div className="bookshelf-card-meta">
                  <span className="book-chip">{entry.groupName}</span>
                  {entry.isPinned ? <span className="book-chip pinned-chip">置顶</span> : null}
                  {entry.isDefault ? <span className="book-chip default-chip">默认分组</span> : null}
                </div>
                <div className="bookshelf-card-actions">
                  <button className="ghost-button compact" type="button" onClick={() => void togglePin(entry)}>
                    {entry.isPinned ? "取消置顶" : "置顶"}
                  </button>
                  <button className="ghost-button compact danger" type="button" onClick={() => void removeEntry(entry.bookId)}>
                    移出
                  </button>
                </div>
              </article>
            ))}
          </div>
          <div className="pagination">
            <button className="ghost-button" type="button" onClick={() => updatePage(page - 1, pageSize)} disabled={page <= 1}>
              <ChevronLeft size={16} aria-hidden="true" />
              上一页
            </button>
            <span>{page} / {totalPages}</span>
            <button className="ghost-button" type="button" onClick={() => updatePage(page + 1, pageSize)} disabled={page >= totalPages}>
              下一页
              <ChevronRight size={16} aria-hidden="true" />
            </button>
          </div>
        </>
      ) : null}
    </main>
  );
}
