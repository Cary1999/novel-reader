import { LogIn } from "lucide-react";
import { FormEvent, useState } from "react";
import { Link, Navigate, useLocation, useNavigate } from "react-router-dom";
import { ApiError } from "../api/client";
import { useAuth } from "../auth/AuthContext";

export function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const { isAuthenticated, login } = useAuth();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  const from = (location.state as { from?: { pathname?: string } } | null)?.from?.pathname ?? "/";

  if (isAuthenticated) {
    return <Navigate to={from} replace />;
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");

    if (!username.trim() || !password) {
      setError("请输入用户名和密码");
      return;
    }

    try {
      setIsSubmitting(true);
      await login(username.trim(), password);
      navigate(from, { replace: true });
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "登录失败，请稍后再试");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <main className="page auth-page shell">
      <section className="auth-shell">
        <div className="auth-aside">
          <p className="eyebrow">登录</p>
          <h1>继续你的阅读进度</h1>
          <p className="muted">登录后可以阅读章节正文、上传 txt 小说，并维护自己的作品内容。</p>
        </div>
        <section className="auth-panel">
          <form className="form-stack" onSubmit={handleSubmit}>
            <label>
              用户名
              <input value={username} maxLength={40} onChange={(event) => setUsername(event.target.value)} autoComplete="username" />
            </label>
            <label>
              密码
              <input value={password} minLength={6} onChange={(event) => setPassword(event.target.value)} type="password" autoComplete="current-password" />
            </label>
            {error ? <p className="form-error">{error}</p> : null}
            <button className="primary-button wide" type="submit" disabled={isSubmitting}>
              <LogIn size={18} aria-hidden="true" />
              {isSubmitting ? "登录中..." : "登录"}
            </button>
          </form>
          <p className="auth-switch">还没有账号？<Link to="/register" state={{ from: location.state }}>去注册</Link></p>
        </section>
      </section>
    </main>
  );
}
