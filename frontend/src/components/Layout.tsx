import { BookOpen, LibraryBig, LogOut, Search, UserPlus } from "lucide-react";
import { Link, NavLink, Outlet, useLocation } from "react-router-dom";
import { useAuth } from "../auth/AuthContext";
import { useSiteSettings } from "../site/SiteSettingsContext";

export function Layout() {
  const { isAuthenticated, isAdmin, user, logout } = useAuth();
  const location = useLocation();
  const { settings } = useSiteSettings();

  return (
    <div className="app-frame">
      <header className="site-header">
        <div className="shell site-header-inner">
          <Link className="brand" to="/" aria-label="回到首页">
            <span className="brand-mark" aria-hidden="true">
              <img src={settings.brandIconUrl} alt="" />
            </span>
            <span className="brand-copy">
              <strong>{settings.brandName}</strong>
              <small>{settings.brandSubtitle}</small>
            </span>
          </Link>
          <nav className="site-nav" aria-label="主导航">
            <NavLink to="/" end>发现</NavLink>
            <NavLink to="/search"><Search size={16} aria-hidden="true" />书库</NavLink>
            {isAuthenticated ? <NavLink to="/my/books"><LibraryBig size={16} aria-hidden="true" />我的作品</NavLink> : null}
            {isAdmin ? <NavLink to="/admin"><LibraryBig size={16} aria-hidden="true" />管理后台</NavLink> : null}
          </nav>
          <div className="account-actions">
            {isAuthenticated ? (
              <>
                <Link className="user-chip" to="/account">
                  <span className="user-chip-label">当前用户</span>
                  <strong>{user?.nickname || user?.username}</strong>
                </Link>
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
        </div>
      </header>
      <Outlet />
      <footer className="site-footer">
        <div className="shell site-footer-inner">
          <span>阅卷书屋 · 本地小说阅读 MVP</span>
          <span>仅使用公开接口与本地上传内容，不复制第三方品牌与视觉资产</span>
        </div>
      </footer>
    </div>
  );
}
