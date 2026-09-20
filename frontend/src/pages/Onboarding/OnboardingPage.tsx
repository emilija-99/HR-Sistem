import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "@/api/client";
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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError("");
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
      setError(err.message || "Failed to create profile");
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
          <form onSubmit={handleSubmit}>
            <VStack gap={4}>
              <HStack gap={4} width="full">
                <Field.Root required width="full">
                  <Field.Label>Ime</Field.Label>
                  <Input
                    value={form.first_name}
                    onChange={(e) => update("first_name", e.target.value)}
                  />
                </Field.Root>
                <Field.Root required width="full">
                  <Field.Label>Prezime</Field.Label>
                  <Input
                    value={form.last_name}
                    onChange={(e) => update("last_name", e.target.value)}
                  />
                </Field.Root>
              </HStack>
              <HStack gap={4} width="full">
                <Field.Root width="full">
                  <Field.Label>Telefon</Field.Label>
                  <Input
                    value={form.phone_number}
                    onChange={(e) => update("phone_number", e.target.value)}
                    placeholder="+381..."
                  />
                </Field.Root>
                <Field.Root width="full">
                  <Field.Label>Privatni email</Field.Label>
                  <Input
                    type="email"
                    value={form.private_email}
                    onChange={(e) => update("private_email", e.target.value)}
                  />
                </Field.Root>
              </HStack>
              <Field.Root required width="full">
                <Field.Label>Država</Field.Label>
                <Select.Root
                  collection={countryCollection}
                  value={form.country ? [String(form.country)] : []}
                  onValueChange={(e: any) =>
                    update("country", parseInt(e.value[0]) || 0)
                  }
                >
                  <Select.Trigger>
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
              </Field.Root>
              <HStack gap={4} width="full">
                <Field.Root width="full">
                  <Field.Label>Grad</Field.Label>
                  <Input
                    value={form.city}
                    onChange={(e) => update("city", e.target.value)}
                  />
                </Field.Root>
                <Field.Root width="full">
                  <Field.Label>Adresa</Field.Label>
                  <Input
                    value={form.street}
                    onChange={(e) => update("street", e.target.value)}
                  />
                </Field.Root>
              </HStack>
              <HStack gap={4} width="full">
                <Field.Root width="full">
                  <Field.Label>Datum rođenja</Field.Label>
                  <Input
                    type="date"
                    value={form.date_of_birth}
                    onChange={(e) => update("date_of_birth", e.target.value)}
                  />
                </Field.Root>
                <Field.Root width="full">
                  <Field.Label>Datum zaposlenja</Field.Label>
                  <Input
                    type="date"
                    value={form.hire_date}
                    onChange={(e) => update("hire_date", e.target.value)}
                  />
                </Field.Root>
              </HStack>
              <Field.Root width="full">
                <Field.Label>Pozicija</Field.Label>
                <Select.Root
                  collection={positionCollection}
                  value={form.position_id ? [String(form.position_id)] : []}
                  onValueChange={(e: any) =>
                    update("position_id", parseInt(e.value[0]) || 0)
                  }
                >
                  <Select.Trigger>
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
                Sačuvaj profil
              </Button>
            </VStack>
          </form>
        </Card.Body>
      </Card.Root>
    </Container>
  );
}
