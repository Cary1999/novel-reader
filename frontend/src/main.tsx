import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import { App } from "./App";
import { AuthProvider } from "./auth/AuthContext";
import { SiteSettingsProvider } from "./site/SiteSettingsContext";
import "./styles.css";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <BrowserRouter>
      <SiteSettingsProvider>
        <AuthProvider>
          <App />
        </AuthProvider>
      </SiteSettingsProvider>
    </BrowserRouter>
  </React.StrictMode>,
);
