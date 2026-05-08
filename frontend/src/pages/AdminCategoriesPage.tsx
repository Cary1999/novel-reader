import { FormEvent, useEffect, useState } from "react";
import { apiClient, ApiError } from "../api/client";
import type { Category } from "../api/types";
import { EmptyState, ErrorState, LoadingState } from "../components/StateViews";

export function AdminCategoriesPage() {
  const [categories, setCategories] = useState<Category[]>([]);
  const [name, setName] = useState("");
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editingName, setEditingName] = useState("");
  const [isLoading, setIsLoading] = useState(true);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

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
      await apiClient.createCategory(name);
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
      await apiClient.updateCategory(editingId, editingName);
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
    <main className="page shell admin-page">
      <section className="admin-header">
        <div>
          <p className="eyebrow">管理员工具</p>
          <h1>分类管理</h1>
          <p className="muted">小说只能选择已有分类，分类由管理员统一维护。</p>
        </div>
      </section>

      {message ? <p className="success-banner">{message}</p> : null}
      {error ? <ErrorState message={error} onRetry={load} /> : null}

      <section className="settings-grid">
        <form className="panel form-stack" onSubmit={create}>
          <p className="eyebrow">新增分类</p>
          <label>
            分类名
            <input value={name} maxLength={64} onChange={(event) => setName(event.target.value)} />
          </label>
          <button className="primary-button compact" type="submit">新增</button>
        </form>

        <div className="panel form-stack">
          <p className="eyebrow">分类列表</p>
          {isLoading ? <LoadingState label="正在加载分类..." /> : null}
          {!isLoading && categories.length === 0 ? <EmptyState title="暂无分类" description="新增分类后，小说表单可以选择它。" /> : null}
          {categories.map((category) => (
            <div className="category-admin-row" key={category.id}>
              {editingId === category.id ? (
                <form className="inline-edit" onSubmit={saveEdit}>
                  <input value={editingName} onChange={(event) => setEditingName(event.target.value)} />
                  <button className="primary-button compact" type="submit">保存</button>
                </form>
              ) : (
                <>
                  <strong>{category.name}</strong>
                  <div className="detail-actions">
                    <button className="ghost-button compact" type="button" onClick={() => {
                      setEditingId(category.id);
                      setEditingName(category.name);
                    }}>重命名</button>
                    <button className="ghost-button compact danger-button" type="button" onClick={() => remove(category)}>删除</button>
                  </div>
                </>
              )}
            </div>
          ))}
        </div>
      </section>
    </main>
  );
}
