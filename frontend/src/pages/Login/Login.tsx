import { useState } from "react";
import { useNavigate, Link as RouterLink } from "react-router-dom";
import { useAuth } from "@/providers/AuthProvider";
import { api } from "@/api/client";
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
} from "@chakra-ui/react";

export default function LoginPage() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      const data = await api("/api/v1/login", {
        method: "POST",
        body: JSON.stringify({ email, password_hash: password }),
      });
      login({ token: data.data.accessToken, user: data.data.user });
      navigate("/home");
    } catch (err: any) {
      setError(err.message || "Login failed");
    } finally {
      setLoading(false);
    }
  };

  return (
    <Container maxW="sm" py={20}>
      <Card.Root>
        <Card.Header>
          <Heading size="lg" textAlign="center">
            HR Sistem — Prijava
          </Heading>
        </Card.Header>
        <Card.Body>
          <form onSubmit={handleSubmit}>
            <VStack gap={4}>
              <Field.Root required>
                <Field.Label>Email</Field.Label>
                <Input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="email@hr-sistem.com"
                />
              </Field.Root>
              <Field.Root required>
                <Field.Label>Lozinka</Field.Label>
                <Input
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="••••••••"
                />
              </Field.Root>
              {error && (
                <Text color="red.500" fontSize="sm">
                  {error}
                </Text>
              )}
              <Button
                type="submit"
                colorPalette="blue"
                width="full"
                loading={loading}
              >
                Prijavi se
              </Button>
            </VStack>
          </form>
        </Card.Body>
        <Card.Footer justifyContent="center">
          <Text fontSize="sm">
            Nemaš nalog?{" "}
            <ChakraLink asChild colorPalette="blue">
              <RouterLink to="/register">Registruj se</RouterLink>
            </ChakraLink>
          </Text>
        </Card.Footer>
      </Card.Root>
    </Container>
  );
}
