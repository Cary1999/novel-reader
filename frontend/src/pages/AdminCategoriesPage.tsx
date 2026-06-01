import { Check, Pencil, Plus, Trash2, X } from "lucide-react";
import { FormEvent, useEffect, useMemo, useState } from "react";
import { apiClient, ApiError } from "../api/client";
import type { Category } from "../api/types";
import { EmptyState, ErrorState, LoadingState } from "../components/StateViews";

export function AdminCategoriesPage() {
  const [categories, setCategories] = useState<Category[]>([]);
  const [name, setName] = useState("");
  const [query, setQuery] = useState("");
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editingName, setEditingName] = useState("");
  const [isLoading, setIsLoading] = useState(true);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  const filteredCategories = useMemo(() => {
    const keyword = query.trim();
    if (!keyword) return categories;
    return categories.filter((item) => item.name.includes(keyword));
  }, [categories, query]);

  async function load() {
    try {
      setIsLoading(true);
      setError("");
      const response = await apiClient.categories();
      setCategories(response.items ?? []);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "分类加载失败");
    } finally {
      setIsLoading(false);
    }
  }

  async function create(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    try {
      setError("");
      if (!name.trim()) {
        setError("分类名不能为空");
        return;
      }
      await apiClient.createCategory(name.trim());
      setName("");
      setMessage("分类已新增");
      await load();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "新增分类失败");
    }
  }

  async function saveEdit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!editingId) return;
    try {
      setError("");
      if (!editingName.trim()) {
        setError("分类名不能为空");
        return;
      }
      await apiClient.updateCategory(editingId, editingName.trim());
      setEditingId(null);
      setEditingName("");
      setMessage("分类已保存");
      await load();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "保存分类失败");
    }
  }

  async function remove(category: Category) {
    if (!window.confirm(`确定删除分类“${category.name}”？`)) return;
    try {
      setError("");
      await apiClient.deleteCategory(category.id);
      setMessage("分类已删除");
      await load();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "删除分类失败");
    }
  }

  useEffect(() => {
    void load();
  }, []);

  return (
    <main className="page shell admin-page category-page">
      <section className="page-banner admin-banner">
        <div>
          <p className="eyebrow">管理员工具</p>
          <h1>分类管理</h1>
          <p className="muted">小说只能选择已有分类，分类由管理员统一维护。</p>
        </div>
      </section>

      {message ? <p className="success-banner">{message}</p> : null}
      {error ? <ErrorState message={error} onRetry={load} /> : null}

      <section className="category-grid">
        <form className="panel category-create" onSubmit={create}>
          <div className="category-create-head">
            <p className="eyebrow">新增分类</p>
            <span className="muted">用于书库筛选与小说表单选择</span>
          </div>
          <div className="category-create-row">
            <input
              aria-label="分类名"
              value={name}
              maxLength={64}
              onChange={(event) => setName(event.target.value)}
              placeholder="输入分类名，例如：玄幻、都市、历史"
            />
            <button className="primary-button compact icon-only" type="submit" title="新增分类" aria-label="新增分类">
              <Plus size={18} aria-hidden="true" />
            </button>
          </div>
          <p className="muted category-hint">建议：尽量保持分类简短、稳定，避免同义重复。</p>
        </form>

        <div className="panel category-list-panel">
          <div className="category-list-head">
            <div>
              <p className="eyebrow">分类列表</p>
              <h2 className="category-subtitle">{categories.length} 个分类</h2>
            </div>
            <input
              className="category-search"
              value={query}
              maxLength={40}
              onChange={(event) => setQuery(event.target.value)}
              placeholder="筛选分类"
              aria-label="筛选分类"
            />
          </div>

          {isLoading ? <LoadingState label="正在加载分类..." /> : null}
          {!isLoading && categories.length === 0 ? <EmptyState title="暂无分类" description="新增分类后，小说表单可以选择它。" /> : null}
          {!isLoading && categories.length > 0 ? (
            <div className="category-rows" role="list">
              {filteredCategories.map((category) => (
                <div className="category-row" role="listitem" key={category.id}>
                  {editingId === category.id ? (
                    <form className="category-inline-edit" onSubmit={saveEdit}>
                      <input
                        value={editingName}
                        maxLength={64}
                        onChange={(event) => setEditingName(event.target.value)}
                        aria-label="编辑分类名"
                        autoFocus
                      />
                      <div className="category-actions">
                        <button className="ghost-button compact icon-only" type="submit" title="保存" aria-label="保存">
                          <Check size={18} aria-hidden="true" />
                        </button>
                        <button
                          className="ghost-button compact icon-only"
                          type="button"
                          title="取消"
                          aria-label="取消"
                          onClick={() => {
                            setEditingId(null);
                            setEditingName("");
                          }}
                        >
                          <X size={18} aria-hidden="true" />
                        </button>
                      </div>
                    </form>
                  ) : (
                    <>
                      <strong className="category-name">{category.name}</strong>
                      <div className="category-actions">
                        <button
                          className="ghost-button compact icon-only"
                          type="button"
                          title="重命名"
                          aria-label="重命名"
                          onClick={() => {
                            setEditingId(category.id);
                            setEditingName(category.name);
                          }}
                        >
                          <Pencil size={18} aria-hidden="true" />
                        </button>
                        <button
                          className="ghost-button compact danger-button icon-only"
                          type="button"
                          title="删除"
                          aria-label="删除"
                          onClick={() => remove(category)}
                        >
                          <Trash2 size={18} aria-hidden="true" />
                        </button>
                      </div>
                    </>
                  )}
                </div>
              ))}
              {!filteredCategories.length ? (
                <EmptyState title="没有匹配的分类" description="换个关键词再试试。" />
              ) : null}
            </div>
          ) : null}
        </div>
      </section>
    </main>
  );
}
