import { Link } from "react-router-dom";

export function NotFoundPage() {
  return (
    <main className="page shell">
      <section className="panel permission-panel">
        <p className="eyebrow">404</p>
        <h1>页面不存在</h1>
        <p className="muted">请回到发现页或重新搜索小说。</p>
        <Link className="primary-button compact" to="/">返回发现</Link>
      </section>
    </main>
  );
}
