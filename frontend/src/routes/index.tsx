import { RouterProvider, createBrowserRouter } from "react-router-dom";
import { useAuth } from "../providers/authProvider";
import { ProtectedRoute } from "./ProtectedRoute";
import HomePage from "@/pages/Home/HomePage";
import type { ReactElement } from "react";
// type AppRoute = {
//   path: string;
//   element: ReactElement;
// };

const Routes = () => {
  const { token } = useAuth();

  const PublicRoutes = [
    {
      path: "/service",
      element: <div>Service</div>,
    },
    {
      path: "/about-us",
      element: <div>About Us</div>,
    },
  ];

  const AuthenticatedOnlyRoutes = [
    {
      path: "/",
      element: <ProtectedRoute />,
      children: [
        {
          path: "/",
          element: HomePage,
        },
        {
          path: "/profile",
          element: <div>User Profile</div>,
        },
        {
          path: "/logout",
          element: <div>Logout</div>,
        },
      ],
    },
  ];

  const NotAuthenticatedOnlyRoutes = [
    {
      path: "/",
      element: <div>Home Page</div>,
    },
    {
      path: "/login",
      element: <div>Login</div>,
    },
  ];
  const router = createBrowserRouter([
    ...PublicRoutes,
    ...(!token ? NotAuthenticatedOnlyRoutes : []),
    ...AuthenticatedOnlyRoutes,
  ]);
  return <RouterProvider router={router} />;
};
export default Routes;
