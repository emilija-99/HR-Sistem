// auth/AuthContext.jsx
import { createContext, useContext, useState, useCallback } from "react";

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [token, setToken] = useState(null); // in-memory only
  const [user, setUser] = useState(null);

  const login = async (email, password) => {
    const res = await fetch("/api/v1/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password_hash: password }),
    });
    const data = await res.json();
    setToken(data.data.accessToken); // ← memory only
    setUser(data.data.user);
  };

  const refreshToken = useCallback(async () => {
    const res = await fetch("/api/v1/refresh", {
      method: "POST",
      credentials: "include", // ← sends httpOnly cookie
    });
    const data = await res.json();
    setToken(data.accessToken);
    return data.accessToken;
  }, []);

  const logout = () => {
    setToken(null);
    setUser(null);
  };

  return (
    <AuthContext.Provider value={{ token, user, login, refreshToken, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export const useAuth = () => useContext(AuthContext);
