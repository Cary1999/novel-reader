import { UserPlus } from "lucide-react";
import { FormEvent, useState } from "react";
import { Link, Navigate, useLocation, useNavigate } from "react-router-dom";
import { ApiError } from "../api/client";
import { useAuth } from "../auth/AuthContext";

export function RegisterPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const { isAuthenticated, register } = useAuth();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  const from = (location.state as { from?: { pathname?: string } } | null)?.from?.pathname ?? "/";

  if (isAuthenticated) {
    return <Navigate to={from} replace />;
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    const cleanUsername = username.trim();

    if (!cleanUsername) {
      setError("用户名不能为空");
      return;
    }
    if (cleanUsername.length > 40) {
      setError("用户名不能超过 40 个字符");
      return;
    }
    if (password.length < 6) {
      setError("密码至少 6 位");
      return;
    }
    if (password !== confirmPassword) {
      setError("两次输入的密码不一致");
      return;
    }

    try {
      setIsSubmitting(true);
      await register(cleanUsername, password);
      navigate(from, { replace: true });
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "注册失败，请稍后再试");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <main className="page auth-page shell">
      <section className="auth-panel">
        <p className="eyebrow">注册</p>
        <h1>创建阅读账号</h1>
        <form className="form-stack" onSubmit={handleSubmit}>
          <label>
            用户名
            <input value={username} maxLength={40} onChange={(event) => setUsername(event.target.value)} autoComplete="username" />
          </label>
          <label>
            密码
            <input value={password} minLength={6} onChange={(event) => setPassword(event.target.value)} type="password" autoComplete="new-password" />
          </label>
          <label>
            确认密码
            <input value={confirmPassword} minLength={6} onChange={(event) => setConfirmPassword(event.target.value)} type="password" autoComplete="new-password" />
          </label>
          {error ? <p className="form-error">{error}</p> : null}
          <button className="primary-button wide" type="submit" disabled={isSubmitting}>
            <UserPlus size={18} aria-hidden="true" />
            {isSubmitting ? "注册中..." : "注册并登录"}
          </button>
        </form>
        <p className="auth-switch">已有账号？<Link to="/login" state={{ from: location.state }}>去登录</Link></p>
      </section>
    </main>
  );
}
