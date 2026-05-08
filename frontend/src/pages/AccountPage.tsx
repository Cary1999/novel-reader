import { FormEvent, useState } from "react";
import { apiClient, ApiError } from "../api/client";
import { useAuth } from "../auth/AuthContext";

export function AccountPage() {
  const { user, refreshUser } = useAuth();
  const [nickname, setNickname] = useState(user?.nickname ?? user?.username ?? "");
  const [oldPassword, setOldPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [isSavingProfile, setIsSavingProfile] = useState(false);
  const [isSavingPassword, setIsSavingPassword] = useState(false);

  async function saveProfile(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    try {
      setIsSavingProfile(true);
      setError("");
      await apiClient.updateMe(nickname.trim());
      await refreshUser();
      setMessage("昵称已保存");
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "昵称保存失败");
    } finally {
      setIsSavingProfile(false);
    }
  }

  async function savePassword(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    try {
      setIsSavingPassword(true);
      setError("");
      await apiClient.changePassword(oldPassword, newPassword);
      setOldPassword("");
      setNewPassword("");
      setMessage("密码已修改");
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "密码修改失败");
    } finally {
      setIsSavingPassword(false);
    }
  }

  return (
    <main className="page shell admin-page">
      <section className="admin-header">
        <div>
          <p className="eyebrow">账号</p>
          <h1>账号设置</h1>
          <p className="muted">维护昵称和登录密码。用户名作为作者引用保持不变。</p>
        </div>
      </section>

      {message ? <p className="success-banner">{message}</p> : null}
      {error ? <p className="form-error">{error}</p> : null}

      <section className="settings-grid">
        <form className="panel form-stack" onSubmit={saveProfile}>
          <p className="eyebrow">资料</p>
          <label>
            用户名
            <input value={user?.username ?? ""} disabled />
          </label>
          <label>
            昵称
            <input value={nickname} maxLength={64} onChange={(event) => setNickname(event.target.value)} />
          </label>
          <label>
            角色
            <input value={user?.role === "admin" ? "管理员" : "普通用户"} disabled />
          </label>
          <button className="primary-button compact" type="submit" disabled={isSavingProfile}>
            {isSavingProfile ? "保存中..." : "保存昵称"}
          </button>
        </form>

        <form className="panel form-stack" onSubmit={savePassword}>
          <p className="eyebrow">安全</p>
          <label>
            当前密码
            <input type="password" value={oldPassword} onChange={(event) => setOldPassword(event.target.value)} />
          </label>
          <label>
            新密码
            <input type="password" value={newPassword} onChange={(event) => setNewPassword(event.target.value)} />
          </label>
          <button className="primary-button compact" type="submit" disabled={isSavingPassword}>
            {isSavingPassword ? "修改中..." : "修改密码"}
          </button>
        </form>
      </section>
    </main>
  );
}
