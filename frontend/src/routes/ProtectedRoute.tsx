import { Navigate, Outlet } from "react-router-dom";
import { useAuth } from "@/api/AuthContext";

type Props = {
  allowedRoles?: string[];
};

export const ProtectedRoute = ({ allowedRoles }: Props) => {
  const { isAuthenticated, user } = useAuth();

  if (!isAuthenticated) return <Navigate to="/login" replace />;

  if (allowedRoles && !allowedRoles.includes(user!.role))
    return <Navigate to="/unauthorized" replace />;

  return <Outlet />;
};
