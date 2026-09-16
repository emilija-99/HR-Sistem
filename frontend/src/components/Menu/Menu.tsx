import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "@/providers/AuthProvider";
import { HStack, Button, Text } from "@chakra-ui/react";

const ADMIN_ROLES = ["PLATFORM_ADMIN", "HR_ADMIN"];
const APPROVER_ROLES = ["PLATFORM_ADMIN", "HR_ADMIN", "MANAGER_PORTAL_ACCESS"];

export default function Menu() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const isAdmin = user && ADMIN_ROLES.includes(user.role);
  const isApprover = user && APPROVER_ROLES.includes(user.role);
  const isPlatformAdmin = user?.role === "PLATFORM_ADMIN";

  const handleLogout = async () => {
    await logout();
    navigate("/login");
  };

  return (
    <HStack
      p={4}
      borderBottom="1px solid"
      borderColor="gray.200"
      justify="space-between"
    >
      <HStack gap={6}>
        <Text fontWeight="bold" fontSize="lg">
          HR Sistem
        </Text>
        <Link to="/home">Početna</Link>
        <Link to="/absences">Odsustva</Link>
        <Link to="/balance">Dostupni dani</Link>
        {isAdmin ? (
          <Link to="/attendance/admin">Prisustvo</Link>
        ) : (
          <Link to="/attendance">Prisustvo</Link>
        )}
        <Link to="/profile">Moj profil</Link>
        {isAdmin && <Link to="/employees">Zaposleni</Link>}
        {isApprover && <Link to="/absences/approvals">Odobravanja</Link>}
        {isAdmin && <Link to="/admin/audit">Audit log</Link>}
        {isPlatformAdmin && <Link to="/admin/users">Korisnici</Link>}
      </HStack>
      <HStack gap={4}>
        <Text fontSize="sm" color="gray.500">
          {user?.email} ({user?.role})
        </Text>
        <Button size="sm" variant="outline" onClick={handleLogout}>
          Odjavi se
        </Button>
      </HStack>
    </HStack>
  );
}
