import { BookOpen, LibraryBig, LogOut, Search, Tags, Upload, UserPlus } from "lucide-react";
import { Link, NavLink, Outlet, useLocation } from "react-router-dom";
import { useAuth } from "../auth/AuthContext";

export function Layout() {
  const { isAuthenticated, isAdmin, user, logout } = useAuth();
  const location = useLocation();

  return (
    <div className="app-frame">
      <header className="site-header">
        <Link className="brand" to="/" aria-label="回到首页">
          <BookOpen size={24} aria-hidden="true" />
          <span>xx书屋</span>
        </Link>
        <nav className="site-nav" aria-label="主导航">
          <NavLink to="/" end>发现</NavLink>
          <NavLink to="/search"><Search size={16} aria-hidden="true" />搜索</NavLink>
          {isAuthenticated ? <NavLink to="/my/books"><LibraryBig size={16} aria-hidden="true" />我的作品</NavLink> : null}
          {isAdmin ? <NavLink to="/admin/books"><LibraryBig size={16} aria-hidden="true" />管理</NavLink> : null}
          {isAdmin ? <NavLink to="/admin/categories"><Tags size={16} aria-hidden="true" />分类</NavLink> : null}
          {isAuthenticated ? <NavLink to={isAdmin ? "/admin/upload" : "/my/upload"}><Upload size={16} aria-hidden="true" />上传</NavLink> : null}
        </nav>
        <div className="account-actions">
          {isAuthenticated ? (
            <>
              <Link className="user-chip" to="/account">{user?.nickname || user?.username}</Link>
              <button className="ghost-button icon-button" type="button" onClick={logout} title="退出登录">
                <LogOut size={18} aria-hidden="true" />
                <span>退出</span>
              </button>
            </>
          ) : (
            <>
              <Link className="ghost-button" to="/login" state={{ from: location }}>登录</Link>
              <Link className="primary-button compact" to="/register">
                <UserPlus size={16} aria-hidden="true" />
                注册
              </Link>
            </>
          )}
        </div>
      </header>
      <Outlet />
      <footer className="site-footer">
        <span>本地小说阅读 MVP</span>
        <span>仅使用公开接口和本地上传内容</span>
      </footer>
    </div>
  );
}
