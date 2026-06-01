import { LibraryBig, Plus, RefreshCw, Save, Trash2, Upload, X } from "lucide-react";
import { FormEvent, useEffect, useMemo, useState } from "react";
import { Link, useLocation } from "react-router-dom";
import { apiClient, ApiError } from "../api/client";
import type { BookSummary, Category, ChapterDetail, ChapterSummary, UploadSummary } from "../api/types";
import { useAuth } from "../auth/AuthContext";
import { EmptyState, ErrorState, LoadingState } from "../components/StateViews";
import { MAX_UPLOAD_LABEL, validateTxtUploadFile } from "../uploadLimits";

const PAGE_SIZE = 12;

type BookForm = {
  title: string;
  categoryId: string;
  description: string;
  recommendScore: string;
};

type ChapterForm = {
  id: number | null;
  title: string;
  content: string;
};

const emptyBookForm: BookForm = {
  title: "",
  categoryId: "",
  description: "",
  recommendScore: "0",
};

const emptyChapterForm: ChapterForm = {
  id: null,
  title: "",
  content: "",
};

export function AdminBooksPage({ embedded = false }: { embedded?: boolean }) {
  const { isAdmin } = useAuth();
  const location = useLocation();
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
  const isAdminScope = location.pathname.startsWith("/admin");
  const pageTitle = isAdminScope ? "小说管理" : "我的作品";

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
      const response = await (isAdminScope ? apiClient.adminUploadBook : apiClient.uploadBook)({
        title: uploadTitle.trim(),
        categoryId: uploadCategoryId,
        description: uploadDescription.trim(),
        file: uploadFile!,
      });
      setUploadSummary(response);
      setMessage("上传解析完成");
      setIsUploadOpen(false);
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
        isAdminScope
          ? apiClient.adminBooks({ q: query.trim(), page: nextPage, pageSize: PAGE_SIZE })
          : apiClient.myBooks({ q: query.trim(), page: nextPage, pageSize: PAGE_SIZE }),
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

  async function selectBook(book: BookSummary) {
    try {
      setError("");
      setMessage("");
      setIsCreating(false);
      setCoverPreviewUrl("");
      setPendingCoverFile(null);
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
        recommendScore: String(nextBook.recommendScore ?? 0),
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
    setCoverPreviewUrl((prev) => {
      if (prev && prev.startsWith("blob:")) URL.revokeObjectURL(prev);
      return "";
    });
    setPendingCoverFile(null);
  }

  function startCreate() {
    clearSelection();
    setIsCreating(true);
    setMessage("");
    setError("");
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
        const created = await apiClient.createBook({
          title: bookForm.title.trim(),
          categoryId: bookForm.categoryId,
          description: bookForm.description.trim(),
        }, isAdmin ? "admin" : "me");
        setMessage("小说已新建");
        setIsCreating(false);
        await loadBooks(1);
        await selectBook(created);
      } else if (selectedBook) {
        const updated = await apiClient.updateBook(selectedBook.id, {
          title: bookForm.title.trim(),
          categoryId: bookForm.categoryId,
          description: bookForm.description.trim(),
          recommendScore: Number(bookForm.recommendScore || 0),
        });
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
            setPendingCoverFile(null);
            setCoverPreviewUrl((prev) => {
              if (prev && prev.startsWith("blob:")) URL.revokeObjectURL(prev);
              return "";
            });
            setMessage("基础信息已保存，封面已更新");
          } finally {
            setIsUploadingCover(false);
          }
        } else {
          setMessage("小说基础信息已保存");
        }
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
      await selectBook(selectedBook);
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
      await selectBook(selectedBook);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "删除章节失败");
    }
  }

  useEffect(() => {
    void loadBooks(page);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, isAdminScope]);

  const selectedHeading = useMemo(() => {
    if (isCreating) return "新建小说";
    return selectedBook ? `管理《${selectedBook.title}》` : "选择一本小说";
  }, [isCreating, selectedBook]);

  const content = (
    <>
      <section className="page-banner admin-banner">
        <div>
          <p className="eyebrow">{isAdmin ? "管理员工具" : "作者工具"}</p>
          <h1>{pageTitle}</h1>
          <p className="muted">维护基础信息、章节正文、章节更新和删除操作。</p>
        </div>
        <div className="detail-actions">
          <button className="primary-button compact" type="button" onClick={startCreate}>
            <Plus size={16} aria-hidden="true" />
            新建小说
          </button>
          <button className="ghost-button compact" type="button" onClick={openUploadModal}>
            <Upload size={16} aria-hidden="true" />
            上传 txt
          </button>
        </div>
      </section>

      {message ? <p className="success-banner">{message}</p> : null}
      {error ? <ErrorState message={error} onRetry={() => loadBooks(page)} /> : null}

      <section className="admin-manager-grid">
        <div className="admin-book-list panel">
          <form className="admin-search" onSubmit={handleSearch}>
            <input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="搜索书名或简介" />
            <button className="primary-button compact" type="submit">搜索</button>
          </form>

          {isLoading ? <LoadingState label="正在加载小说..." /> : null}
          {!isLoading && books.length === 0 ? <EmptyState title="暂无小说" description="新建或上传小说后会出现在这里。" /> : null}
          {!isLoading && books.length > 0 ? (
            <div className="admin-books">
              {books.map((book) => (
                <button
                  className={`admin-book-row ${selectedBook?.id === book.id ? "active" : ""}`}
                  type="button"
                  key={book.id}
                  onClick={() => selectBook(book)}
                >
                  <LibraryBig size={18} aria-hidden="true" />
                  <span>
                    <strong>{book.title}</strong>
                    <small>{book.author} · {book.category || "未分类"} · {book.chapterCount} 章</small>
                  </span>
                </button>
              ))}
            </div>
          ) : null}

          <div className="pagination compact-pagination">
            <button className="ghost-button" type="button" disabled={page <= 1} onClick={() => setPage((value) => value - 1)}>上一页</button>
            <span>{page} / {totalPages}</span>
            <button className="ghost-button" type="button" disabled={page >= totalPages} onClick={() => setPage((value) => value + 1)}>下一页</button>
          </div>
        </div>

        <section className="admin-editor panel">
          <div className="section-heading">
            <div>
              <p className="eyebrow">{isCreating ? "创建" : "编辑"}</p>
              <h2>{selectedHeading}</h2>
            </div>
            {selectedBook && !isCreating ? (
              <button className="ghost-button compact danger-button" type="button" onClick={handleDeleteBook}>
                <Trash2 size={16} aria-hidden="true" />
                删除整本
              </button>
            ) : null}
          </div>

          {!selectedBook && !isCreating ? <EmptyState title="尚未选择小说" description="从左侧列表选择一本小说，或先新建小说。" /> : null}

          {selectedBook || isCreating ? (
            <>
              <form className="form-stack admin-edit-form" onSubmit={handleSaveBook}>
                <div className="form-grid">
                  <label>
                    书名
                    <input value={bookForm.title} maxLength={120} onChange={(event) => setBookForm({ ...bookForm, title: event.target.value })} />
                  </label>
                  <label>
                    作者
                    <input value={isCreating ? (isAdmin ? "系统" : "当前用户") : selectedBook?.author ?? ""} disabled />
                  </label>
                </div>
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
                  推荐度
                  <input
                    type="number"
                    min={0}
                    step={1}
                    value={bookForm.recommendScore}
                    onChange={(event) => setBookForm({ ...bookForm, recommendScore: event.target.value })}
                  />
                </label>
                <label>
                  简介
                  <textarea value={bookForm.description} rows={4} maxLength={1000} onChange={(event) => setBookForm({ ...bookForm, description: event.target.value })} />
                </label>
                {!isCreating && selectedBook ? (
                  <label>
                    书籍封面
                    <input
                      className="cover-file-input"
                      id="book-cover-file"
                      type="file"
                      accept="image/png,image/jpeg,image/webp"
                      onChange={async (event) => {
                        const file = event.currentTarget.files?.[0];
                        if (!file) return;
                        if (file.size > 10 * 1024 * 1024) {
                          setError("封面图片不能超过 10MB");
                          event.currentTarget.value = "";
                          return;
                        }

                        // Immediate optimistic preview.
                        const objectUrl = URL.createObjectURL(file);
                        setCoverPreviewUrl((prev) => {
                          if (prev && prev.startsWith("blob:")) URL.revokeObjectURL(prev);
                          return objectUrl;
                        });
                        setPendingCoverFile(file);
                        setMessage("已选择新封面，点击保存后生效");
                        event.currentTarget.value = "";
                      }}
                    />

                    <div className={`cover-preview ${isUploadingCover ? "is-uploading" : ""}`}>
                      <label className="cover-preview-hit" htmlFor="book-cover-file" aria-label="点击更换封面">
                        <img src={coverPreviewUrl || selectedBook.coverUrl || ""} alt="" />
                        {!coverPreviewUrl && !selectedBook.coverUrl ? (
                          <span className="cover-preview-empty">点击选择封面</span>
                        ) : null}
                        {isUploadingCover ? <span className="cover-preview-badge">上传中...</span> : pendingCoverFile ? <span className="cover-preview-badge">待保存</span> : null}
                      </label>
                    </div>
                  </label>
                ) : null}
                <button className="primary-button compact" type="submit" disabled={isSavingBook}>
                  <Save size={16} aria-hidden="true" />
                  {isSavingBook ? "保存中..." : isCreating ? "创建小说" : "保存基础信息"}
                </button>
              </form>

              {!isCreating && selectedBook ? (
                <div className="chapter-admin">
                  <div className="section-heading compact-heading">
                    <div>
                      <p className="eyebrow">章节</p>
                      <h2>{chapters.length} 章</h2>
                    </div>
                    <button className="ghost-button compact" type="button" onClick={() => setChapterForm(emptyChapterForm)}>
                      <Plus size={16} aria-hidden="true" />
                      新增章节
                    </button>
                  </div>

                  <div className="chapter-admin-grid">
                    <div className="chapter-admin-list">
                      {chapters.map((chapter) => (
                        <div className="chapter-admin-row" key={chapter.id}>
                          <button type="button" onClick={() => editChapter(chapter)}>
                            <span>{chapter.index}</span>
                            <strong>{chapter.title}</strong>
                          </button>
                          <button className="icon-danger" type="button" title="删除章节" onClick={() => handleDeleteChapter(chapter)}>
                            <Trash2 size={16} aria-hidden="true" />
                          </button>
                        </div>
                      ))}
                      {chapters.length === 0 ? <EmptyState title="暂无章节" description="可以在右侧新增第一章。" /> : null}
                    </div>

                    <form className="form-stack chapter-edit-form" onSubmit={handleSaveChapter}>
                      <p className="eyebrow">{chapterForm.id ? "编辑章节" : "新增章节"}</p>
                      <label>
                        章节标题
                        <input value={chapterForm.title} maxLength={120} onChange={(event) => setChapterForm({ ...chapterForm, title: event.target.value })} />
                      </label>
                      <label>
                        正文
                        <textarea value={chapterForm.content} rows={12} onChange={(event) => setChapterForm({ ...chapterForm, content: event.target.value })} />
                      </label>
                      <div className="detail-actions">
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
                    </form>
                  </div>
                </div>
              ) : null}
            </>
          ) : null}
        </section>
      </section>

      {isUploadOpen ? (
        <div className="modal-backdrop" role="dialog" aria-modal="true" aria-label="上传 txt 小说">
          <div className="modal panel">
            <div className="modal-head">
              <div>
                <p className="eyebrow">{isAdminScope ? "管理员工具" : "作者工具"}</p>
                <h2>上传 txt 小说</h2>
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
                  {isUploadingTxt ? "上传解析中..." : "上传并解析"}
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
