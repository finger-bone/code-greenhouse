import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { HashRouter } from "react-router-dom";
import "./index.css";
import App from "./App.tsx";
import { AuthStateProvider } from "./providers/auth-state-provider.tsx";
import { ThemeProvider } from "./providers/theme-provider.tsx";
import { TokenProvider } from "./providers/token-provider.tsx";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <HashRouter>
      <ThemeProvider defaultTheme="system" storageKey="vite-ui-theme">
        <TokenProvider>
          <AuthStateProvider>
            <App />
          </AuthStateProvider>
        </TokenProvider>
      </ThemeProvider>
    </HashRouter>
  </StrictMode>,
);
