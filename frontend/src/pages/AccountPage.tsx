import { FormEvent, useEffect, useState } from "react";
import { apiClient, ApiError } from "../api/client";
import { useAuth } from "../auth/AuthContext";
import type { AuthorApplication } from "../api/types";

export function AccountPage() {
  const { isAuthor, isReader, user, refreshUser } = useAuth();
  const frontUser = user && "nickname" in user ? user : null;
  const [nickname, setNickname] = useState(frontUser?.nickname ?? frontUser?.username ?? "");
  const [avatarFile, setAvatarFile] = useState<File | null>(null);
  const [oldPassword, setOldPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [penName, setPenName] = useState("");
  const [reason, setReason] = useState("");
  const [application, setApplication] = useState<AuthorApplication | null>(null);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [isSavingAvatar, setIsSavingAvatar] = useState(false);
  const [isSavingProfile, setIsSavingProfile] = useState(false);
  const [isSavingPassword, setIsSavingPassword] = useState(false);
  const [isSubmittingApplication, setIsSubmittingApplication] = useState(false);

  useEffect(() => {
    setNickname(frontUser?.nickname ?? frontUser?.username ?? "");
  }, [frontUser]);

  useEffect(() => {
    if (!isReader) {
      setApplication(null);
      return;
    }
    void (async () => {
      try {
        setApplication(await apiClient.myAuthorApplication());
      } catch (err) {
        if (err instanceof ApiError && err.status === 404) {
          setApplication(null);
          return;
        }
      }
    })();
  }, [isReader]);

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

  async function saveAvatar(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!avatarFile) {
      setError("请选择头像文件");
      return;
    }
    try {
      setIsSavingAvatar(true);
      setError("");
      await apiClient.uploadAvatar(avatarFile);
      await refreshUser();
      setAvatarFile(null);
      setMessage("头像已更新");
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "头像上传失败");
    } finally {
      setIsSavingAvatar(false);
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

  async function submitApplication(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    try {
      setIsSubmittingApplication(true);
      setError("");
      const nextApplication = await apiClient.submitAuthorApplication(penName.trim(), reason.trim());
      setApplication(nextApplication);
      setPenName("");
      setReason("");
      setMessage("作者申请已提交，请等待后台审核");
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "申请提交失败");
    } finally {
      setIsSubmittingApplication(false);
    }
  }

  const roleLabel = isAuthor ? "作者" : "读者";
  const applicationStatusLabel = application?.status === "approved"
    ? "已通过"
    : application?.status === "rejected"
      ? "已拒绝"
      : application?.status === "pending"
        ? "审核中"
        : "未申请";

  return (
    <main className="page shell admin-page">
      <section className="page-banner admin-banner">
        <div>
          <p className="eyebrow">账号</p>
          <h1>账号设置</h1>
          <p className="muted">维护昵称和登录密码。用户名作为作者引用保持不变。</p>
        </div>
      </section>

      {message ? <p className="success-banner">{message}</p> : null}
      {error ? <p className="form-error">{error}</p> : null}

      <section className="settings-grid">
        <form className="panel form-stack" onSubmit={saveAvatar}>
          <p className="eyebrow">头像</p>
          <div className="avatar-preview">
            {frontUser?.avatarUrl ? <img src={frontUser.avatarUrl} alt="" /> : null}
          </div>
          <label>
            上传头像
            <input type="file" accept=".jpg,.jpeg,.png,.webp" onChange={(event) => setAvatarFile(event.target.files?.[0] ?? null)} />
          </label>
          <button className="primary-button compact" type="submit" disabled={isSavingAvatar}>
            {isSavingAvatar ? "上传中..." : "更新头像"}
          </button>
        </form>

        <form className="panel form-stack" onSubmit={saveProfile}>
          <p className="eyebrow">资料</p>
          <label>
            用户名
            <input value={frontUser?.username ?? ""} disabled />
          </label>
          <label>
            昵称
            <input value={nickname} maxLength={64} onChange={(event) => setNickname(event.target.value)} />
          </label>
          <label>
            角色
            <input value={roleLabel} disabled />
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

        {isReader ? (
          <form className="panel form-stack" onSubmit={submitApplication}>
            <p className="eyebrow">作者申请</p>
            <label>
              当前状态
              <input value={applicationStatusLabel} disabled />
            </label>
            <label>
              笔名
              <input value={penName} maxLength={64} onChange={(event) => setPenName(event.target.value)} />
            </label>
            <label>
              申请说明
              <textarea value={reason} maxLength={500} rows={5} onChange={(event) => setReason(event.target.value)} />
            </label>
            {application?.reviewNote ? <p className="muted">审核备注：{application.reviewNote}</p> : null}
            <button className="primary-button compact" type="submit" disabled={isSubmittingApplication || application?.status === "pending"}>
              {isSubmittingApplication ? "提交中..." : application?.status === "pending" ? "审核中" : "提交作者申请"}
            </button>
          </form>
        ) : null}
      </section>
    </main>
  );
}
