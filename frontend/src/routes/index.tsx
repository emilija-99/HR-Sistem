import {
  RouterProvider,
  createBrowserRouter,
  Navigate,
} from "react-router-dom";
import { ProtectedRoute } from "./ProtectedRoute";
import HomePage from "@/pages/Home/HomePage";
import LoginPage from "@/pages/Login/Login";
import RegisterPage from "@/pages/Register/RegisterPage";
import OnboardingPage from "@/pages/Onboarding/OnboardingPage";
import EmployeesPage from "@/pages/Employees/EmployeesPage";
import EmployeeDetailPage from "@/pages/Employees/EmployeeDetailPage";
import Logout from "@/routes/Logout";

const router = createBrowserRouter([
  { path: "/", element: <Navigate to="/login" replace /> },
  { path: "/login", element: <LoginPage /> },
  { path: "/register", element: <RegisterPage /> },
  {
    element: <ProtectedRoute />,
    children: [
      { path: "/home", element: <HomePage /> },
      { path: "/onboarding", element: <OnboardingPage /> },
      {
        element: (
          <ProtectedRoute allowedRoles={["PLATFORM_ADMIN", "HR_ADMIN"]} />
        ),
        children: [
          { path: "/employees", element: <EmployeesPage /> },
          { path: "/employees/:id", element: <EmployeeDetailPage /> },
        ],
      },
      { path: "/logout", element: <Logout /> },
    ],
  },
  { path: "/unauthorized", element: <div>Unauthorized</div> },
]);

export default function Routes() {
  return <RouterProvider router={router} />;
}
