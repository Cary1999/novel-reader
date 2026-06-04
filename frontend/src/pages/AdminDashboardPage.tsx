import { CheckCircle2, Settings2, ShieldCheck, Sparkles, Tags, Users } from "lucide-react";
import { FormEvent, useEffect, useState } from "react";
import { apiClient, ApiError } from "../api/client";
import { useAuth } from "../auth/AuthContext";
import type { AuthorApplication, BookSummary, CurrentOperator, FrontUserSummary } from "../api/types";
import { EmptyState, ErrorState, LoadingState } from "../components/StateViews";
import { AdminCategoriesPage } from "./AdminCategoriesPage";
import { AdminSiteSettingsPage } from "./AdminSiteSettingsPage";

type TabKey = "applications" | "users" | "operators" | "recommendations" | "categories" | "settings";

export function AdminDashboardPage() {
  const [tab, setTab] = useState<TabKey>("applications");

  return (
    <main className="page shell admin-page">
      <section className="page-banner admin-banner">
        <div>
          <p className="eyebrow">运营后台</p>
          <h1>审核与治理后台</h1>
          <p className="muted">后台只负责作者申请审核、读者升作者、运营人员管理、分类治理和站点设置，不再直接编辑小说内容。</p>
        </div>
      </section>

      <div className="admin-tabs panel">
        <button type="button" className={`tab-button ${tab === "applications" ? "active" : ""}`} onClick={() => setTab("applications")}>
          <CheckCircle2 size={16} aria-hidden="true" />
          作者申请
        </button>
        <button type="button" className={`tab-button ${tab === "users" ? "active" : ""}`} onClick={() => setTab("users")}>
          <Users size={16} aria-hidden="true" />
          用户列表
        </button>
        <button type="button" className={`tab-button ${tab === "operators" ? "active" : ""}`} onClick={() => setTab("operators")}>
          <ShieldCheck size={16} aria-hidden="true" />
          运营人员
        </button>
        <button type="button" className={`tab-button ${tab === "recommendations" ? "active" : ""}`} onClick={() => setTab("recommendations")}>
          <Sparkles size={16} aria-hidden="true" />
          推荐值
        </button>
        <button type="button" className={`tab-button ${tab === "categories" ? "active" : ""}`} onClick={() => setTab("categories")}>
          <Tags size={16} aria-hidden="true" />
          分类治理
        </button>
        <button type="button" className={`tab-button ${tab === "settings" ? "active" : ""}`} onClick={() => setTab("settings")}>
          <Settings2 size={16} aria-hidden="true" />
          站点设置
        </button>
      </div>

      {tab === "applications" ? <ApplicationsPanel /> : null}
      {tab === "users" ? <UsersPanel /> : null}
      {tab === "operators" ? <OperatorsPanel /> : null}
      {tab === "recommendations" ? <RecommendationsPanel /> : null}
      {tab === "categories" ? <AdminCategoriesPage /> : null}
      {tab === "settings" ? <AdminSiteSettingsPage /> : null}
    </main>
  );
}

