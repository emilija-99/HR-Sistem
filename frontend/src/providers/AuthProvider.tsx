import {
  createContext,
  useContext,
  useState,
  useEffect,
  useMemo,
  ReactNode,
} from "react";
import { User, AuthContextType } from "../auth/types";
import axios from "axios";
const AuthContext = createContext<AuthContextType | undefined>(undefined);

type Props = { children: ReactNode };

export const AuthProvider = ({ children }: Props) => {
  /* local storage  */
  const [token, setToken] = useState<string | null>(
    localStorage.getItem("token"),
  );
  // const [token, setToken] = useState<string | null>(null);

  const [user, setUser] = useState<User | null>(() => {
    const stored = localStorage.getItem("user");
    return stored ? JSON.parse(stored) : null;
  });

  const isAuthenticated = !!token && !!user;

  const login = ({ token, user }: { token: string; user: User }) => {
    setToken(token);
    setUser(user);
  };

  const logout = () => {
    setToken(null);
    setUser(null);
  };

  // attach token to axios
  useEffect(() => {
    if (token) {
      axios.defaults.headers.common["Authorization"] = "Bearer " + token;
      localStorage.setItem("token", token);
    } else {
      delete axios.defaults.headers.common["Authorization"];
      localStorage.removeItem("token");
    }
  }, [token]);
  // useEffect(() => {
  //   const bootstrapAuth = async () => {
  //     try {
  //       const res = await axios.post("/refresh", {}, { withCredentials: true });
  //       setToken(res.data.accessToken);
  //     } catch {
  //       logout();
  //     }
  //   };
  //   bootstrapAuth();
  // }, []);

  // persist user
  useEffect(() => {
    if (user) localStorage.setItem("user", JSON.stringify(user));
    else localStorage.removeItem("user");
  }, [user]);

  const value = useMemo(
    () => ({ token, user, isAuthenticated, login, logout }),
    [token, user],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};
export const useAuth = () => {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used inside AuthProvider");
  return ctx;
};
