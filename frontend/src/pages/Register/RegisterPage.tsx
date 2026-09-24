import { useState } from "react";
import { useNavigate, Link as RouterLink } from "react-router-dom";
import { api } from "@/api/client";
import { validateAccount, hasErrors, FORM_INCOMPLETE } from "@/lib/validation";
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
  Image,
  Box,
} from "@chakra-ui/react";

export default function RegisterPage() {
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState("");
  // Greška koju vraća server za konkretno polje (npr. 409 — email već postoji).
  const [emailError, setEmailError] = useState("");
  const [loading, setLoading] = useState(false);
  const [submitted, setSubmitted] = useState(false);
  const [touched, setTouched] = useState<Record<string, boolean>>({});

  const errors = validateAccount(email, password, confirmPassword);
  const markTouched = (field: string) =>
    setTouched((prev) => ({ ...prev, [field]: true }));
  // Greška se prikazuje tek kad je polje „dirano" ili je forma poslata —
  // da korisnik ne vidi crveno dok tek počinje da kuca.
  const shown = (field: string) =>
    submitted || touched[field] ? errors[field] : "";
  // Greška sa servera ima prednost nad klijentskom (email je sintaksno ispravan,
  // ali je zauzet).
  const emailFieldError = shown("email") || emailError;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitted(true);
    setError("");
    setEmailError("");

    if (hasErrors(errors)) return;

    setLoading(true);
    try {
      await api("/api/v1/register", {
        method: "POST",
        body: JSON.stringify({ email, password_hash: password }),
      });
      navigate("/login", { state: { registered: true } });
    } catch (err: any) {
      // 409 = email je već zauzet → greška ide na polje Email, ne globalno.
      if (err?.status === 409) setEmailError(err.message);
      else setError(err.message || "Registracija nije uspela");
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
              Registracija
            </Heading>
            <Text fontSize="sm" color="fg.muted" textAlign="center" mt={1}>
              Napravi nalog zaposlenog
            </Text>
          </Card.Header>
          <Card.Body>
            <form onSubmit={handleSubmit} noValidate>
              <VStack gap={4} align="stretch">
                <Field.Root required invalid={!!emailFieldError}>
                  <Field.Label>Email</Field.Label>
                  <Input
                    type="email"
                    value={email}
                    onChange={(e) => {
                      setEmail(e.target.value);
                      setEmailError("");
                    }}
                    onBlur={() => markTouched("email")}
                    placeholder="email@hr-sistem.com"
                  />
                  <Field.ErrorText>{emailFieldError}</Field.ErrorText>
                </Field.Root>

                <Field.Root required invalid={!!shown("password")}>
                  <Field.Label>Lozinka</Field.Label>
                  <Input
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    onBlur={() => markTouched("password")}
                    placeholder="min 8 karaktera"
                  />
                  <Field.ErrorText>{shown("password")}</Field.ErrorText>
                </Field.Root>

                <Field.Root required invalid={!!shown("confirm")}>
                  <Field.Label>Potvrdi lozinku</Field.Label>
                  <Input
                    type="password"
                    value={confirmPassword}
                    onChange={(e) => setConfirmPassword(e.target.value)}
                    onBlur={() => markTouched("confirm")}
                    placeholder="ponovi lozinku"
                  />
                  <Field.ErrorText>{shown("confirm")}</Field.ErrorText>
                </Field.Root>

                {submitted && hasErrors(errors) && (
                  <Text color="red.500" fontSize="sm">
                    {FORM_INCOMPLETE}
                  </Text>
                )}
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
                  Registruj se
                </Button>
              </VStack>
            </form>
          </Card.Body>
          <Card.Footer justifyContent="center">
            <Text fontSize="sm" color="fg.muted">
              Već imaš nalog?{" "}
              <ChakraLink asChild colorPalette="brand" fontWeight="600">
                <RouterLink to="/login">Prijavi se</RouterLink>
              </ChakraLink>
            </Text>
          </Card.Footer>
        </Card.Root>
      </VStack>
    </Container>
  );
}