function RecommendationsPanel() {
  const [items, setItems] = useState<BookSummary[]>([]);
  const [query, setQuery] = useState("");
  const [scores, setScores] = useState<Record<number, string>>({});
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");

  async function load(nextQuery = query) {
    try {
      setIsLoading(true);
      setError("");
      const response = await apiClient.adminBooks({ q: nextQuery.trim(), page: 1, pageSize: 20 });
      const nextItems = response.items ?? [];
      setItems(nextItems);
      setScores(Object.fromEntries(nextItems.map((item) => [item.id, String(item.recommendScore ?? 0)])));
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "加载书籍列表失败");
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    void load("");
  }, []);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await load(query);
  }

  async function save(book: BookSummary) {
    try {
      setError("");
      setMessage("");
      const value = Number(scores[book.id] ?? 0);
      await apiClient.updateRecommendScore(book.id, value);
      setMessage(`《${book.title}》推荐值已更新为 ${value}`);
      await load(query);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "更新推荐值失败");
    }
  }

  return (
    <section className="panel form-stack">
      <div className="section-heading">
        <div>
          <p className="eyebrow">书籍治理</p>
          <h2>推荐值调整</h2>
        </div>
      </div>
      {message ? <p className="success-banner">{message}</p> : null}
      {error ? <ErrorState message={error} onRetry={() => load(query)} /> : null}
      <form className="admin-search" onSubmit={submit}>
        <input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="搜索书名或简介" />
        <button className="primary-button compact" type="submit">搜索</button>
      </form>
      {isLoading ? <LoadingState label="正在加载书籍..." /> : null}
      {!isLoading && items.length === 0 ? <EmptyState title="暂无书籍" description="当前没有可调整推荐值的书籍。" /> : null}
      {!isLoading && items.length > 0 ? (
        <div className="admin-books">
          {items.map((item) => (
            <div key={item.id} className="admin-book-row active">
              <span>
                <strong>{item.title}</strong>
                <small>{item.author} · {item.category || "未分类"} · {item.chapterCount} 章</small>
              </span>
              <div className="detail-actions">
                <input
                  style={{ width: 96 }}
                  type="number"
                  min={0}
                  step={1}
                  value={scores[item.id] ?? "0"}
                  onChange={(event) => setScores((current) => ({ ...current, [item.id]: event.target.value }))}
                />
                <button className="primary-button compact" type="button" onClick={() => save(item)}>
                  保存
                </button>
              </div>
            </div>
          ))}
        </div>
      ) : null}
    </section>
  );
}

function ApplicationsPanel() {
  const [items, setItems] = useState<AuthorApplication[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");

  async function load() {
    try {
      setIsLoading(true);
      setError("");
      const response = await apiClient.listAuthorApplications();
      setItems(response.items ?? []);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "加载申请记录失败");
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, []);

  async function review(applicationId: number, decision: "approved" | "rejected") {
    try {
      setMessage("");
      await apiClient.reviewAuthorApplication(applicationId, decision);
      setMessage(decision === "approved" ? "申请已通过，读者已升级为作者" : "申请已拒绝");
      await load();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "审核失败");
    }
  }

  return (
    <section className="panel form-stack">
      <div className="section-heading">
        <div>
          <p className="eyebrow">审核队列</p>
          <h2>作者申请记录</h2>
        </div>
      </div>
      {message ? <p className="success-banner">{message}</p> : null}
      {error ? <ErrorState message={error} onRetry={load} /> : null}
      {isLoading ? <LoadingState label="正在加载申请..." /> : null}
      {!isLoading && items.length === 0 ? <EmptyState title="暂无申请" description="读者提交作者申请后会出现在这里。" /> : null}
      {!isLoading && items.length > 0 ? (
        <div className="admin-books">
          {items.map((item) => (
            <div key={item.id} className="admin-book-row active">
              <span>
                <strong>{item.penName} / {item.username}</strong>
                <small>{item.status} · {item.reason}</small>
              </span>
              {item.status === "pending" ? (
                <div className="detail-actions">
                  <button className="primary-button compact" type="button" onClick={() => review(item.id, "approved")}>通过</button>
                  <button className="ghost-button compact" type="button" onClick={() => review(item.id, "rejected")}>拒绝</button>
                </div>
              ) : null}
            </div>
          ))}
        </div>
      ) : null}
    </section>
  );
}

