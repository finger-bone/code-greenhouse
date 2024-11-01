import { createContext, useContext, useState } from "react";

type StateProviderProps = {
  children: React.ReactNode;
  initialValue?: string;
  storageKey?: string;
};

type StateProviderState = {
  value: string | null;
  setValue: (newValue: string | null) => void;
};

const initialState: StateProviderState = {
  value: null,
  setValue: () => null,
};

const AuthStateProviderContext = createContext<StateProviderState>(
  initialState,
);

export function AuthStateProvider({
  children,
  initialValue = "",
  storageKey = "auth-state",
  ...props
}: StateProviderProps) {
  const [value, setValue] = useState<string | null>(() =>
    localStorage.getItem(storageKey) || initialValue
  );

  const handleSetValue = (newValue: string | null) => {
    if (newValue) {
      localStorage.setItem(storageKey, newValue);
    } else {
      localStorage.removeItem(storageKey);
    }
    setValue(newValue);
  };

  const contextValue = {
    value,
    setValue: handleSetValue,
  };

  return (
    <AuthStateProviderContext.Provider {...props} value={contextValue}>
      {children}
    </AuthStateProviderContext.Provider>
  );
}

export const useAuthState = () => {
  const context = useContext(AuthStateProviderContext);

  if (context === undefined) {
    throw new Error("useAppState must be used within a StateProvider");
  }

  return context;
};
