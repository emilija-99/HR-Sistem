import { useEffect } from "react";
import { Navigate } from "react-router-dom";
import { useAuth } from "@/api/AuthContext";

const Logout = () => {
  const { logout } = useAuth();

  useEffect(() => {
    logout();
  }, []);

  return <Navigate to="/login" replace />;
};

export default Logout;
