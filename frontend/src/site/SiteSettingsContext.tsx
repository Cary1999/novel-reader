import { createContext, useContext, useEffect, useMemo, useState } from "react";
import { apiClient, ApiError } from "../api/client";
import type { SiteSettings } from "../api/types";
import { defaultSiteSettings } from "./defaults";

interface SiteSettingsContextValue {
  settings: SiteSettings;
  isLoading: boolean;
  error: string;
  refresh: () => Promise<SiteSettings>;
}

const SiteSettingsContext = createContext<SiteSettingsContextValue | null>(null);

export function SiteSettingsProvider({ children }: { children: React.ReactNode }) {
  const [settings, setSettings] = useState<SiteSettings>(defaultSiteSettings);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");

  async function refresh() {
    try {
      setError("");
      const next = await apiClient.siteSettings();
      setSettings({
        ...defaultSiteSettings,
        ...next,
        brandIconUrl: next.brandIconUrl || defaultSiteSettings.brandIconUrl,
      });
      return {
        ...defaultSiteSettings,
        ...next,
        brandIconUrl: next.brandIconUrl || defaultSiteSettings.brandIconUrl,
      };
    } catch (err) {
      setSettings(defaultSiteSettings);
      setError(err instanceof ApiError ? err.message : "站点设置加载失败");
      return defaultSiteSettings;
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    void refresh();
  }, []);

  const value = useMemo<SiteSettingsContextValue>(() => ({
    settings,
    isLoading,
    error,
    refresh,
  }), [error, isLoading, settings]);

  return <SiteSettingsContext.Provider value={value}>{children}</SiteSettingsContext.Provider>;
}

export function useSiteSettings() {
  const context = useContext(SiteSettingsContext);
  if (!context) {
    throw new Error("useSiteSettings must be used within SiteSettingsProvider");
  }
  return context;
}
