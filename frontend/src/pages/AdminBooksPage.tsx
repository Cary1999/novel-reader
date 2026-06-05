import { LibraryBig, Plus, RefreshCw, Save, Trash2, Upload, X } from "lucide-react";
import { FormEvent, useEffect, useState } from "react";
import { apiClient, ApiError } from "../api/client";
import type { BookSummary, Category, ChapterDetail, ChapterSummary, UploadSummary } from "../api/types";
import { EmptyState, ErrorState, LoadingState } from "../components/StateViews";
import { MAX_COVER_LABEL, MAX_UPLOAD_LABEL, validateCoverUploadFile, validateTxtUploadFile } from "../uploadLimits";

const PAGE_SIZE = 12;

type BookForm = {
  title: string;
  categoryId: string;
  description: string;
};

type ChapterForm = {
  id: number | null;
  title: string;
  content: string;
};

type WorkspaceSection = "basic" | "cover" | "chapters";

const emptyBookForm: BookForm = {
  title: "",
  categoryId: "",
  description: "",
};

const emptyChapterForm: ChapterForm = {
  id: null,
  title: "",
  content: "",
};

function formatDateTime(value?: string | null) {
  if (!value) return "";
  const timestamp = Date.parse(value);
  if (!Number.isFinite(timestamp)) return "";
  return new Intl.DateTimeFormat("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  }).format(timestamp);
}

export function AdminBooksPage({ embedded = false }: { embedded?: boolean }) {
  const [query, setQuery] = useState("");
  const [page, setPage] = useState(1);
  const [books, setBooks] = useState<BookSummary[]>([]);
  const [total, setTotal] = useState(0);
  const [categories, setCategories] = useState<Category[]>([]);
  const [selectedBook, setSelectedBook] = useState<BookSummary | null>(null);
  const [bookForm, setBookForm] = useState<BookForm>(emptyBookForm);
  const [isCreating, setIsCreating] = useState(false);
  const [chapters, setChapters] = useState<ChapterSummary[]>([]);
  const [chapterForm, setChapterForm] = useState<ChapterForm>(emptyChapterForm);
  const [isLoading, setIsLoading] = useState(true);
  const [isSavingBook, setIsSavingBook] = useState(false);
  const [isSavingChapter, setIsSavingChapter] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [activeSection, setActiveSection] = useState<WorkspaceSection>("basic");
  const [lastSavedAt, setLastSavedAt] = useState<string | null>(null);
  const [coverPreviewUrl, setCoverPreviewUrl] = useState<string>("");
  const [pendingCoverFile, setPendingCoverFile] = useState<File | null>(null);
  const [isUploadingCover, setIsUploadingCover] = useState(false);
  const [isUploadOpen, setIsUploadOpen] = useState(false);
  const [uploadTitle, setUploadTitle] = useState("");
  const [uploadCategoryId, setUploadCategoryId] = useState("");
  const [uploadDescription, setUploadDescription] = useState("");
  const [uploadFile, setUploadFile] = useState<File | null>(null);
  const [uploadError, setUploadError] = useState("");
  const [uploadSummary, setUploadSummary] = useState<UploadSummary | null>(null);
  const [isUploadingTxt, setIsUploadingTxt] = useState(false);

  const totalPages = Math.max(Math.ceil(total / PAGE_SIZE), 1);
  const pageTitle = "我的作品";
  const toolLabel = "作者工具";

  function clearPendingCover() {
    setPendingCoverFile(null);
    setCoverPreviewUrl((prev) => {
      if (prev && prev.startsWith("blob:")) URL.revokeObjectURL(prev);
      return "";
    });
  }

  function stageCoverFile(file: File) {
    const fileError = validateCoverUploadFile(file);
    if (fileError) {
      setError(fileError);
      return false;
    }

    const objectUrl = URL.createObjectURL(file);
    setCoverPreviewUrl((prev) => {
      if (prev && prev.startsWith("blob:")) URL.revokeObjectURL(prev);
      return objectUrl;
    });
    setPendingCoverFile(file);
    setError("");
    return true;
  }

  function openUploadModal() {
    setUploadError("");
    setUploadSummary(null);
    setUploadTitle("");
    setUploadCategoryId("");
    setUploadDescription("");
    setUploadFile(null);
    setIsUploadOpen(true);
  }

  async function submitUpload() {
    setUploadError("");
    setUploadSummary(null);
    if (!uploadTitle.trim()) {
      setUploadError("书名必填");
      return;
    }
    if (!uploadCategoryId) {
      setUploadError("请选择分类");
      return;
    }
    const fileError = validateTxtUploadFile(uploadFile);
    if (fileError) {
      setUploadError(fileError);
      return;
    }
    try {
      setIsUploadingTxt(true);
      const response = await apiClient.uploadBook({
        title: uploadTitle.trim(),
        categoryId: uploadCategoryId,
        description: uploadDescription.trim(),
        file: uploadFile!,
      });
      setUploadSummary(response);
      setMessage("导入解析完成");
      setLastSavedAt(new Date().toISOString());
      setIsUploadOpen(false);
      setPage(1);
      await loadBooks(1);
    } catch (err) {
      setUploadError(err instanceof ApiError ? err.message : "上传失败，请稍后再试");
    } finally {
      setIsUploadingTxt(false);
    }
  }

  async function loadBooks(nextPage = page) {
    try {
      setIsLoading(true);
      setError("");
      const [categoryResponse, bookResponse] = await Promise.all([
        apiClient.categories(),
        apiClient.myBooks({ q: query.trim(), page: nextPage, pageSize: PAGE_SIZE }),
      ]);
      const nextBooks = bookResponse.items ?? [];
      setCategories(categoryResponse.items ?? []);
      setBooks(nextBooks);
      setTotal(bookResponse.total ?? 0);
      if (selectedBook && !nextBooks.some((book) => book.id === selectedBook.id)) {
        clearSelection();
      }
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "加载小说列表失败");
    } finally {
      setIsLoading(false);
    }
  }

  async function selectBook(
    book: BookSummary,
    options?: { preservePendingCover?: boolean; lastSavedAt?: string | null },
  ) {
    try {
      setError("");
      setMessage("");
      setIsCreating(false);
      setActiveSection("basic");
      setLastSavedAt(options?.lastSavedAt ?? null);
      if (!options?.preservePendingCover) {
        clearPendingCover();
      }
      const [detail, chapterResponse] = await Promise.all([
        apiClient.book(book.id),
        apiClient.chapters(book.id),
      ]);
      const nextBook = { ...book, ...detail };
      setSelectedBook(nextBook);
      setBookForm({
        title: nextBook.title,
        categoryId: String(nextBook.categoryId ?? ""),
        description: nextBook.description ?? "",
      });
      setChapters(chapterResponse.items ?? []);
      setChapterForm(emptyChapterForm);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "加载小说详情失败");
    }
  }

  function clearSelection() {
    setSelectedBook(null);
    setBookForm(emptyBookForm);
    setChapters([]);
    setChapterForm(emptyChapterForm);
    setIsCreating(false);
    setActiveSection("basic");
    setLastSavedAt(null);
    clearPendingCover();
  }

  function startCreate() {
    clearSelection();
    setIsCreating(true);
    setMessage("");
    setError("");
    setActiveSection("basic");
    setLastSavedAt(null);
  }

  async function handleSearch(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setPage(1);
    await loadBooks(1);
  }

  async function handleSaveBook(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!bookForm.title.trim()) {
      setError("书名必填");
      return;
    }
    if (!bookForm.categoryId) {
      setError("请选择分类");
      return;
    }
    try {
      setIsSavingBook(true);
      setError("");
      if (isCreating) {
        const coverFile = pendingCoverFile;
        const savedAt = new Date().toISOString();
        const created = await apiClient.createBook({
          title: bookForm.title.trim(),
          categoryId: bookForm.categoryId,
          description: bookForm.description.trim(),
        });
        let coverUploadError = "";
        let createdMessage = "小说已新建";

        if (coverFile) {
          try {
            setIsUploadingCover(true);
            await apiClient.uploadBookCover(created.id, coverFile);
            createdMessage = "小说和封面已创建";
          } catch (err) {
            coverUploadError = err instanceof ApiError ? err.message : "封面上传失败，请稍后再试";
          } finally {
            setIsUploadingCover(false);
          }
        }

        setIsCreating(false);
        setPage(1);
        await loadBooks(1);
        await selectBook(created, { preservePendingCover: Boolean(coverUploadError), lastSavedAt: savedAt });
        if (coverUploadError) {
          setError(`小说已创建，但封面上传失败：${coverUploadError}`);
          setMessage("");
        } else {
          setMessage(createdMessage);
        }
      } else if (selectedBook) {
        const payload = {
          title: bookForm.title.trim(),
          categoryId: bookForm.categoryId,
          description: bookForm.description.trim(),
        };
        const savedAt = new Date().toISOString();
        const updated = await apiClient.updateBook(selectedBook.id, payload);
        let nextSelected = { ...selectedBook, ...updated };
        setSelectedBook(nextSelected);
        setBooks((items) => items.map((book) => book.id === updated.id ? { ...book, ...updated } : book));

        // Cover upload is applied only after clicking "save".
        if (pendingCoverFile) {
          try {
            setIsUploadingCover(true);
            const result = await apiClient.uploadBookCover(selectedBook.id, pendingCoverFile);
            nextSelected = { ...nextSelected, coverUrl: result.coverUrl };
            setSelectedBook(nextSelected);
            setBooks((items) => items.map((book) => book.id === selectedBook.id ? { ...book, coverUrl: result.coverUrl } : book));
            clearPendingCover();
            setMessage("基础信息已保存，封面已更新");
          } finally {
            setIsUploadingCover(false);
          }
        } else {
          setMessage("小说基础信息已保存");
        }
        setLastSavedAt(savedAt);
      }
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "保存小说失败");
    } finally {
      setIsSavingBook(false);
    }
  }

  async function handleDeleteBook() {
    if (!selectedBook) return;
    if (!window.confirm(`确定删除《${selectedBook.title}》？该操作会删除全部章节。`)) return;
    try {
      setError("");
      await apiClient.deleteBook(selectedBook.id);
      setMessage("小说已删除");
      clearSelection();
      await loadBooks(page);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "删除小说失败");
    }
  }

  async function editChapter(chapter: ChapterSummary) {
    if (!selectedBook) return;
    try {
      setError("");
      setActiveSection("chapters");
      const detail: ChapterDetail = await apiClient.chapter(selectedBook.id, chapter.id);
      setChapterForm({ id: detail.id, title: detail.title, content: detail.content });
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "加载章节正文失败");
    }
  }

  async function handleSaveChapter(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedBook) return;
    if (!chapterForm.title.trim() || !chapterForm.content.trim()) {
      setError("章节标题和正文必填");
      return;
    }
    try {
      setIsSavingChapter(true);
      setError("");
      const payload = { title: chapterForm.title.trim(), content: chapterForm.content.trim() };
      if (chapterForm.id) {
        await apiClient.updateChapter(selectedBook.id, chapterForm.id, payload);
        setMessage("章节已保存");
      } else {
        await apiClient.addChapter(selectedBook.id, payload);
        setMessage("新章节已追加");
      }
      setActiveSection("chapters");
      await selectBook(selectedBook, { lastSavedAt: new Date().toISOString() });
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "保存章节失败");
    } finally {
      setIsSavingChapter(false);
    }
  }

  async function handleDeleteChapter(chapter: ChapterSummary) {
    if (!selectedBook) return;
    if (!window.confirm(`确定删除章节“${chapter.title}”？后续章节序号会自动重排。`)) return;
    try {
      setError("");
      await apiClient.deleteChapter(selectedBook.id, chapter.id);
      setMessage("章节已删除并重新排序");
      setActiveSection("chapters");
      await selectBook(selectedBook, { lastSavedAt: new Date().toISOString() });
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "删除章节失败");
    }
  }

  useEffect(() => {
    void loadBooks(page);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page]);

  const hasEditorTarget = Boolean(selectedBook || isCreating);
  const overviewTitle = isCreating ? "新建小说" : selectedBook?.title ?? "还没有打开作品";
  const overviewCoverUrl = coverPreviewUrl || selectedBook?.coverUrl || "";
  const selectedCategoryName = isCreating
    ? (bookForm.categoryId ? categories.find((item) => String(item.id) === bookForm.categoryId)?.name ?? "未分类" : "未选择")
    : (selectedBook?.category || "未分类");
  const overviewMeta = isCreating
    ? [
        { label: "作者", value: "当前账号" },
        { label: "分类", value: selectedCategoryName },
        { label: "章节", value: "0 章" },
      ]
    : [
        { label: "作者", value: selectedBook?.author ?? "未知" },
        { label: "分类", value: selectedCategoryName },
        { label: "章节", value: `${selectedBook?.chapterCount ?? 0} 章` },
      ];
  const canEditChapters = Boolean(selectedBook && !isCreating);
  const overviewStatusText = pendingCoverFile
    ? isCreating
      ? "封面待随作品创建"
      : "封面待保存"
    : lastSavedAt
      ? `最近保存 ${formatDateTime(lastSavedAt)}`
      : selectedBook?.createdAt
        ? `创建于 ${formatDateTime(selectedBook.createdAt)}`
        : "等待编辑";
  const overviewLatestText = isCreating
    ? "待创建后自动生成目录"
    : selectedBook?.latestChapterTitle
      ? `最新章节：${selectedBook.latestChapterTitle}`
      : "暂无章节";
  const summaryCoverLabel = overviewCoverUrl ? "已设置封面" : "暂无封面";

  const content = (
    <>
      <section className="page-banner admin-banner">
        <div>
          <p className="eyebrow">{toolLabel}</p>
          <h1>{pageTitle}</h1>
          <p className="muted">按作品概览、基础信息、封面和章节拆开编辑，先看清状态，再进入具体操作。</p>
        </div>
        <div className="detail-actions">
          <button className="primary-button compact" type="button" onClick={startCreate}>
            <Plus size={16} aria-hidden="true" />
            新建小说
          </button>
          <button className="ghost-button compact" type="button" onClick={openUploadModal}>
            <Upload size={16} aria-hidden="true" />
            导入 txt
          </button>
        </div>
      </section>

      {message ? <p className="success-banner">{message}</p> : null}
      {error && !hasEditorTarget && !books.length ? <ErrorState message={error} onRetry={() => loadBooks(page)} /> : null}
      {error && (hasEditorTarget || books.length > 0) ? <p className="form-error">{error}</p> : null}

      <section className="author-workspace">
        <aside className="panel author-library">
          <div className="section-heading author-library-head">
            <div>
              <p className="eyebrow">作品列表</p>
              <h2>第 {page} 页 / 共 {totalPages} 页</h2>
            </div>
            <button className="ghost-button compact" type="button" onClick={() => void loadBooks(page)}>
              刷新
            </button>
          </div>

          <form className="admin-search author-search" onSubmit={handleSearch}>
            <input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="搜索书名或简介" />
            <button className="primary-button compact" type="submit">搜索</button>
          </form>

          {isLoading ? <LoadingState label="正在加载作品..." /> : null}
          {!isLoading && books.length === 0 ? <EmptyState title="暂无作品" description="先新建小说，或者导入一个 txt 作品。" /> : null}
          {!isLoading && books.length > 0 ? (
            <div className="author-book-list">
              {books.map((book) => (
                <button
                  className={`author-book-row ${selectedBook?.id === book.id ? "active" : ""}`}
                  type="button"
                  key={book.id}
                  onClick={() => selectBook(book)}
                >
                  <span className="author-book-thumb" aria-hidden="true">
                    {book.coverUrl ? (
                      <img src={book.coverUrl} alt="" />
                    ) : (
                      <LibraryBig size={18} aria-hidden="true" />
                    )}
                  </span>
                  <span className="author-book-copy">
                    <strong>{book.title}</strong>
                    <small>{book.author} · {book.category || "未分类"} · {book.chapterCount} 章</small>
                    <span>{book.latestChapterTitle ? `最新：${book.latestChapterTitle}` : "暂无章节"}</span>
                  </span>
                </button>
              ))}
            </div>
          ) : null}

          <div className="pagination compact-pagination author-pagination">
            <button className="ghost-button" type="button" disabled={page <= 1} onClick={() => setPage((value) => value - 1)}>上一页</button>
            <span>{page} / {totalPages}</span>
            <button className="ghost-button" type="button" disabled={page >= totalPages} onClick={() => setPage((value) => value + 1)}>下一页</button>
          </div>
        </aside>

        <div className="author-main">
          {hasEditorTarget ? (
            <>
              <section className="panel author-overview">
                <div className="author-overview-cover">
                  {overviewCoverUrl ? (
                    <img src={overviewCoverUrl} alt="" />
                  ) : (
                    <div className="author-cover-empty">
                      <LibraryBig size={28} aria-hidden="true" />
                      <span>{isCreating ? "待设置封面" : "暂无封面"}</span>
                    </div>
                  )}
                </div>

                <div className="author-overview-copy">
                  <p className="eyebrow">作品概览</p>
                  <h1 className="author-workspace-title">{overviewTitle}</h1>

                  <div className="author-overview-meta">
                    {overviewMeta.map((item) => (
                      <div key={item.label} className="author-stat">
                        <span>{item.label}</span>
                        <strong>{item.value}</strong>
                      </div>
                    ))}
                  </div>
                </div>
              </section>

              <section className="panel author-tabs" aria-label="作品编辑切换">
                <button
                  type="button"
                  className={`author-tab ${activeSection === "basic" ? "active" : ""}`}
                  onClick={() => setActiveSection("basic")}
                >
                  基础信息
                </button>
                <button
                  type="button"
                  className={`author-tab ${activeSection === "cover" ? "active" : ""}`}
                  onClick={() => setActiveSection("cover")}
                >
                  封面设置
                </button>
                <button
                  type="button"
                  className={`author-tab ${activeSection === "chapters" ? "active" : ""}`}
                  onClick={() => setActiveSection("chapters")}
                  disabled={!canEditChapters}
                  title={!canEditChapters ? "保存作品后才能管理章节" : undefined}
                >
                  章节管理
                </button>
                <div className="author-tab-spacer" />
                <div className="author-tab-actions">
                  {selectedBook && !isCreating ? (
                    <button className="ghost-button compact danger-button" type="button" onClick={handleDeleteBook}>
                      <Trash2 size={16} aria-hidden="true" />
                      删除作品
                    </button>
                  ) : null}
                </div>
              </section>

              {activeSection === "basic" ? (
                <section id="book-info" className="panel author-section author-info-card">
                <div className="section-heading compact-heading">
                  <div>
                    <p className="eyebrow">基础信息</p>
                    <h2>作品信息与封面设置</h2>
                  </div>
                  <div className="detail-actions">
                    <span className="muted">{overviewStatusText}</span>
                  </div>
                </div>

                  <div className="author-basic-grid">
                    <form id="book-info-form" className="form-stack author-info-form" onSubmit={handleSaveBook}>
                      <div className="detail-actions author-form-actions author-form-actions-top">
                        <button className="primary-button compact" type="submit" disabled={isSavingBook}>
                          <Save size={16} aria-hidden="true" />
                          {isSavingBook ? "保存中..." : isCreating ? "创建小说" : "保存作品信息"}
                        </button>
                        <button
                          className="ghost-button compact"
                          type="button"
                          onClick={() => setBookForm(selectedBook || isCreating ? {
                            title: isCreating ? "" : selectedBook?.title ?? "",
                            categoryId: isCreating ? "" : String(selectedBook?.categoryId ?? ""),
                            description: isCreating ? "" : selectedBook?.description ?? "",
                          } : emptyBookForm)}
                        >
                          <RefreshCw size={16} aria-hidden="true" />
                          重置表单
                        </button>
                      </div>
                      <div className="form-grid">
                        <label>
                          书名
                          <input value={bookForm.title} maxLength={120} onChange={(event) => setBookForm({ ...bookForm, title: event.target.value })} />
                        </label>
                        <label>
                          分类
                          <select value={bookForm.categoryId} onChange={(event) => setBookForm({ ...bookForm, categoryId: event.target.value })}>
                            <option value="">请选择分类</option>
                            {categories.map((category) => (
                              <option key={category.id} value={category.id}>{category.name}</option>
                            ))}
                          </select>
                        </label>
                        <label>
                          作者
                          <input value={isCreating ? "当前作者" : selectedBook?.author ?? ""} disabled />
                        </label>
                        <label>
                          保存状态
                          <input value={overviewStatusText} disabled />
                        </label>
                      </div>
                      <label>
                        简介
                        <textarea
                          value={bookForm.description}
                          rows={8}
                          maxLength={1000}
                          onChange={(event) => setBookForm({ ...bookForm, description: event.target.value })}
                        />
                      </label>
                  </form>

                  <aside className="author-basic-note panel-inset">
                    <p className="eyebrow">当前状态</p>
                    <div className="author-side-stats">
                      <div>
                        <span>保存状态</span>
                        <strong>{overviewStatusText}</strong>
                      </div>
                      <div>
                        <span>最新章节</span>
                        <strong>{overviewLatestText.replace("最新章节：", "")}</strong>
                      </div>
                      <div>
                        <span>封面状态</span>
                        <strong>{summaryCoverLabel}</strong>
                      </div>
                    </div>
                  </aside>
                </div>
                </section>
              ) : null}

              {activeSection === "cover" ? (
                <section id="cover" className="panel author-section author-cover-panel">
                  <div className="section-heading compact-heading">
                    <div>
                      <p className="eyebrow">封面设置</p>
                      <h2>封面预览与上传</h2>
                    </div>
                    <span className="muted">JPG / PNG / WebP，最大 {MAX_COVER_LABEL}</span>
                  </div>

                  <div className="author-cover-panel-grid">
                    <div className="author-cover-stage">
                      <div className="detail-actions author-form-actions author-form-actions-top">
                        <label className="ghost-button wide cover-select-button" htmlFor="book-cover-file">
                          选择封面
                        </label>
                        <button className="primary-button compact" type="submit" form="book-info-form">
                          <Save size={16} aria-hidden="true" />
                          保存并同步
                        </button>
                      </div>
                      <div className="author-cover-preview large">
                        {overviewCoverUrl ? (
                          <img src={overviewCoverUrl} alt="" />
                        ) : (
                          <div className="author-cover-empty large">
                            <Upload size={22} aria-hidden="true" />
                            <span>当前还没有封面</span>
                          </div>
                        )}
                        {isUploadingCover ? <span className="cover-preview-badge">上传中...</span> : pendingCoverFile ? <span className="cover-preview-badge">{isCreating ? "待创建" : "待保存"}</span> : null}
                      </div>
                      <input
                        className="cover-file-input"
                        id="book-cover-file"
                        type="file"
                        accept="image/png,image/jpeg,image/webp"
                        onChange={(event) => {
                          const file = event.currentTarget.files?.[0];
                          if (file && stageCoverFile(file)) {
                            setMessage(isCreating ? "已选择封面，创建小说时会一并上传" : "已选择新封面，保存作品信息后生效");
                          }
                          event.currentTarget.value = "";
                        }}
                      />
                    </div>

                    <aside className="author-basic-note panel-inset">
                      <p className="eyebrow">说明</p>
                      <ul className="author-tip-list">
                        <li>封面会在保存作品信息时同步上传。</li>
                        <li>如果你先选了图，再切去别的 tab，预览会保留。</li>
                        <li>创建新作品时也可以先预选封面。</li>
                      </ul>
                    </aside>
                  </div>
                </section>
              ) : null}

              {activeSection === "chapters" ? (
                <section id="chapters" className="panel author-section author-chapter-card">
                  <div className="section-heading compact-heading">
                    <div>
                      <p className="eyebrow">章节管理</p>
                      <h2>{chapters.length} 章</h2>
                    </div>
                    <div className="detail-actions">
                      <button className="ghost-button compact" type="button" onClick={() => { setChapterForm(emptyChapterForm); setActiveSection("chapters"); }}>
                        <Plus size={16} aria-hidden="true" />
                        新增章节
                      </button>
                    </div>
                  </div>

                  <div className="author-chapter-grid">
                    <div className="author-chapter-list">
                      <div className="detail-actions author-form-actions author-form-actions-top">
                        <button className="ghost-button compact" type="button" onClick={() => { setChapterForm(emptyChapterForm); setActiveSection("chapters"); }}>
                          <Plus size={16} aria-hidden="true" />
                          新增章节
                        </button>
                      </div>
                      {chapters.map((chapter) => (
                        <div className="author-chapter-row" key={chapter.id}>
                          <button type="button" onClick={() => editChapter(chapter)}>
                            <span>{chapter.index}</span>
                            <strong>{chapter.title}</strong>
                          </button>
                          <button className="icon-danger" type="button" title="删除章节" onClick={() => handleDeleteChapter(chapter)}>
                            <Trash2 size={16} aria-hidden="true" />
                          </button>
                        </div>
                      ))}
                      {chapters.length === 0 ? <EmptyState title="暂无章节" description="先新增第一章，或者从左侧选择别的作品。" /> : null}
                    </div>

                    <form className="form-stack author-chapter-editor" onSubmit={handleSaveChapter}>
                      <div className="section-heading compact-heading">
                        <div>
                          <p className="eyebrow">{chapterForm.id ? "编辑章节" : "新增章节"}</p>
                          <h2>{chapterForm.id ? "当前章节" : "准备写入新章节"}</h2>
                        </div>
                        <span className="muted">{chapterForm.id ? "编辑模式" : "新增模式"}</span>
                      </div>
                      <div className="detail-actions author-form-actions author-form-actions-top">
                        <button className="primary-button compact" type="submit" disabled={isSavingChapter}>
                          <Save size={16} aria-hidden="true" />
                          {isSavingChapter ? "保存中..." : "保存章节"}
                        </button>
                        {chapterForm.id ? (
                          <button className="ghost-button compact" type="button" onClick={() => setChapterForm(emptyChapterForm)}>
                            <RefreshCw size={16} aria-hidden="true" />
                            改为新增
                          </button>
                        ) : null}
                      </div>
                      <label>
                        章节标题
                        <input value={chapterForm.title} maxLength={120} onChange={(event) => setChapterForm({ ...chapterForm, title: event.target.value })} />
                      </label>
                      <label>
                        正文
                        <textarea value={chapterForm.content} rows={14} onChange={(event) => setChapterForm({ ...chapterForm, content: event.target.value })} />
                      </label>
                    </form>
                  </div>
                </section>
              ) : null}
            </>
          ) : (
            <section className="panel author-empty-panel">
              <EmptyState title="还没有打开作品" description="从左侧选择一本作品，或者先新建小说、导入 txt 开始编辑。" />
            </section>
          )}
        </div>
      </section>

      {isUploadOpen ? (
        <div className="modal-backdrop" role="dialog" aria-modal="true" aria-label="导入 txt 作品">
          <div className="modal panel">
            <div className="modal-head">
              <div>
                <p className="eyebrow">{toolLabel}</p>
                <h2>导入 txt 作品</h2>
              </div>
              <button className="ghost-button icon-button" type="button" onClick={() => setIsUploadOpen(false)} aria-label="关闭">
                <X size={18} aria-hidden="true" />
              </button>
            </div>

            <div className="form-stack">
              <label>
                书名
                <input value={uploadTitle} maxLength={120} onChange={(e) => setUploadTitle(e.target.value)} />
              </label>
              <label>
                分类
                <select value={uploadCategoryId} onChange={(e) => setUploadCategoryId(e.target.value)}>
                  <option value="">请选择分类</option>
                  {categories.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
                </select>
              </label>
              <label>
                简介
                <textarea value={uploadDescription} rows={4} maxLength={1000} onChange={(e) => setUploadDescription(e.target.value)} placeholder="可选" />
              </label>
              <label className="file-drop">
                <Upload size={22} aria-hidden="true" />
                <span>{uploadFile ? uploadFile.name : "选择 .txt 文件"}</span>
                <small>{uploadFile ? `${(uploadFile.size / 1024).toFixed(1)} KB` : `最大 ${MAX_UPLOAD_LABEL}`}</small>
                <input
                  type="file"
                  accept=".txt,text/plain"
                  onChange={(e) => {
                    const nextFile = e.target.files?.[0] ?? null;
                    setUploadFile(nextFile);
                    setUploadError(validateTxtUploadFile(nextFile));
                  }}
                />
              </label>
              {uploadError ? <p className="form-error">{uploadError}</p> : null}
              {uploadSummary ? <p className="muted">已解析：{uploadSummary.chapterCount} 章</p> : null}
              <div className="modal-actions">
                <button className="ghost-button" type="button" onClick={() => setIsUploadOpen(false)}>取消</button>
                <button className="primary-button" type="button" disabled={isUploadingTxt} onClick={() => void submitUpload()}>
                  <Upload size={18} aria-hidden="true" />
                  {isUploadingTxt ? "解析中..." : "上传并解析"}
                </button>
              </div>
            </div>
          </div>
        </div>
      ) : null}
    </>
  );

  if (embedded) return content;

  return (
    <main className="page shell admin-page">
      {content}
    </main>
  );
}
