import { LibraryBig, LogOut, ShieldCheck, UserPlus } from "lucide-react";
import { Link, NavLink, Outlet, useLocation } from "react-router-dom";
import { useAuth } from "../auth/AuthContext";
import { useSiteSettings } from "../site/SiteSettingsContext";

export function Layout() {
  const { isAuthenticated, isAdminPortal, isAuthor, isFrontPortal, user, logout } = useAuth();
  const location = useLocation();
  const { settings } = useSiteSettings();
  const inAdmin = location.pathname.startsWith("/admin");
  const frontDisplayName = user && "nickname" in user ? user.nickname || user.username : user?.username ?? "";

  return (
    <div className="app-frame">
      <header className="site-header">
        <div className="shell site-header-inner">
          <Link className="brand" to={inAdmin ? "/admin" : "/"} aria-label="回到首页">
            <span className="brand-mark" aria-hidden="true">
              <img src={settings.brandIconUrl} alt="" />
            </span>
            <span className="brand-copy">
              <strong>{settings.brandName}</strong>
              <small>{settings.brandSubtitle}</small>
            </span>
          </Link>
          <nav className="site-nav" aria-label="主导航">
            {inAdmin ? (
              <>
                <NavLink to="/admin"><ShieldCheck size={16} aria-hidden="true" />审核后台</NavLink>
              </>
            ) : (
              <>
                <NavLink to="/" end>发现</NavLink>
                <NavLink to="/bookshelf"><LibraryBig size={16} aria-hidden="true" />我的书架</NavLink>
                {isAuthenticated && isAuthor ? <NavLink to="/author"><LibraryBig size={16} aria-hidden="true" />作者端</NavLink> : null}
              </>
            )}
          </nav>
          <div className="account-actions">
            {isAuthenticated ? (
              <>
                {isFrontPortal ? (
                  <Link className="user-chip" to="/account">
                    {user && "avatarUrl" in user && user.avatarUrl ? <img className="user-chip-avatar" src={user.avatarUrl} alt="" /> : null}
                    <span className="user-chip-label">当前用户</span>
                    <strong>{frontDisplayName}</strong>
                  </Link>
                ) : (
                  <span className="user-chip">
                    <span className="user-chip-label">后台账号</span>
                    <strong>{user?.username}</strong>
                  </span>
                )}
                <button className="ghost-button icon-button" type="button" onClick={logout} title="退出登录">
                  <LogOut size={18} aria-hidden="true" />
                  <span>退出</span>
                </button>
              </>
            ) : (
              <>
                {inAdmin ? (
                  <Link className="ghost-button" to="/admin/login" state={{ from: location }}>后台登录</Link>
                ) : (
                  <>
                    <Link className="ghost-button" to="/login" state={{ from: location }}>登录</Link>
                    <Link className="primary-button compact" to="/register">
                      <UserPlus size={16} aria-hidden="true" />
                      注册
                    </Link>
                  </>
                )}
              </>
            )}
          </div>
        </div>
      </header>
      <Outlet />
      <footer className="site-footer">
        <div className="shell site-footer-inner">
          <span>本地小说阅读</span>
          <span>仅使用公开接口与本地上传内容</span>
        </div>
      </footer>
    </div>
  );
}
