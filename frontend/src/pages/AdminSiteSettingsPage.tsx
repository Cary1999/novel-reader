import { FormEvent, useEffect, useState } from "react";
import { apiClient, ApiError } from "../api/client";
import type { SiteSettingsInput } from "../api/types";
import { ErrorState, LoadingState } from "../components/StateViews";
import { useSiteSettings } from "../site/SiteSettingsContext";

export function AdminSiteSettingsPage() {
  const { settings, refresh, isLoading: isSiteLoading } = useSiteSettings();
  const [form, setForm] = useState<SiteSettingsInput>({
    brandName: settings.brandName,
    brandSubtitle: settings.brandSubtitle,
    heroEyebrow: settings.heroEyebrow,
    heroTitle: settings.heroTitle,
    heroDescription: settings.heroDescription,
  });
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [isSaving, setIsSaving] = useState(false);
  const [isUploading, setIsUploading] = useState(false);
  const [selectedFileName, setSelectedFileName] = useState("");
  const [previewIconUrl, setPreviewIconUrl] = useState("");

  useEffect(() => {
    setForm({
      brandName: settings.brandName,
      brandSubtitle: settings.brandSubtitle,
      heroEyebrow: settings.heroEyebrow,
      heroTitle: settings.heroTitle,
      heroDescription: settings.heroDescription,
    });
  }, [settings]);

  async function save(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    try {
      setIsSaving(true);
      setError("");
      setMessage("");
      await apiClient.updateSiteSettings(form);
      await refresh();
      setMessage("站点设置已保存");
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "保存站点设置失败");
    } finally {
      setIsSaving(false);
    }
  }

  async function uploadIcon(file: File | null) {
    if (!file) return;
    try {
      setIsUploading(true);
      setError("");
      setMessage("");
      setSelectedFileName(file.name);
      setPreviewIconUrl(await readFileAsDataUrl(file));
      await apiClient.uploadSiteIcon(file);
      await refresh();
      setMessage("站点图标已更新");
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "上传站点图标失败");
    } finally {
      setIsUploading(false);
    }
  }

  if (isSiteLoading) {
    return <LoadingState label="正在加载站点设置..." />;
  }

  return (
    <section className="site-settings-grid">
      {message ? <p className="success-banner">{message}</p> : null}
      {error ? <ErrorState message={error} onRetry={() => void refresh()} /> : null}

      <form className="panel form-stack site-settings-form" onSubmit={save}>
        <div>
          <p className="eyebrow">系统设置</p>
          <h2>站点品牌文案</h2>
          <p className="muted">这些内容会同步到品牌区和首页主视觉。</p>
        </div>

        <label>
          品牌主名称
          <input
            value={form.brandName}
            maxLength={120}
            onChange={(event) => setForm((current) => ({ ...current, brandName: event.target.value }))}
          />
        </label>

        <label>
          品牌副标题
          <input
            value={form.brandSubtitle}
            maxLength={255}
            onChange={(event) => setForm((current) => ({ ...current, brandSubtitle: event.target.value }))}
          />
        </label>

        <label>
          首页眉标题
          <input
            value={form.heroEyebrow}
            maxLength={120}
            onChange={(event) => setForm((current) => ({ ...current, heroEyebrow: event.target.value }))}
          />
        </label>

        <label>
          首页主标题
          <input
            value={form.heroTitle}
            maxLength={255}
            onChange={(event) => setForm((current) => ({ ...current, heroTitle: event.target.value }))}
          />
        </label>

        <label>
          首页说明小字
          <textarea
            value={form.heroDescription}
            maxLength={500}
            rows={4}
            onChange={(event) => setForm((current) => ({ ...current, heroDescription: event.target.value }))}
          />
        </label>

        <button className="primary-button" type="submit" disabled={isSaving}>
          {isSaving ? "保存中..." : "保存设置"}
        </button>
      </form>

      <div className="panel site-icon-card">
        <div>
          <p className="eyebrow">品牌图标</p>
          <h2>站点图标上传</h2>
          <p className="muted">支持 JPG、PNG、WebP、SVG，最大 10MB。</p>
        </div>

        <label className="site-icon-picker">
          <div className="site-icon-preview">
            <img src={previewIconUrl || settings.brandIconUrl} alt={`${settings.brandName} 图标`} />
          </div>
          <input
            type="file"
            accept=".jpg,.jpeg,.png,.webp,.svg,image/jpeg,image/png,image/webp,image/svg+xml"
            onChange={(event) => {
              const [file] = Array.from(event.target.files ?? []);
              void uploadIcon(file ?? null);
              event.currentTarget.value = "";
            }}
          />
        </label>

        <p className="site-icon-meta">
          {isUploading
            ? "正在上传图标..."
            : selectedFileName
              ? `已选择并上传：${selectedFileName}`
              : "点击上方图片，选中文件后在弹窗里点“打开”即可上传"}
        </p>
      </div>
    </section>
  );
}

function readFileAsDataUrl(file: File) {
  return new Promise<string>((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(typeof reader.result === "string" ? reader.result : "");
    reader.onerror = () => reject(reader.error ?? new Error("读取图片失败"));
    reader.readAsDataURL(file);
  });
}
