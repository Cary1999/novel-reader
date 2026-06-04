import { Navigate, Route, Routes, useLocation } from "react-router-dom";
import { Layout } from "./components/Layout";
import { AccountPage } from "./pages/AccountPage";
import { AdminBooksPage } from "./pages/AdminBooksPage";
import { AdminDashboardPage } from "./pages/AdminDashboardPage";
import { BookDetailPage } from "./pages/BookDetailPage";
import { HomePage } from "./pages/HomePage";
import { LoginPage } from "./pages/LoginPage";
import { NotFoundPage } from "./pages/NotFoundPage";
import { ReaderPage } from "./pages/ReaderPage";
import { RegisterPage } from "./pages/RegisterPage";
import { SearchPage } from "./pages/SearchPage";
import { useAuth } from "./auth/AuthContext";

function AdminRoute({ children }: { children: React.ReactNode }) {
  const location = useLocation();
  const { isAuthenticated, isAdminPortal, isReviewer, isLoading } = useAuth();

  if (isLoading) {
    return <main className="page shell"><div className="panel">正在校验权限...</div></main>;
  }

  if (!isAuthenticated) {
    return <Navigate to="/admin/login" replace state={{ from: location }} />;
  }

  if (!isAdminPortal || !isReviewer) {
    return (
      <main className="page shell">
        <div className="panel permission-panel">
          <p className="eyebrow">无权限</p>
          <h1>管理后台仅对审核与运营账号开放</h1>
          <p className="muted">前台读者和作者账号不能直接进入后台。</p>
        </div>
      </main>
    );
  }

  return children;
}

function FrontAuthRoute({ children }: { children: React.ReactNode }) {
  const location = useLocation();
  const { isAuthenticated, isFrontPortal, isLoading } = useAuth();

  if (isLoading) {
    return <main className="page shell"><div className="panel">正在校验登录状态...</div></main>;
  }

  if (!isAuthenticated || !isFrontPortal) {
    return <Navigate to="/login" replace state={{ from: location }} />;
  }

  return children;
}

function AuthorRoute({ children }: { children: React.ReactNode }) {
  const location = useLocation();
  const { isAuthenticated, isAuthor, isLoading } = useAuth();

  if (isLoading) {
    return <main className="page shell"><div className="panel">正在校验作者权限...</div></main>;
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace state={{ from: location }} />;
  }

  if (!isAuthor) {
    return (
      <main className="page shell">
        <div className="panel permission-panel">
          <p className="eyebrow">作者端</p>
          <h1>只有作者可以管理作品</h1>
          <p className="muted">读者可以先在账号页提交作者申请，审核通过后再进入作者端。</p>
        </div>
      </main>
    );
  }

  return children;
}

export function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route index element={<HomePage />} />
        <Route path="login" element={<LoginPage />} />
        <Route path="admin/login" element={<LoginPage />} />
        <Route path="register" element={<RegisterPage />} />
        <Route path="search" element={<SearchPage />} />
        <Route path="books/:bookId" element={<BookDetailPage />} />
        <Route path="books/:bookId/chapters/:chapterId" element={<ReaderPage />} />
        <Route
          path="account"
          element={
            <FrontAuthRoute>
              <AccountPage />
            </FrontAuthRoute>
          }
        />
        <Route
          path="author"
          element={
            <AuthorRoute>
              <AdminBooksPage />
            </AuthorRoute>
          }
        />
        <Route
          path="admin"
          element={
            <AdminRoute>
              <AdminDashboardPage />
            </AdminRoute>
          }
        />
        <Route path="*" element={<NotFoundPage />} />
      </Route>
    </Routes>
  );
}
