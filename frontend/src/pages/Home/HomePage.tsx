import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
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
  Separator,
} from "@chakra-ui/react";

export default function HomePage() {
  const { user } = useAuth();
  const navigate = useNavigate();
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
      <Container py={10} textAlign="center">
        <Spinner size="xl" />
      </Container>
    );

  if (!me) {
    navigate("/onboarding", { replace: true });
    return null;
  }

  return (
    <Container maxW="container.md" py={6}>
      <Heading mb={6}>Dobrodošli, {me.first_name}!</Heading>

      <Card.Root borderColor="brand.200" boxShadow="md">
        <Card.Header>
          <Heading size="md" color="brand.800">
            Moj profil
          </Heading>
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
                <Badge colorPalette="brand">{me.department_name}</Badge>
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
                          ? "brand"
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
                Nadređeni:
              </Text>
              <Text>{me.supervisor_name || "-"}</Text>
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
    </Container>
  );
}
