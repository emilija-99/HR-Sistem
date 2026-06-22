import {
  createContext,
  useContext,
  useState,
  useEffect,
  useMemo,
  ReactNode,
  useCallback,
} from "react";
import { User, AuthContextType } from "../auth/types";
import { setToken as setClientToken } from "../api/client";

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [token, setToken] = useState<string | null>(null);
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true); // for initial refresh

  // Try silent refresh on mount
  useEffect(() => {
    const tryRefresh = async () => {
      try {
        const data = await fetch("/api/v1/refresh", {
          method: "POST",
          credentials: "include",
        }).then((r) => r.json());
        if (data.accessToken) {
          setToken(data.accessToken);
          setClientToken(data.accessToken);
        }
      } catch {
        /* not logged in */
      }
      setLoading(false);
    };
    tryRefresh();
  }, []);

  const login = useCallback(
    ({ token: t, user: u }: { token: string; user: User }) => {
      setToken(t);
      setClientToken(t);
      setUser(u);
    },
    [],
  );

  const logout = useCallback(() => {
    setToken(null);
    setClientToken(null);
    setUser(null);
  }, []);

  const isAuthenticated = !!token;

  const value = useMemo(
    () => ({ token, user, isAuthenticated, login, logout, loading }),
    [token, user, loading],
  );
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const useAuth = () => {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used inside AuthProvider");
  return ctx;
};
