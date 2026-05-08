import { Navigate, Route, Routes, useLocation } from "react-router-dom";
import { Layout } from "./components/Layout";
import { AccountPage } from "./pages/AccountPage";
import { AdminBooksPage } from "./pages/AdminBooksPage";
import { AdminCategoriesPage } from "./pages/AdminCategoriesPage";
import { AdminUploadPage } from "./pages/AdminUploadPage";
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
  const { isAuthenticated, isAdmin, isLoading } = useAuth();

  if (isLoading) {
    return <main className="page shell"><div className="panel">正在校验权限...</div></main>;
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace state={{ from: location }} />;
  }

  if (!isAdmin) {
    return (
      <main className="page shell">
        <div className="panel permission-panel">
          <p className="eyebrow">无权限</p>
          <h1>管理员后台仅对管理员开放</h1>
          <p className="muted">当前账号可以阅读章节，但不能上传小说。</p>
        </div>
      </main>
    );
  }

  return children;
}

function AuthRoute({ children }: { children: React.ReactNode }) {
  const location = useLocation();
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return <main className="page shell"><div className="panel">正在校验登录状态...</div></main>;
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace state={{ from: location }} />;
  }

  return children;
}

export function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route index element={<HomePage />} />
        <Route path="login" element={<LoginPage />} />
        <Route path="register" element={<RegisterPage />} />
        <Route path="search" element={<SearchPage />} />
        <Route path="books/:bookId" element={<BookDetailPage />} />
        <Route path="books/:bookId/chapters/:chapterId" element={<ReaderPage />} />
        <Route
          path="account"
          element={
            <AuthRoute>
              <AccountPage />
            </AuthRoute>
          }
        />
        <Route
          path="my/books"
          element={
            <AuthRoute>
              <AdminBooksPage />
            </AuthRoute>
          }
        />
        <Route
          path="my/upload"
          element={
            <AuthRoute>
              <AdminUploadPage />
            </AuthRoute>
          }
        />
        <Route
          path="admin/books"
          element={
            <AdminRoute>
              <AdminBooksPage />
            </AdminRoute>
          }
        />
        <Route
          path="admin/categories"
          element={
            <AdminRoute>
              <AdminCategoriesPage />
            </AdminRoute>
          }
        />
        <Route
          path="admin/upload"
          element={
            <AdminRoute>
              <AdminUploadPage />
            </AdminRoute>
          }
        />
        <Route path="*" element={<NotFoundPage />} />
      </Route>
    </Routes>
  );
}
