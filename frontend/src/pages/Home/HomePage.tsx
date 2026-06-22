import { useState, useEffect } from "react";
import { Link as RouterLink } from "react-router-dom";
import Menu from "@/components/Menu/Menu";
import { api } from "@/api/client";
import { useAuth } from "@/providers/AuthProvider";
import {
  Container,
  Heading,
  Text,
  VStack,
  HStack,
  Card,
  Badge,
  Spinner,
  Button,
  SimpleGrid,
  Separator,
} from "@chakra-ui/react";

const ADMIN_ROLES = ["PLATFORM_ADMIN", "HR_ADMIN"];

export default function HomePage() {
  const { user } = useAuth();
  const isAdmin = user && ADMIN_ROLES.includes(user.role);
  const [me, setMe] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api("/api/v1/employees/me")
      .then(setMe)
      .catch(() => {
        /* no profile yet */
      })
      .finally(() => setLoading(false));
  }, []);

  if (loading)
    return (
      <>
        <Menu />
        <Container py={10} textAlign="center">
          <Spinner size="xl" />
        </Container>
      </>
    );

  if (!me) {
    window.location.href = "/onboarding";
    return null;
  }

  return (
    <>
      <Menu />
      <Container maxW="container.xl" py={6}>
        <Heading mb={6}>Dobrodošli, {me.first_name}!</Heading>
        <SimpleGrid columns={{ base: 1, md: 2 }} gap={6}>
          <Card.Root>
            <Card.Header>
              <Heading size="md">Moj profil</Heading>
            </Card.Header>
            <Card.Body>
              <VStack align="stretch" gap={3}>
                <HStack>
                  <Text fontWeight="bold" w={32}>
                    Ime:
                  </Text>
                  <Text>
                    {me.first_name} {me.last_name}
                  </Text>
                </HStack>
                <Separator />
                <HStack>
                  <Text fontWeight="bold" w={32}>
                    Email:
                  </Text>
                  <Text>{user?.email}</Text>
                </HStack>
                <Separator />
                <HStack>
                  <Text fontWeight="bold" w={32}>
                    Departman:
                  </Text>
                  {me.department_name ? (
                    <Badge colorPalette="purple">{me.department_name}</Badge>
                  ) : (
                    <Text>-</Text>
                  )}
                </HStack>
                <Separator />
                <HStack>
                  <Text fontWeight="bold" w={32}>
                    Pozicija:
                  </Text>
                  <Text>{me.position_title || "-"}</Text>
                  {me.position_level && (
                    <Badge
                      colorPalette={
                        me.position_level === "LEAD"
                          ? "orange"
                          : me.position_level === "SENIOR"
                            ? "red"
                            : me.position_level === "MEDIOR"
                              ? "blue"
                              : "gray"
                      }
                    >
                      {me.position_level}
                    </Badge>
                  )}
                </HStack>
                <Separator />
                <HStack>
                  <Text fontWeight="bold" w={32}>
                    Grad:
                  </Text>
                  <Text>{me.city || "-"}</Text>
                </HStack>
                <Separator />
                <HStack>
                  <Text fontWeight="bold" w={32}>
                    Datum zaposlenja:
                  </Text>
                  <Text>{me.hire_date || "-"}</Text>
                </HStack>
              </VStack>
            </Card.Body>
          </Card.Root>
          <Card.Root>
            <Card.Header>
              <Heading size="md">Brzi linkovi</Heading>
            </Card.Header>
            <Card.Body>
              <VStack gap={3} align="stretch">
                {isAdmin && (
                  <RouterLink to="/employees">
                    <Button
                      variant="outline"
                      justifyContent="start"
                      width="full"
                    >
                      👥 Pregled zaposlenih
                    </Button>
                  </RouterLink>
                )}
                <RouterLink to={`/employees/${me.id}`}>
                  <Button variant="outline" justifyContent="start" width="full">
                    👤 Moj profil (detalji)
                  </Button>
                </RouterLink>
              </VStack>
            </Card.Body>
          </Card.Root>
        </SimpleGrid>
      </Container>
    </>
  );
}
