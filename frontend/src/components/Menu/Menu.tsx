import { Link, useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "@/providers/AuthProvider";
import { HStack, Button, Text, Badge, Box } from "@chakra-ui/react";
import Brand from "@/components/Brand/Brand";

const ADMIN_ROLES = ["PLATFORM_ADMIN", "HR_ADMIN"];
const APPROVER_ROLES = ["PLATFORM_ADMIN", "HR_ADMIN", "MANAGER_PORTAL_ACCESS"];

type NavItem = { to: string; label: string };

export default function Menu() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const { pathname } = useLocation();

  const isAdmin = !!user && ADMIN_ROLES.includes(user.role);
  const isApprover = !!user && APPROVER_ROLES.includes(user.role);
  const isPlatformAdmin = user?.role === "PLATFORM_ADMIN";

  const items: NavItem[] = [
    { to: "/home", label: "Početna" },
    { to: "/absences", label: "Odsustva" },
    { to: "/balance", label: "Dostupni dani" },
    {
      to: isAdmin ? "/attendance/admin" : "/attendance",
      label: "Prisustvo",
    },
    { to: "/profile", label: "Moj profil" },
    ...(isAdmin ? [{ to: "/employees", label: "Zaposleni" }] : []),
    ...(isApprover
      ? [{ to: "/absences/approvals", label: "Odobravanja" }]
      : []),
    ...(isAdmin ? [{ to: "/admin/audit", label: "Audit log" }] : []),
    ...(isPlatformAdmin ? [{ to: "/admin/users", label: "Korisnici" }] : []),
  ];

  // Najduža putanja koja odgovara trenutnoj ruti je aktivna (npr. na
  // /absences/approvals ne želimo da bude aktivna i /absences).
  const activeTo = items
    .filter((i) => pathname === i.to || pathname.startsWith(`${i.to}/`))
    .sort((a, b) => b.to.length - a.to.length)[0]?.to;

  const handleLogout = async () => {
    await logout();
    navigate("/login");
  };

  return (
    <HStack
      px={6}
      py={3}
      bg="bg.panel"
      borderBottomWidth="1px"
      borderColor="brand.200"
      justify="space-between"
      gap={6}
      wrap="wrap"
      position="sticky"
      top={0}
      zIndex={10}
    >
      <HStack gap={6} wrap="wrap">
        <Link to="/home" style={{ textDecoration: "none" }}>
          <Brand fontSize="xl" />
        </Link>
        <HStack gap={1} wrap="wrap">
          {items.map((item) => {
            const active = activeTo === item.to;
            return (
              <Link key={item.to} to={item.to} style={{ textDecoration: "none" }}>
                <Text
                  px={3}
                  py={2}
                  borderRadius="md"
                  fontSize="sm"
                  whiteSpace="nowrap"
                  fontWeight={active ? 600 : 500}
                  color={active ? "brand.600" : "fg.muted"}
                  bg={active ? "brand.subtle" : "transparent"}
                  transition="background 0.15s ease, color 0.15s ease"
                  _hover={{ color: "brand.600", bg: "brand.subtle" }}
                >
                  {item.label}
                </Text>
              </Link>
            );
          })}
        </HStack>
      </HStack>

      <HStack gap={3}>
        <Box textAlign="right">
          <Text fontSize="sm" color="fg.muted" lineHeight="short">
            {user?.email}
          </Text>
          <Badge colorPalette="brand" size="sm">
            {user?.role}
          </Badge>
        </Box>
        <Button
          size="sm"
          variant="outline"
          colorPalette="brand"
          onClick={handleLogout}
        >
          Odjavi se
        </Button>
      </HStack>
    </HStack>
  );
}
