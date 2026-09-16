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
import { userFromToken } from "../auth/jwt";
import { setToken as setClientToken, getToken } from "../api/client";

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [token, setToken] = useState<string | null>(null);
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true); // for initial refresh

  // Try silent refresh on mount
  useEffect(() => {
    const tryRefresh = async () => {
      try {
        const res = await fetch("/api/v1/refresh", {
          method: "POST",
          credentials: "include",
        });
        if (res.ok) {
          const data = await res.json();
          if (data.accessToken) {
            setToken(data.accessToken);
            setClientToken(data.accessToken);
            // The user object is not persisted, so rebuild it from the token
            // claims (role is needed for role-based routing/menu).
            setUser(userFromToken(data.accessToken));
          }
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

  const logout = useCallback(async () => {
    try {
      // Revoke the refresh token server-side and clear the httpOnly cookie.
      const t = getToken();
      await fetch("/api/v1/logout", {
        method: "POST",
        credentials: "include",
        headers: t ? { Authorization: `Bearer ${t}` } : undefined,
      });
    } catch {
      /* best effort — local state is cleared regardless */
    }
    setToken(null);
    setClientToken(null);
    setUser(null);
  }, []);

  const isAuthenticated = !!token;

  const value = useMemo(
    () => ({ token, user, isAuthenticated, login, logout, loading }),
    [token, user, loading, logout],
  );
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const useAuth = () => {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used inside AuthProvider");
  return ctx;
};
