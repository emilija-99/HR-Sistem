import { Navigate, Outlet } from "react-router-dom";
import { useAuth } from "@/providers/AuthProvider";
import { Spinner, Center } from "@chakra-ui/react";

type Props = {
  allowedRoles?: string[];
};

export const ProtectedRoute = ({ allowedRoles }: Props) => {
  const { isAuthenticated, user, loading } = useAuth();

  // While the silent refresh is still resolving, show a spinner
  // instead of redirecting to /login (prevents bouncing users out).
  if (loading) {
    return (
      <Center h="100vh">
        <Spinner size="xl" />
      </Center>
    );
  }

  if (!isAuthenticated) return <Navigate to="/login" replace />;

  if (allowedRoles && !allowedRoles.includes(user!.role))
    return <Navigate to="/unauthorized" replace />;

  return <Outlet />;
};