function UsersPanel() {
  const [items, setItems] = useState<FrontUserSummary[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");

  async function load() {
    try {
      setIsLoading(true);
      setError("");
      const response = await apiClient.listUsers();
      setItems(response.items ?? []);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "加载用户失败");
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, []);

  async function promote(userId: number) {
    try {
      await apiClient.promoteUserToAuthor(userId);
      setMessage("已直接将读者提升为作者");
      await load();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "提升作者失败");
    }
  }

  return (
    <section className="panel form-stack">
      <div className="section-heading">
        <div>
          <p className="eyebrow">用户治理</p>
          <h2>前台用户列表</h2>
        </div>
      </div>
      {message ? <p className="success-banner">{message}</p> : null}
      {error ? <ErrorState message={error} onRetry={load} /> : null}
      {isLoading ? <LoadingState label="正在加载用户..." /> : null}
      {!isLoading && items.length === 0 ? <EmptyState title="暂无用户" description="注册用户会出现在这里。" /> : null}
      {!isLoading && items.length > 0 ? (
        <div className="admin-books">
          {items.map((item) => (
            <div key={item.id} className="admin-book-row active">
              <span>
                <strong>{item.nickname} / {item.username}</strong>
                <small>{item.role} · 申请状态 {item.latestApplicationStatus || "none"}</small>
              </span>
              {item.role === "reader" ? (
                <button className="primary-button compact" type="button" onClick={() => promote(item.id)}>升为作者</button>
              ) : null}
            </div>
          ))}
        </div>
      ) : null}
    </section>
  );
}

function OperatorsPanel() {
  const { isSuperAdmin } = useAuth();
  const [items, setItems] = useState<CurrentOperator[]>([]);
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [role, setRole] = useState<"reviewer" | "super_admin">("reviewer");
  const [isLoading, setIsLoading] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");

  async function load() {
    try {
      setIsLoading(true);
      setError("");
      const response = await apiClient.listOperators();
      setItems(response.items ?? []);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "加载运营人员失败");
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, []);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    try {
      setIsSubmitting(true);
      setError("");
      await apiClient.createOperator(username.trim(), password, role);
      setUsername("");
      setPassword("");
      setRole("reviewer");
      setMessage("运营人员已创建");
      await load();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "创建运营人员失败");
    } finally {
      setIsSubmitting(false);
    }
  }

  async function changeRole(operatorId: number, nextRole: "reviewer" | "super_admin") {
    try {
      await apiClient.updateOperator(operatorId, nextRole);
      setMessage("运营人员角色已更新");
      await load();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "更新角色失败");
    }
  }

  return (
    <section className="panel form-stack">
      <div className="section-heading">
        <div>
          <p className="eyebrow">后台账号</p>
          <h2>运营人员管理</h2>
        </div>
      </div>
      {message ? <p className="success-banner">{message}</p> : null}
      {error ? <ErrorState message={error} onRetry={load} /> : null}
      {isLoading ? <LoadingState label="正在加载运营人员..." /> : null}
      {!isLoading && items.length > 0 ? (
        <div className="admin-books">
          {items.map((item) => (
            <div key={item.id} className="admin-book-row active">
              <span>
                <strong>{item.username}</strong>
                <small>{item.role}</small>
              </span>
              {isSuperAdmin ? (
                <button
                  className="ghost-button compact"
                  type="button"
                  onClick={() => changeRole(item.id, item.role === "super_admin" ? "reviewer" : "super_admin")}
                >
                  切换角色
                </button>
              ) : null}
            </div>
          ))}
        </div>
      ) : null}

      {isSuperAdmin ? (
        <form className="form-stack" onSubmit={submit}>
          <label>
            用户名
            <input value={username} onChange={(event) => setUsername(event.target.value)} />
          </label>
          <label>
            初始密码
            <input type="password" value={password} onChange={(event) => setPassword(event.target.value)} />
          </label>
          <label>
            角色
            <select value={role} onChange={(event) => setRole(event.target.value as "reviewer" | "super_admin")}>
              <option value="reviewer">reviewer</option>
              <option value="super_admin">super_admin</option>
            </select>
          </label>
          <button className="primary-button compact" type="submit" disabled={isSubmitting}>
            {isSubmitting ? "创建中..." : "创建后台账号"}
          </button>
        </form>
      ) : (
        <p className="muted">只有超级管理员可以创建或调整后台运营人员。</p>
      )}
    </section>
  );
}
