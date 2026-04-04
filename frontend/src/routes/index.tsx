import { RouterProvider, createBrowserRouter } from "react-router-dom";
import { useAuth } from "../providers/authProvider";
import { ProtectedRoute } from "./ProtectedRoute";
import HomePage from "@/pages/Home/HomePage";
import type { ReactElement } from "react";
import Login from "@/pages/Login/Login";
// type AppRoute = {
//   path: string;
//   element: ReactElement;
// };

const Routes = () => {
  const { token } = useAuth();

  const PublicRoutes = [
    {
      path: "/login",
      element: <Login />,
    },
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
          path: "/home",
          element: <HomePage />,
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
      path: "/login",
      element: <Login />,
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
