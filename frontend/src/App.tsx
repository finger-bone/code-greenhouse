import en from "@/locales/en.json";
import zh_hans from "@/locales/zh-hans.json";
import axios from "axios";
import i18n from "i18next";
import { useEffect } from "react";
import { initReactI18next, useTranslation } from "react-i18next";
import { Route, Routes, useLocation, useNavigate } from "react-router-dom";
import Header from "./components/layouts/header";
import { useAuthState } from "./providers/auth-state-provider";
import { useAuthToken } from "./providers/token-provider";
import { lazy, Suspense } from 'react';

// i18n configuration remains the same
i18n.use(initReactI18next).init({
  resources: {
    en: { translation: en },
    zh_hans: { translation: zh_hans },
  },
  lng: "en",
  fallbackLng: "en",
  interpolation: {
    escapeValue: false,
  },
});

async function getToken(authorizationCode: string, provider: string) {
  const params = {
    code: authorizationCode,
    provider: provider,
  };

  const response = await axios.get("/api/auth/token", { params });
  return response.data.data.accessToken as string;
}

function App() {
  const auth = useAuthToken();
  const authState = useAuthState();
  const navigate = useNavigate();

  const handleAuthError = (query: URLSearchParams) => {
    if (query.get("error")) {
      auth.setProvider(null);
      auth.setToken(null);
      auth.setSubject(null);
      navigate("/login");
      return true;
    }
    return false;
  };

  const handleAuthState = (query: URLSearchParams) => {
    if (authState.value == query.get("state")) {
      auth.setToken(null);
      authState.setValue(null);
      return true;
    }
    navigate("/login");
    return false;
  };

  const getSubject = async (token: string, provider: string) => {
    const resp = await axios.get(`/api/user/subject?provider=${provider}`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });
    console.log(resp.data);
    return resp.data.data.subject;
  };

  const handleAuthCode = async (
    authorizationCode: string | null,
    provider: string | null,
  ) => {
    if (authorizationCode && provider) {
      const token = await getToken(authorizationCode, provider);
      auth.setToken(token);
      const subject = await getSubject(token, provider);
      auth.setSubject(subject);
    }
  };

  const setupHeader = (token: string, provider: string) => {
    axios.defaults.headers.common["Authorization"] = `Bearer ${token}`;
    axios.defaults.headers["Provider"] = provider;
  };

  const checkTokenValidity = async () => {
    const resp = await axios.get(`/api/ping/auth`);
    if (resp.status != 200) {
      auth.setToken(null);
      auth.setProvider(null);
      auth.setSubject(null);
      navigate("/login");
    }
  };

  const singleUserMode = async () => {
    const resp = await axios.get(`/api/auth/single-user`);
    return resp.data.data.enabled as boolean;
  };

  const SINGLE_USER_PROVIDER = "localhost";
  const SINGLE_USER_SUBJECT = "subject";

  useEffect(() => {
    singleUserMode().then(
      (enabled: boolean) => {
        if (enabled) {
          setupHeader("", SINGLE_USER_PROVIDER);
          auth.setSubject(SINGLE_USER_SUBJECT);
          auth.setProvider(SINGLE_USER_PROVIDER);
          auth.setSingleUser(true);
          checkTokenValidity();
        } else {
          auth.setSingleUser(false);
          if (auth.token) {
            setupHeader(auth.token, auth.provider!);
            checkTokenValidity();
            return;
          }

          const query = new URLSearchParams(window.location.search);

          if (handleAuthError(query)) return;

          const authorizationCode = query.get("code");
          if (!handleAuthState(query)) return;

          handleAuthCode(authorizationCode, auth.provider);
        }
      },
    );
  }, [useLocation()]);

  const ChallengePage = lazy(() => import("./pages/challenge-page"));
  const LoginPage = lazy(() => import("./pages/login-page"));
  const RepositoriesPage = lazy(() => import("./pages/repositories-page"));
  const RepositoryPage = lazy(() => import("./pages/repository-page"));
  const UserPage = lazy(() => import("./pages/user-page"));

  return (
    <>
      <div className="flex flex-col min-h-screen w-full items-center">
        <Header
          items={[
            { title: "Challenge", href: "/challenge" },
            { title: "Repositories", href: "/repositories" },
            { title: "User", href: "/user" },
          ]}
        />
        <Routes>
          <Route path="/" element={<div>Change the tab!</div>} />
          <Route path="/challenge" element={
            <Suspense fallback={<div>Loading...</div>}>
              <ChallengePage />
            </Suspense>
          } />
          <Route path="/repository/:repoId" element={
            <Suspense fallback={<div>Loading...</div>}>
              <RepositoryPage />
            </Suspense>
          } />
          <Route path="/repositories" element={
            <Suspense fallback={<div>Loading...</div>}>
              <RepositoriesPage />
            </Suspense>
          } />
          <Route path="/user" element={
            <Suspense fallback={<div>Loading...</div>}>
              <UserPage />
            </Suspense>
          } />
          <Route path="/about" element={<div>about</div>} />
          <Route path="/login" element={
            <Suspense fallback={<div>Loading...</div>}>
              <LoginPage />
            </Suspense>
          } />
        </Routes>
      </div>
    </>
  );
}

export default App;
