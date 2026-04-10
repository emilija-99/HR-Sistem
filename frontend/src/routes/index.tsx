import { RouterProvider, createBrowserRouter } from "react-router-dom";
import { ProtectedRoute } from "./ProtectedRoute";
import HomePage from "@/pages/Home/HomePage";
import Login from "@/pages/Login/Login";
import Logout from "@/routes/Logout";

const router = createBrowserRouter([
  { path: "/login", element: <Login /> },

  {
    element: <ProtectedRoute />, // authenticated users
    children: [
      { path: "/home", element: <HomePage /> },
      { path: "/profile", element: <div>Profile</div> },
      { path: "/logout", element: <Logout /> },
    ],
  },

  {
    element: <ProtectedRoute allowedRoles={["admin"]} />,
    children: [{ path: "/admin", element: <div>Admin page</div> }],
  },

  { path: "/unauthorized", element: <div>Unauthorized</div> },
]);

export default function Routes() {
  return <RouterProvider router={router} />;
}
