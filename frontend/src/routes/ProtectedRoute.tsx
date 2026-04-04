import { Navigate, Outlet } from "react-router-dom";
import { useAuth } from "@/providers/authProvider";

/*
  Protected Route component will serve as a wrapper for authnticated routes

  Outlet component acts as a placeholder to display the
  child component defined in the parent route.
*/

export const ProtectedRoute = () => {
  /*
  - check if the token exists, if the user is not authenticated,
  navigate component to redirect to the login page
  - if user is authenticated, render the child routes
  using the outlet component.
*/
  const { token } = useAuth();
  if (!token) {
    return <Navigate to="/login" />;
  }

  return <Outlet />;
};
