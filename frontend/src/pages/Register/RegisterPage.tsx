import { useState } from "react";
import { useNavigate, Link as RouterLink } from "react-router-dom";
import { api } from "@/api/client";
import {
  Container,
  Card,
  Heading,
  VStack,
  Field,
  Input,
  Button,
  Text,
  Link as ChakraLink,
} from "@chakra-ui/react";

export default function RegisterPage() {
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError("");
    if (password !== confirmPassword) {
      setError("Lozinke se ne poklapaju");
      setLoading(false);
      return;
    }
    try {
      await api("/api/v1/register", {
        method: "POST",
        body: JSON.stringify({ email, password_hash: password }),
      });
      navigate("/login", { state: { registered: true } });
    } catch (err: any) {
      setError(err.message || "Registracija nije uspela");
    } finally {
      setLoading(false);
    }
  };

  return (
    <Container maxW="sm" py={20}>
      <Card.Root>
        <Card.Header>
          <Heading size="lg" textAlign="center">
            HR Sistem — Registracija
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
                  placeholder="min 8 karaktera"
                />
              </Field.Root>
              <Field.Root required>
                <Field.Label>Potvrdi lozinku</Field.Label>
                <Input
                  type="password"
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  placeholder="ponovi lozinku"
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
                Registruj se
              </Button>
            </VStack>
          </form>
        </Card.Body>
        <Card.Footer justifyContent="center">
          <Text fontSize="sm">
            Već imaš nalog?{" "}
            <ChakraLink asChild colorPalette="blue">
              <RouterLink to="/login">Prijavi se</RouterLink>
            </ChakraLink>
          </Text>
        </Card.Footer>
      </Card.Root>
    </Container>
  );
}
