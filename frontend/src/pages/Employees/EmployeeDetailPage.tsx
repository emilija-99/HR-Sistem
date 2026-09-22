import { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { api } from "@/api/client";
import {
  validateContact,
  hasErrors,
  FORM_INCOMPLETE,
  MAX_NAME,
  MAX_PHONE,
  onlyDigits,
} from "@/lib/validation";
import ChangePasswordCard from "@/components/Account/ChangePasswordCard";
import {
  Container, Card, Heading, VStack, HStack, Field, Input,
  Button, Text, Spinner, Badge, SimpleGrid, Select, createListCollection,
} from "@chakra-ui/react";

export default function EmployeeDetailPage({ me = false }: { me?: boolean }) {
  const { id } = useParams();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [editing, setEditing] = useState(false);
  const [error, setError] = useState("");
  const [submitted, setSubmitted] = useState(false);
  const [touched, setTouched] = useState<Record<string, boolean>>({});
  const [form, setForm] = useState<any>({});
  const [countries, setCountries] = useState<any[]>([]);
  const [positions, setPositions] = useState<any[]>([]);
  const [employees, setEmployees] = useState<any[]>([]);

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
    items: employees.filter((e) => e.id !== form.id),
    itemToString: (item) => `${item.first_name} ${item.last_name}`,
    itemToValue: (item) => String(item.id),
  });

  useEffect(() => {
    const empPath = me ? "/api/v1/employees/me" : `/api/v1/employees/${id}`;
    // Only HR/admin may list employees (needed for the supervisor picker).
    const extras: Promise<any>[] = [
      api("/api/v1/countries"),
      api("/api/v1/positions"),
    ];
    if (!me) extras.push(api("/api/v1/employees"));

    Promise.all([api(empPath), ...extras])
      .then(([emp, countries, positions, employees]) => {
        setForm(emp);
        setCountries(countries);
        setPositions(positions);
        if (employees) setEmployees(employees);
      })
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [id, me]);

  const update = (field: string, value: any) =>
    setForm((prev: any) => ({ ...prev, [field]: value }));

  const errors = validateContact(form);
  const markTouched = (field: string) =>
    setTouched((prev) => ({ ...prev, [field]: true }));
  // Greške se vide samo u režimu izmene, nakon klika na „Sačuvaj" ili blur-a.
  const shown = (field: string) =>
    editing && (submitted || touched[field]) ? errors[field] : "";

  const startEditing = () => {
    setSubmitted(false);
    setTouched({});
    setError("");
    setEditing(true);
  };

  const handleSave = async () => {
    setSubmitted(true);
    setError("");

    if (hasErrors(errors)) return;

    setSaving(true);
    try {
      const updated = await api(
        me ? "/api/v1/employees/me" : `/api/v1/employees/${id}`,
        {
          method: "PATCH",
          body: JSON.stringify({
            first_name: form.first_name,
            last_name: form.last_name,
            // Prazna opciona polja se izostavljaju (ne šalje se ""), jer bi
            // "" za DATE kolonu izazvalo 500 — server čuva postojeću vrednost.
            phone_number: form.phone_number || undefined,
            private_email: form.private_email || undefined,
            street: form.street || undefined,
            country: form.country,
            city: form.city || undefined,
            date_of_birth: form.date_of_birth || undefined,
            hire_date: form.hire_date || undefined,
            position_id: form.position_id || undefined,
            // the API ignores this for self-service updates
            supervisor_id: form.supervisor_id || undefined,
          }),
        },
      );
      setForm(updated);
      setEditing(false);
    } catch (err: any) {
      setError(err.message || "Izmena nije uspela.");
    } finally {
      setSaving(false);
    }
  };

  if (loading)
    return (
      <Container py={10} textAlign="center">
        <Spinner size="xl" />
      </Container>
    );

  return (
    <Container maxW="container.md" py={6}>
      <Button
        variant="ghost"
        mb={4}
        onClick={() => navigate(me ? "/home" : "/employees")}
      >
        ← {me ? "Nazad na početnu" : "Nazad na listu"}
      </Button>

      <Card.Root>
        <Card.Header>
          <HStack justify="space-between">
            <VStack align="start" gap={1}>
              <Heading size="lg">
                {form.first_name} {form.last_name}
              </Heading>
              {form.position_title && (
                <Text color="fg.muted" fontSize="md">
                  {form.position_title}
                  {form.position_level && (
                    <Badge
                      ml={2}
                      colorPalette={
                        form.position_level === "LEAD"
                          ? "orange"
                          : form.position_level === "SENIOR"
                            ? "red"
                            : form.position_level === "MEDIOR"
                              ? "blue"
                              : "gray"
                      }
                    >
                      {form.position_level}
                    </Badge>
                  )}
                </Text>
              )}
              {form.department_name && (
                <Badge colorPalette="brand">{form.department_name}</Badge>
              )}
            </VStack>
            <Button
              onClick={() => (editing ? handleSave() : startEditing())}
              colorPalette={editing ? "green" : "brand"}
              loading={saving}
            >
              {editing ? "Sačuvaj" : "Izmeni"}
            </Button>
          </HStack>
        </Card.Header>
        <Card.Body>
          <SimpleGrid columns={{ base: 1, md: 2 }} gap={4}>
            <Field.Root required invalid={!!shown("first_name")}>
              <Field.Label>Ime</Field.Label>
              <Input
                maxLength={MAX_NAME}
                value={form.first_name || ""}
                onChange={(e) => update("first_name", e.target.value)}
                onBlur={() => markTouched("first_name")}
                disabled={!editing}
              />
              <Field.ErrorText>{shown("first_name")}</Field.ErrorText>
            </Field.Root>
            <Field.Root required invalid={!!shown("last_name")}>
              <Field.Label>Prezime</Field.Label>
              <Input
                maxLength={MAX_NAME}
                value={form.last_name || ""}
                onChange={(e) => update("last_name", e.target.value)}
                onBlur={() => markTouched("last_name")}
                disabled={!editing}
              />
              <Field.ErrorText>{shown("last_name")}</Field.ErrorText>
            </Field.Root>
            <Field.Root invalid={!!shown("phone_number")}>
              <Field.Label>Telefon</Field.Label>
              <Input
                inputMode="numeric"
                value={form.phone_number || ""}
                onChange={(e) =>
                  update(
                    "phone_number",
                    onlyDigits(e.target.value).slice(0, MAX_PHONE),
                  )
                }
                onBlur={() => markTouched("phone_number")}
                disabled={!editing}
                placeholder="npr. 0641234567"
              />
              <Field.ErrorText>{shown("phone_number")}</Field.ErrorText>
            </Field.Root>
            <Field.Root invalid={!!shown("private_email")}>
              <Field.Label>Privatni email</Field.Label>
              <Input
                type="email"
                value={form.private_email || ""}
                onChange={(e) => update("private_email", e.target.value)}
                onBlur={() => markTouched("private_email")}
                disabled={!editing}
              />
              <Field.ErrorText>{shown("private_email")}</Field.ErrorText>
            </Field.Root>
            <Field.Root>
              <Field.Label>Grad</Field.Label>
              <Input
                value={form.city || ""}
                onChange={(e) => update("city", e.target.value)}
                disabled={!editing}
              />
            </Field.Root>
            <Field.Root>
              <Field.Label>Adresa</Field.Label>
              <Input
                value={form.street || ""}
                onChange={(e) => update("street", e.target.value)}
                disabled={!editing}
              />
            </Field.Root>
            <Field.Root>
              <Field.Label>Država</Field.Label>
              <Select.Root
                collection={countryCollection}
                value={form.country ? [String(form.country)] : []}
                onValueChange={(e: any) =>
                  update("country", parseInt(e.value[0]) || 0)
                }
                disabled={!editing}
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
            <Field.Root>
              <Field.Label>Pozicija</Field.Label>
              <Select.Root
                collection={positionCollection}
                value={form.position_id ? [String(form.position_id)] : []}
                onValueChange={(e: any) =>
                  update("position_id", parseInt(e.value[0]) || 0)
                }
                disabled={!editing}
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
            <Field.Root>
              <Field.Label>Nadređeni</Field.Label>
              {me ? (
                <Input value={form.supervisor_name || "-"} disabled />
              ) : (
                <Select.Root
                  collection={supervisorCollection}
                  value={
                    form.supervisor_id ? [String(form.supervisor_id)] : []
                  }
                  onValueChange={(e: any) =>
                    update("supervisor_id", parseInt(e.value[0]) || 0)
                  }
                  disabled={!editing}
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
              )}
            </Field.Root>
            <Field.Root>
              <Field.Label>Datum rođenja</Field.Label>
              <Input
                type="date"
                value={form.date_of_birth || ""}
                onChange={(e) => update("date_of_birth", e.target.value)}
                disabled={!editing}
              />
            </Field.Root>
            <Field.Root>
              <Field.Label>Datum zaposlenja</Field.Label>
              <Input
                type="date"
                value={form.hire_date || ""}
                onChange={(e) => update("hire_date", e.target.value)}
                disabled={!editing}
              />
            </Field.Root>
          </SimpleGrid>
          {submitted && hasErrors(errors) && (
            <Text color="red.500" fontSize="sm" mt={4}>
              {FORM_INCOMPLETE}
            </Text>
          )}
          {error && (
            <Text color="red.500" fontSize="sm" mt={4}>
              {error}
            </Text>
          )}
        </Card.Body>
      </Card.Root>

      {me && <ChangePasswordCard />}
    </Container>
  );
}
