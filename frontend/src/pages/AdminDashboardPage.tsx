import { LibraryBig, Tags } from "lucide-react";
import { useState } from "react";
import { AdminBooksPage } from "./AdminBooksPage";
import { AdminCategoriesPage } from "./AdminCategoriesPage";

type TabKey = "books" | "categories";

export function AdminDashboardPage() {
  const [tab, setTab] = useState<TabKey>("books");

  return (
    <main className="page shell admin-page">
      <section className="page-banner admin-banner">
        <div>
          <p className="eyebrow">管理员工具</p>
          <h1>管理后台</h1>
          <p className="muted">统一管理小说内容与分类治理。</p>
        </div>
      </section>

      <div className="admin-tabs panel">
        <button
          type="button"
          className={`tab-button ${tab === "books" ? "active" : ""}`}
          onClick={() => setTab("books")}
        >
          <LibraryBig size={16} aria-hidden="true" />
          内容管理
        </button>
        <button
          type="button"
          className={`tab-button ${tab === "categories" ? "active" : ""}`}
          onClick={() => setTab("categories")}
        >
          <Tags size={16} aria-hidden="true" />
          分类治理
        </button>
      </div>

      {tab === "books" ? <AdminBooksPage embedded /> : <AdminCategoriesPage />}
    </main>
  );
}
