import { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { api } from "@/api/client";
import {
  Container,
  Card,
  Heading,
  VStack,
  HStack,
  Field,
  Input,
  Button,
  Text,
  Spinner,
  Badge,
  SimpleGrid,
  Separator,
} from "@chakra-ui/react";

export default function EmployeeDetailPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [editing, setEditing] = useState(false);
  const [error, setError] = useState("");
  const [form, setForm] = useState<any>({});

  useEffect(() => {
    api(`/api/v1/employees/${id}`)
      .then((data) => setForm(data))
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [id]);

  const update = (field: string, value: any) =>
    setForm((prev: any) => ({ ...prev, [field]: value }));

  const handleSave = async () => {
    setSaving(true);
    setError("");
    try {
      const updated = await api(`/api/v1/employees/${id}`, {
        method: "PATCH",
        body: JSON.stringify({
          first_name: form.first_name,
          last_name: form.last_name,
          phone_number: form.phone_number,
          private_email: form.private_email,
          street: form.street,
          country: form.country,
          city: form.city,
          date_of_birth: form.date_of_birth,
          hire_date: form.hire_date,
          position_id: form.position_id,
        }),
      });
      setForm(updated);
      setEditing(false);
    } catch (err: any) {
      setError(err.message || "Failed to update");
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
      <Button variant="ghost" mb={4} onClick={() => navigate("/employees")}>
        ← Nazad na listu
      </Button>

      <Card.Root>
        <Card.Header>
          <HStack justify="space-between">
            <VStack align="start" gap={1}>
              <Heading size="lg">
                {form.first_name} {form.last_name}
              </Heading>
              {form.position_title && (
                <Text color="gray.500" fontSize="md">
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
                <Badge colorPalette="purple">{form.department_name}</Badge>
              )}
            </VStack>
            <Button
              onClick={() => (editing ? handleSave() : setEditing(true))}
              colorPalette={editing ? "green" : "blue"}
              loading={saving}
            >
              {editing ? "Sačuvaj" : "Izmeni"}
            </Button>
          </HStack>
        </Card.Header>
        <Card.Body>
          <SimpleGrid columns={{ base: 1, md: 2 }} gap={4}>
            <Field.Root>
              <Field.Label>Ime</Field.Label>
              <Input
                value={form.first_name || ""}
                onChange={(e) => update("first_name", e.target.value)}
                disabled={!editing}
              />
            </Field.Root>
            <Field.Root>
              <Field.Label>Prezime</Field.Label>
              <Input
                value={form.last_name || ""}
                onChange={(e) => update("last_name", e.target.value)}
                disabled={!editing}
              />
            </Field.Root>
            <Field.Root>
              <Field.Label>Telefon</Field.Label>
              <Input
                value={form.phone_number || ""}
                onChange={(e) => update("phone_number", e.target.value)}
                disabled={!editing}
              />
            </Field.Root>
            <Field.Root>
              <Field.Label>Privatni email</Field.Label>
              <Input
                value={form.private_email || ""}
                onChange={(e) => update("private_email", e.target.value)}
                disabled={!editing}
              />
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
              <Field.Label>Država ID</Field.Label>
              <Input
                type="number"
                value={form.country ?? ""}
                onChange={(e) =>
                  update("country", parseInt(e.target.value) || 0)
                }
                disabled={!editing}
              />
            </Field.Root>
            <Field.Root>
              <Field.Label>Pozicija ID</Field.Label>
              <Input
                type="number"
                value={form.position_id ?? ""}
                onChange={(e) =>
                  update("position_id", parseInt(e.target.value) || 0)
                }
                disabled={!editing}
              />
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
          {error && (
            <Text color="red.500" fontSize="sm" mt={4}>
              {error}
            </Text>
          )}
        </Card.Body>
      </Card.Root>
    </Container>
  );
}
