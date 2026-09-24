import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "@/api/client";
import {
  validateAccount,
  validateProfile,
  hasErrors,
  PROFILE_INCOMPLETE,
  MAX_NAME,
  MAX_PHONE,
  birthDateLimit,
  onlyDigits,
} from "@/lib/validation";
import {
  Container, Heading, Card, VStack, SimpleGrid, Field, Input,
  Button, Text, Spinner, Select, createListCollection,
} from "@chakra-ui/react";

export default function NewEmployeePage() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [submitted, setSubmitted] = useState(false);
  const [touched, setTouched] = useState<Record<string, boolean>>({});
  const [countries, setCountries] = useState<any[]>([]);
  const [positions, setPositions] = useState<any[]>([]);
  const [employees, setEmployees] = useState<any[]>([]);

  const [form, setForm] = useState<any>({
    email: "",
    password: "",
    first_name: "",
    last_name: "",
    phone_number: "",
    private_email: "",
    street: "",
    city: "",
    country: 0,
    date_of_birth: "",
    hire_date: "",
    position_id: 0,
    supervisor_id: 0,
  });

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

  const supervisorCollection = createListCollection({
    items: employees,
    itemToString: (item) => `${item.first_name} ${item.last_name}`,
    itemToValue: (item) => String(item.id),
  });

  useEffect(() => {
    Promise.all([
      api("/api/v1/countries"),
      api("/api/v1/positions"),
      api("/api/v1/employees"),
    ])
      .then(([c, p, e]) => {
        setCountries(c);
        setPositions(p);
        setEmployees(e);
      })
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  const update = (field: string, value: any) =>
    setForm((prev: any) => ({ ...prev, [field]: value }));

  const errors = {
    ...validateAccount(form.email, form.password),
    ...validateProfile(form),
  };
  const markTouched = (field: string) =>
    setTouched((prev) => ({ ...prev, [field]: true }));
  const shown = (field: string) =>
    submitted || touched[field] ? errors[field] : "";

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitted(true);
    setError("");

    if (hasErrors(errors)) return;

    setSaving(true);
    try {
      const created = await api("/api/v1/admin/employees", {
        method: "POST",
        body: JSON.stringify({
          email: form.email,
          password: form.password,
          first_name: form.first_name,
          last_name: form.last_name,
          phone_number: form.phone_number || undefined,
          private_email: form.private_email || undefined,
          street: form.street || undefined,
          city: form.city || undefined,
          country: form.country,
          date_of_birth: form.date_of_birth || undefined,
          hire_date: form.hire_date || undefined,
          position_id: form.position_id || undefined,
          supervisor_id: form.supervisor_id || undefined,
        }),
      });
      navigate(`/employees/${created.id}`);
    } catch (err: any) {
      setError(err.message || "Kreiranje zaposlenog nije uspelo");
    } finally {
      setSaving(false);
    }
  };

  if (loading)
    return (
      <>
        <Container py={10} textAlign="center">
          <Spinner size="xl" />
        </Container>
      </>
    );

  return (
    <>
      <Container maxW="container.md" py={6}>
        <Button variant="ghost" mb={4} onClick={() => navigate("/employees")}>
          ← Nazad na listu
        </Button>
        <Card.Root borderColor="brand.200" boxShadow="md">
          <Card.Header>
            <Heading size="lg" color="brand.800">
              Novi zaposleni
            </Heading>
            <Text fontSize="sm" color="fg.muted" mt={1}>
              Kreira nalog i profil zaposlenog.
            </Text>
          </Card.Header>
          <Card.Body>
            <form onSubmit={handleSubmit} noValidate>
              <VStack gap={4} align="stretch">
                <SimpleGrid columns={{ base: 1, md: 2 }} gap={4}>
                  <Field.Root required invalid={!!shown("email")}>
                    <Field.Label>Email (nalog)</Field.Label>
                    <Input
                      type="email"
                      value={form.email}
                      onChange={(e) => update("email", e.target.value)}
                      onBlur={() => markTouched("email")}
                    />
                    <Field.ErrorText>{shown("email")}</Field.ErrorText>
                  </Field.Root>
                  <Field.Root required invalid={!!shown("password")}>
                    <Field.Label>Lozinka</Field.Label>
                    <Input
                      type="password"
                      value={form.password}
                      onChange={(e) => update("password", e.target.value)}
                      onBlur={() => markTouched("password")}
                      placeholder="8–25 znakova"
                    />
                    <Field.ErrorText>{shown("password")}</Field.ErrorText>
                  </Field.Root>
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
                  <Field.Root>
                    <Field.Label>Grad</Field.Label>
                    <Input
                      value={form.city}
                      onChange={(e) => update("city", e.target.value)}
                    />
                  </Field.Root>
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
                  <Field.Root>
                    <Field.Label>Nadređeni</Field.Label>
                    <Select.Root
                      collection={supervisorCollection}
                      value={
                        form.supervisor_id ? [String(form.supervisor_id)] : []
                      }
                      onValueChange={(e: any) =>
                        update("supervisor_id", parseInt(e.value[0]) || 0)
                      }
                    >
                      <Select.Trigger>
                        <Select.ValueText placeholder="Bez nadređenog" />
                      </Select.Trigger>
                      <Select.Content>
                        {supervisorCollection.items.map((s) => (
                          <Select.Item key={s.id} item={s}>
                            {s.first_name} {s.last_name}
                          </Select.Item>
                        ))}
                      </Select.Content>
                    </Select.Root>
                  </Field.Root>
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
                  <Field.Root>
                    <Field.Label>Adresa</Field.Label>
                    <Input
                      value={form.street}
                      onChange={(e) => update("street", e.target.value)}
                    />
                  </Field.Root>
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
                </SimpleGrid>

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

                <Button type="submit" colorPalette="brand" loading={saving}>
                  Kreiraj zaposlenog
                </Button>
              </VStack>
            </form>
          </Card.Body>
        </Card.Root>
      </Container>
    </>
  );
}
