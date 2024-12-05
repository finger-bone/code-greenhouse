import axios from "axios";
import { createContext, useContext, useEffect, useState } from "react";

type TokenProviderProps = {
  children: React.ReactNode;
  storageKey?: string;
  providerStorageKey?: string;
  subjectStorageKey?: string;
};

type TokenProviderState = {
  token: string | null;
  provider: string | null;
  subject: string | null;
  singleUser: boolean;
  setToken: (token: string | null) => void;
  setProvider: (provider: string | null) => void;
  setSubject: (subject: string | null) => void;
  setSingleUser: (singleUser: boolean) => void;
};

const initialState: TokenProviderState = {
  token: null,
  provider: null,
  subject: null,
  singleUser: false,
  setToken: () => null,
  setProvider: () => null,
  setSubject: () => null,
  setSingleUser: () => null,
};

const TokenProviderContext = createContext<TokenProviderState>(initialState);

export function TokenProvider({
  children,
  storageKey = "auth-token",
  providerStorageKey = "provider",
  subjectStorageKey = "subject",
  ...props
}: TokenProviderProps) {
  const [token, setTokenState] = useState<string | null>(() =>
    localStorage.getItem(storageKey)
  );

  const handleSetToken = (newToken: string | null) => {
    if (newToken) {
      localStorage.setItem(storageKey, newToken);
    } else {
      localStorage.removeItem(storageKey);
    }
    setTokenState(newToken);
  };

  const [provider, setProviderState] = useState<string | null>(() =>
    localStorage.getItem(providerStorageKey)
  );

  const handleSetProvider = (newProvider: string | null) => {
    if (newProvider) {
      localStorage.setItem(providerStorageKey, newProvider);
    } else {
      localStorage.removeItem(providerStorageKey);
    }
    setProviderState(newProvider);
  };

  const [subject, setSubjectState] = useState<string | null>(() =>
    localStorage.getItem(subjectStorageKey)
  );

  const handleSetSubject = (newSubject: string | null) => {
    if (newSubject) {
      localStorage.setItem(subjectStorageKey, newSubject);
    } else {
      localStorage.removeItem(subjectStorageKey);
    }
    setSubjectState(newSubject);
  };

  const [singleUser, setSingleUser] = useState<boolean>(false);

  const value = {
    token,
    provider,
    subject,
    singleUser,
    setToken: handleSetToken,
    setProvider: handleSetProvider,
    setSubject: handleSetSubject,
    setSingleUser,
  };

  return (
    <TokenProviderContext.Provider {...props} value={value}>
      {children}
    </TokenProviderContext.Provider>
  );
}

export const useAuthToken = () => {
  const context = useContext(TokenProviderContext);

  if (context === undefined) {
    throw new Error("useAuth must be used within an TokenProvider");
  }

  return context;
};
