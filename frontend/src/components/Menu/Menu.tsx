import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "@/providers/AuthProvider";
import { HStack, Button, Text } from "@chakra-ui/react";

const ADMIN_ROLES = ["PLATFORM_ADMIN", "HR_ADMIN"];

export default function Menu() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const isAdmin = user && ADMIN_ROLES.includes(user.role);

  const handleLogout = () => {
    logout();
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
        {isAdmin && <Link to="/employees">Zaposleni</Link>}
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
