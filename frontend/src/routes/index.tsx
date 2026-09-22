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
import NewEmployeePage from "@/pages/Employees/NewEmployeePage";
import AbsencesPage from "@/pages/Absences/AbsencesPage";
import AbsenceApprovalsPage from "@/pages/Absences/AbsenceApprovalsPage";
import BalancePage from "@/pages/Balance/BalancePage";
import AttendancePage from "@/pages/Attendance/AttendancePage";
import AttendanceAdminPage from "@/pages/Attendance/AttendanceAdminPage";
import UsersPage from "@/pages/Admin/UsersPage";
import AuditLogPage from "@/pages/Admin/AuditLogPage";
import Logout from "@/routes/Logout";

const router = createBrowserRouter([
  { path: "/", element: <Navigate to="/login" replace /> },
  { path: "/login", element: <LoginPage /> },
  { path: "/register", element: <RegisterPage /> },
  {
    element: <ProtectedRoute />,
    children: [
      { path: "/home", element: <HomePage /> },
      { path: "/profile", element: <EmployeeDetailPage me /> },
      { path: "/onboarding", element: <OnboardingPage /> },
      { path: "/absences", element: <AbsencesPage /> },
      { path: "/balance", element: <BalancePage /> },
      { path: "/attendance", element: <AttendancePage /> },
      { path: "/logout", element: <Logout /> },
      {
        element: (
          <ProtectedRoute allowedRoles={["PLATFORM_ADMIN"]} />
        ),
        children: [{ path: "/admin/users", element: <UsersPage /> }],
      },
      {
        element: (
          <ProtectedRoute
            allowedRoles={[
              "PLATFORM_ADMIN",
              "HR_ADMIN",
              "MANAGER_PORTAL_ACCESS",
            ]}
          />
        ),
        children: [
          { path: "/absences/approvals", element: <AbsenceApprovalsPage /> },
        ],
      },
      {
        element: (
          <ProtectedRoute allowedRoles={["PLATFORM_ADMIN", "HR_ADMIN"]} />
        ),
        children: [
          { path: "/employees", element: <EmployeesPage /> },
          { path: "/employees/new", element: <NewEmployeePage /> },
          { path: "/employees/:id", element: <EmployeeDetailPage /> },
          { path: "/attendance/admin", element: <AttendanceAdminPage /> },
          { path: "/admin/audit", element: <AuditLogPage /> },
        ],
      },
    ],
  },
  { path: "/unauthorized", element: <div>Nemate dozvolu za pristup ovoj stranici.</div> },
]);

export default function Routes() {
  return <RouterProvider router={router} />;
}
