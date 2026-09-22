import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "@/api/client";
import {
  validateProfile,
  hasErrors,
  PROFILE_INCOMPLETE,
  MAX_NAME,
  MAX_PHONE,
  birthDateLimit,
  onlyDigits,
} from "@/lib/validation";
import {
  Container, Card, Heading, VStack, HStack, Field, Input,
  Button, Text, Select, Spinner, createListCollection,
} from "@chakra-ui/react";
import Brand from "@/components/Brand/Brand";

interface Country {
  country_id: number;
  country_name: string;
  iso: string;
}

interface Position {
  id: number;
  department_id: number;
  department_name: string;
  title: string;
  level: string;
}

export default function OnboardingPage() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [checking, setChecking] = useState(true);
  const [error, setError] = useState("");
  const [submitted, setSubmitted] = useState(false);
  const [touched, setTouched] = useState<Record<string, boolean>>({});

  const [countries, setCountries] = useState<Country[]>([]);
  const [positions, setPositions] = useState<Position[]>([]);

  const countryCollection = createListCollection({
    items: countries,
    itemToString: (item) => item.country_name,
    itemToValue: (item) => String(item.country_id),
  });

  const positionCollection = createListCollection({
    items: positions,
    itemToString: (item) => `${item.title} (${item.level}) — ${item.department_name}`,
    itemToValue: (item) => String(item.id),
  });

  const [form, setForm] = useState({
    first_name: "",
    last_name: "",
    phone_number: "",
    private_email: "",
    street: "",
    country: 0,
    city: "",
    date_of_birth: "",
    hire_date: "",
    position_id: 0,
  });

  useEffect(() => {
    Promise.all([
      api("/api/v1/countries").then(setCountries),
      api("/api/v1/positions").then(setPositions),
    ]).catch(() => {});
  }, []);

  // Check if profile already exists
  useEffect(() => {
    api("/api/v1/employees/me")
      .then(() => navigate("/home", { replace: true }))
      .catch(() => setChecking(false));
  }, [navigate]);

  const update = (field: string, value: any) =>
    setForm((prev) => ({ ...prev, [field]: value }));

  const errors = validateProfile(form);
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
      await api("/api/v1/employees", {
        method: "POST",
        body: JSON.stringify({
          ...form,
          position_id: form.position_id || undefined,
          phone_number: form.phone_number || undefined,
          private_email: form.private_email || undefined,
          street: form.street || undefined,
          city: form.city || undefined,
          date_of_birth: form.date_of_birth || undefined,
          hire_date: form.hire_date || undefined,
        }),
      });
      navigate("/home", { replace: true });
    } catch (err: any) {
      setError(err.message || "Čuvanje profila nije uspelo.");
    } finally {
      setLoading(false);
    }
  };

  if (checking)
    return (
      <Container py={10} textAlign="center">
        <Spinner size="xl" />
      </Container>
    );

  return (
    <Container maxW="md" py={10}>
      <Card.Root borderColor="brand.200" boxShadow="md">
        <Card.Header>
          <VStack gap={1} align="center">
            <Brand fontSize="2xl" />
            <Heading size="md" color="brand.800">
              Popuni profil
            </Heading>
            <Text fontSize="sm" color="fg.muted" textAlign="center">
              Još samo osnovni podaci pa da počnemo.
            </Text>
          </VStack>
        </Card.Header>
        <Card.Body>
          <form onSubmit={handleSubmit} noValidate>
            <VStack gap={4} align="stretch">
              <HStack gap={4} width="full" align="flex-start">
                <Field.Root required invalid={!!shown("first_name")}>
                  <Field.Label>Ime</Field.Label>
                  <Input
                    maxLength={MAX_NAME}
                    value={form.first_name}
                    onChange={(e) => update("first_name", e.target.value)}
                    onBlur={() => markTouched("first_name")}
                  />
                  <Field.ErrorText>{shown("first_name")}</Field.ErrorText>
                </Field.Root>
                <Field.Root required invalid={!!shown("last_name")}>
                  <Field.Label>Prezime</Field.Label>
                  <Input
                    maxLength={MAX_NAME}
                    value={form.last_name}
                    onChange={(e) => update("last_name", e.target.value)}
                    onBlur={() => markTouched("last_name")}
                  />
                  <Field.ErrorText>{shown("last_name")}</Field.ErrorText>
                </Field.Root>
              </HStack>

              <HStack gap={4} width="full" align="flex-start">
                <Field.Root invalid={!!shown("phone_number")}>
                  <Field.Label>Telefon</Field.Label>
                  <Input
                    inputMode="numeric"
                    value={form.phone_number}
                    onChange={(e) =>
                      update(
                        "phone_number",
                        onlyDigits(e.target.value).slice(0, MAX_PHONE),
                      )
                    }
                    onBlur={() => markTouched("phone_number")}
                    placeholder="npr. 0641234567"
                  />
                  <Field.ErrorText>{shown("phone_number")}</Field.ErrorText>
                  {!shown("phone_number") && (
                    <Field.HelperText>
                      Samo cifre, najviše {MAX_PHONE}.
                    </Field.HelperText>
                  )}
                </Field.Root>
                <Field.Root invalid={!!shown("private_email")}>
                  <Field.Label>Privatni email</Field.Label>
                  <Input
                    type="email"
                    value={form.private_email}
                    onChange={(e) => update("private_email", e.target.value)}
                    onBlur={() => markTouched("private_email")}
                  />
                  <Field.ErrorText>{shown("private_email")}</Field.ErrorText>
                </Field.Root>
              </HStack>

              <Field.Root required invalid={!!shown("country")}>
                <Field.Label>Država</Field.Label>
                <Select.Root
                  collection={countryCollection}
                  value={form.country ? [String(form.country)] : []}
                  onValueChange={(e: any) => {
                    update("country", parseInt(e.value[0]) || 0);
                    markTouched("country");
                  }}
                >
                  <Select.Trigger aria-invalid={!!shown("country")}>
                    <Select.ValueText placeholder="Izaberi državu" />
                  </Select.Trigger>
                  <Select.Content>
                    {countryCollection.items.map((c) => (
                      <Select.Item key={c.country_id} item={c}>
                        {c.country_name}
                      </Select.Item>
                    ))}
                  </Select.Content>
                </Select.Root>
                <Field.ErrorText>{shown("country")}</Field.ErrorText>
              </Field.Root>

              <HStack gap={4} width="full" align="flex-start">
                <Field.Root>
                  <Field.Label>Grad</Field.Label>
                  <Input
                    value={form.city}
                    onChange={(e) => update("city", e.target.value)}
                  />
                </Field.Root>
                <Field.Root>
                  <Field.Label>Adresa</Field.Label>
                  <Input
                    value={form.street}
                    onChange={(e) => update("street", e.target.value)}
                  />
                </Field.Root>
              </HStack>

              <HStack gap={4} width="full" align="flex-start">
                <Field.Root required invalid={!!shown("date_of_birth")}>
                  <Field.Label>Datum rođenja</Field.Label>
                  <Input
                    type="date"
                    max={birthDateLimit()}
                    value={form.date_of_birth}
                    onChange={(e) => update("date_of_birth", e.target.value)}
                    onBlur={() => markTouched("date_of_birth")}
                  />
                  <Field.ErrorText>{shown("date_of_birth")}</Field.ErrorText>
                </Field.Root>
                <Field.Root required invalid={!!shown("hire_date")}>
                  <Field.Label>Datum zaposlenja</Field.Label>
                  <Input
                    type="date"
                    value={form.hire_date}
                    onChange={(e) => update("hire_date", e.target.value)}
                    onBlur={() => markTouched("hire_date")}
                  />
                  <Field.ErrorText>{shown("hire_date")}</Field.ErrorText>
                </Field.Root>
              </HStack>

              <Field.Root required invalid={!!shown("position_id")}>
                <Field.Label>Pozicija</Field.Label>
                <Select.Root
                  collection={positionCollection}
                  value={form.position_id ? [String(form.position_id)] : []}
                  onValueChange={(e: any) => {
                    update("position_id", parseInt(e.value[0]) || 0);
                    markTouched("position_id");
                  }}
                >
                  <Select.Trigger aria-invalid={!!shown("position_id")}>
                    <Select.ValueText placeholder="Izaberi poziciju" />
                  </Select.Trigger>
                  <Select.Content>
                    {positionCollection.items.map((p) => (
                      <Select.Item key={p.id} item={p}>
                        {p.title} ({p.level}) — {p.department_name}
                      </Select.Item>
                    ))}
                  </Select.Content>
                </Select.Root>
                <Field.ErrorText>{shown("position_id")}</Field.ErrorText>
              </Field.Root>

              {submitted && hasErrors(errors) && (
                <Text color="red.500" fontSize="sm">
                  {PROFILE_INCOMPLETE}
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
                Sačuvaj profil
              </Button>
            </VStack>
          </form>
        </Card.Body>
      </Card.Root>
    </Container>
  );
}
