import { useAuth } from "../providers/AuthProvider";
import { useEffect } from "react";
import { useNavigate } from "react-router-dom";

export default function Logout() {
  const { logout } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    // Revokes the refresh token server-side, then clears local state.
    logout().finally(() => navigate("/login", { replace: true }));
  }, []);

  return null;
}
