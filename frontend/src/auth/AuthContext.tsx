import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";
import { useLocation } from "react-router-dom";
import { adminTokenStore, apiClient, ApiError, frontTokenStore } from "../api/client";
import type { CurrentOperator, CurrentUser } from "../api/types";

type PortalScope = "front" | "admin";
type SessionAccount = CurrentUser | CurrentOperator;

interface AuthContextValue {
  user: SessionAccount | null;
  token: string | null;
  scope: PortalScope;
  isLoading: boolean;
  isAuthenticated: boolean;
  isAdminPortal: boolean;
  isFrontPortal: boolean;
  isAuthor: boolean;
  isReader: boolean;
  isReviewer: boolean;
  isSuperAdmin: boolean;
  login: (username: string, password: string) => Promise<void>;
  register: (username: string, password: string) => Promise<void>;
  logout: () => void;
  refreshUser: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

function isAdminPath(pathname: string) {
  return pathname.startsWith("/admin");
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const location = useLocation();
  const scope: PortalScope = isAdminPath(location.pathname) ? "admin" : "front";
  const tokenStore = scope === "admin" ? adminTokenStore : frontTokenStore;
  const [frontUser, setFrontUser] = useState<CurrentUser | null>(null);
  const [adminUser, setAdminUser] = useState<CurrentOperator | null>(null);
  const [isLoading, setIsLoading] = useState(Boolean(frontTokenStore.get() || adminTokenStore.get()));

  const token = tokenStore.get();
  const user = scope === "admin" ? adminUser : frontUser;

  const logout = useCallback(() => {
    if (scope === "admin") {
      adminTokenStore.clear();
      setAdminUser(null);
      return;
    }
    frontTokenStore.clear();
    setFrontUser(null);
  }, [scope]);

  const refreshUser = useCallback(async () => {
    const activeToken = tokenStore.get();
    if (!activeToken) {
      setIsLoading(false);
      if (scope === "admin") {
        setAdminUser(null);
      } else {
        setFrontUser(null);
      }
      return;
    }

    try {
      setIsLoading(true);
      if (scope === "admin") {
        setAdminUser(await apiClient.adminMe());
      } else {
        setFrontUser(await apiClient.me());
      }
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        logout();
      }
    } finally {
      setIsLoading(false);
    }
  }, [logout, scope, tokenStore]);

  useEffect(() => {
    void refreshUser();
  }, [refreshUser]);

  const login = useCallback(async (username: string, password: string) => {
    if (scope === "admin") {
      const response = await apiClient.adminLogin(username, password);
      adminTokenStore.set(response.token);
      setAdminUser(response.operator);
      return;
    }
    const response = await apiClient.login(username, password);
    frontTokenStore.set(response.token);
    setFrontUser(response.user);
  }, [scope]);

  const register = useCallback(async (username: string, password: string) => {
    if (scope === "admin") {
      throw new ApiError("后台账号只能由超级管理员创建", "FORBIDDEN", 403);
    }
    await apiClient.register(username, password);
    await login(username, password);
  }, [login, scope]);

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      token,
      scope,
      isLoading,
      isAuthenticated: Boolean(user && token),
      isAdminPortal: scope === "admin",
      isFrontPortal: scope === "front",
      isAuthor: scope === "front" && user?.role === "author",
      isReader: scope === "front" && user?.role === "reader",
      isReviewer: scope === "admin" && (user?.role === "reviewer" || user?.role === "super_admin"),
      isSuperAdmin: scope === "admin" && user?.role === "super_admin",
      login,
      register,
      logout,
      refreshUser,
    }),
    [isLoading, login, logout, refreshUser, register, scope, token, user],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const value = useContext(AuthContext);
  if (!value) {
    throw new Error("useAuth must be used inside AuthProvider");
  }
  return value;
}
