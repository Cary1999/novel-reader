import { CheckCircle2, FileText, Upload } from "lucide-react";
import { FormEvent, useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { apiClient, ApiError } from "../api/client";
import type { Category, UploadSummary } from "../api/types";
import { useAuth } from "../auth/AuthContext";
import { MAX_UPLOAD_LABEL, validateTxtUploadFile } from "../uploadLimits";

export function AdminUploadPage() {
  const { isAdmin } = useAuth();
  const [title, setTitle] = useState("");
  const [categoryId, setCategoryId] = useState("");
  const [categories, setCategories] = useState<Category[]>([]);
  const [description, setDescription] = useState("");
  const [file, setFile] = useState<File | null>(null);
  const [error, setError] = useState("");
  const [summary, setSummary] = useState<UploadSummary | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setSummary(null);

    if (!title.trim()) {
      setError("书名必填");
      return;
    }
    if (!categoryId) {
      setError("请选择分类");
      return;
    }

    const fileError = validateTxtUploadFile(file);
    if (fileError) {
      setError(fileError);
      return;
    }

    try {
      setIsSubmitting(true);
      const response = await (isAdmin ? apiClient.adminUploadBook : apiClient.uploadBook)({
        title: title.trim(),
        categoryId,
        description: description.trim(),
        file: file!,
      });
      setSummary(response);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "上传失败，请稍后再试");
    } finally {
      setIsSubmitting(false);
    }
  }

  useEffect(() => {
    apiClient.categories()
      .then((response) => setCategories(response.items ?? []))
      .catch(() => setError("分类加载失败"));
  }, []);

  return (
    <main className="page shell admin-page">
      <section className="page-banner admin-banner">
        <div>
          <p className="eyebrow">{isAdmin ? "管理员工具" : "作者工具"}</p>
          <h1>上传 txt 小说</h1>
          <p className="muted">提交后端会保存源文件并解析章节，作者由当前账号自动确定。</p>
        </div>
      </section>

      <section className="admin-layout">
        <form className="upload-form" onSubmit={handleSubmit}>
          <label>
            书名
            <input value={title} maxLength={120} onChange={(event) => setTitle(event.target.value)} />
          </label>
          <label>
            分类
            <select value={categoryId} onChange={(event) => setCategoryId(event.target.value)}>
              <option value="">请选择分类</option>
              {categories.map((category) => (
                <option key={category.id} value={category.id}>{category.name}</option>
              ))}
            </select>
          </label>
          <label>
            简介
            <textarea value={description} rows={6} maxLength={1000} onChange={(event) => setDescription(event.target.value)} placeholder="可选" />
          </label>
          <label className="file-drop">
            <FileText size={28} aria-hidden="true" />
            <span>{file ? file.name : "选择 .txt 文件"}</span>
            <small>{file ? `${(file.size / 1024).toFixed(1)} KB` : `最大 ${MAX_UPLOAD_LABEL}`}</small>
            <input
              type="file"
              accept=".txt,text/plain"
              onChange={(event) => {
                const nextFile = event.target.files?.[0] ?? null;
                setFile(nextFile);
                setError(validateTxtUploadFile(nextFile));
              }}
            />
          </label>
          {error ? <p className="form-error">{error}</p> : null}
          <button className="primary-button wide" type="submit" disabled={isSubmitting}>
            <Upload size={18} aria-hidden="true" />
            {isSubmitting ? "上传解析中..." : "上传并解析"}
          </button>
        </form>

        <aside className="upload-summary">
          {summary ? (
            <>
              <CheckCircle2 size={30} aria-hidden="true" />
              <p className="eyebrow">解析完成</p>
              <h2>{summary.chapterCount} 章已入库</h2>
              <dl>
                <div>
                  <dt>首章</dt>
                  <dd>{summary.firstChapterTitle}</dd>
                </div>
                <div>
                  <dt>末章</dt>
                  <dd>{summary.lastChapterTitle}</dd>
                </div>
                <div>
                  <dt>上传编号</dt>
                  <dd>{summary.uploadId}</dd>
                </div>
              </dl>
              <Link className="primary-button compact" to={`/books/${summary.bookId}`}>查看书籍</Link>
            </>
          ) : (
            <>
              <FileText size={30} aria-hidden="true" />
              <p className="eyebrow">等待上传</p>
              <h2>解析摘要会显示在这里</h2>
              <p className="muted">包括章节数、首章标题和末章标题。上传前不做章节预览。</p>
            </>
          )}
        </aside>
      </section>
    </main>
  );
}
