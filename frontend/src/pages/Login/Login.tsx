import { useState } from "react";
import { useNavigate, Link as RouterLink } from "react-router-dom";
import { useAuth } from "@/providers/AuthProvider";
import { api } from "@/api/client";
import { validateAccount, hasErrors } from "@/lib/validation";
import {
  Card,
  Input,
  Button,
  Text,
  Heading,
  VStack,
  Field,
  Link as ChakraLink,
  Container,
  Image,
  Box,
} from "@chakra-ui/react";

const BAD_CREDENTIALS =
  "Lozinka ili email adresa ne postoje. Molimo vas unesite ponovo.";

export default function LoginPage() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [submitted, setSubmitted] = useState(false);
  const [touched, setTouched] = useState<Record<string, boolean>>({});

  const errors = validateAccount(email, password);
  const markTouched = (field: string) =>
    setTouched((prev) => ({ ...prev, [field]: true }));
  const shown = (field: string) =>
    submitted || touched[field] ? errors[field] : "";

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitted(true);
    setError("");

    if (hasErrors(errors)) return;

    setLoading(true);
    try {
      const data = await api("/api/v1/login", {
        method: "POST",
        body: JSON.stringify({ email, password_hash: password }),
      });
      login({ token: data.accessToken, user: data.user });
      navigate("/home");
    } catch (err: any) {
      // Server vraća gotovu srpsku poruku (401: „Lozinka ili email adresa ne
      // postoje…“). Ako iz nekog razloga stigne prazna, koristimo rezervnu.
      setError(err?.message || BAD_CREDENTIALS);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Container maxW="sm" py={16}>
      <VStack gap={6} align="stretch">
        <Box textAlign="center">
          <Image src="/logo.png" alt="HR Sistem" mx="auto" maxH="120px" />
        </Box>

        <Card.Root borderColor="brand.200" boxShadow="md">
          <Card.Header>
            <Heading size="lg" textAlign="center" color="brand.800">
              Prijava
            </Heading>
            <Text fontSize="sm" color="fg.muted" textAlign="center" mt={1}>
              Dobrodošli nazad
            </Text>
          </Card.Header>
          <Card.Body>
            <form onSubmit={handleSubmit} noValidate>
              <VStack gap={4} align="stretch">
                <Field.Root required invalid={!!shown("email")}>
                  <Field.Label>Email</Field.Label>
                  <Input
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    onBlur={() => markTouched("email")}
                    placeholder="email@hr-sistem.com"
                  />
                  <Field.ErrorText>{shown("email")}</Field.ErrorText>
                </Field.Root>

                <Field.Root required invalid={!!shown("password")}>
                  <Field.Label>Lozinka</Field.Label>
                  <Input
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    onBlur={() => markTouched("password")}
                    placeholder="••••••••"
                  />
                  <Field.ErrorText>{shown("password")}</Field.ErrorText>
                </Field.Root>

                {error && (
                  <Text color="red.500" fontSize="sm">
                    {error}
                  </Text>
                )}

                <Button
                  type="submit"
                  colorPalette="brand"
                  width="full"
                  loading={loading}
                >
                  Prijavi se
                </Button>
              </VStack>
            </form>
          </Card.Body>
          <Card.Footer justifyContent="center">
            <Text fontSize="sm" color="fg.muted">
              Nemaš nalog?{" "}
              <ChakraLink asChild colorPalette="brand" fontWeight="600">
                <RouterLink to="/register">Registruj se</RouterLink>
              </ChakraLink>
            </Text>
          </Card.Footer>
        </Card.Root>
      </VStack>
    </Container>
  );
}
