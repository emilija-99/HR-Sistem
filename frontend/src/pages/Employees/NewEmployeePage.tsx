import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import Menu from "@/components/Menu/Menu";
import { api } from "@/api/client";
import {
  Container, Heading, Card, VStack, SimpleGrid, Field, Input,
  Button, Text, Spinner, Select, createListCollection,
} from "@chakra-ui/react";

export default function NewEmployeePage() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    setError("");
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

  const canSave =
    !!form.email &&
    form.password.length >= 8 &&
    !!form.first_name &&
    !!form.last_name &&
    !!form.country;

  if (loading)
    return (
      <>
        <Menu />
        <Container py={10} textAlign="center">
          <Spinner size="xl" />
        </Container>
      </>
    );

  return (
    <>
      <Menu />
      <Container maxW="container.md" py={6}>
        <Button variant="ghost" mb={4} onClick={() => navigate("/employees")}>
          ← Nazad na listu
        </Button>
        <Card.Root>
          <Card.Header>
            <Heading size="lg">Novi zaposleni</Heading>
            <Text fontSize="sm" color="fg.muted" mt={1}>
              Kreira nalog i profil zaposlenog.
            </Text>
          </Card.Header>
          <Card.Body>
            <form onSubmit={handleSubmit}>
              <VStack gap={4} align="stretch">
                <SimpleGrid columns={{ base: 1, md: 2 }} gap={4}>
                  <Field.Root required>
                    <Field.Label>Email (nalog)</Field.Label>
                    <Input
                      type="email"
                      value={form.email}
                      onChange={(e) => update("email", e.target.value)}
                    />
                  </Field.Root>
                  <Field.Root required>
                    <Field.Label>Lozinka</Field.Label>
                    <Input
                      type="password"
                      value={form.password}
                      onChange={(e) => update("password", e.target.value)}
                      placeholder="8–25 znakova"
                    />
                  </Field.Root>
                  <Field.Root required>
                    <Field.Label>Ime</Field.Label>
                    <Input
                      value={form.first_name}
                      onChange={(e) => update("first_name", e.target.value)}
                    />
                  </Field.Root>
                  <Field.Root required>
                    <Field.Label>Prezime</Field.Label>
                    <Input
                      value={form.last_name}
                      onChange={(e) => update("last_name", e.target.value)}
                    />
                  </Field.Root>
                  <Field.Root required>
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
                  <Field.Root>
                    <Field.Label>Grad</Field.Label>
                    <Input
                      value={form.city}
                      onChange={(e) => update("city", e.target.value)}
                    />
                  </Field.Root>
                  <Field.Root>
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
                  <Field.Root>
                    <Field.Label>Telefon</Field.Label>
                    <Input
                      value={form.phone_number}
                      onChange={(e) => update("phone_number", e.target.value)}
                    />
                  </Field.Root>
                  <Field.Root>
                    <Field.Label>Privatni email</Field.Label>
                    <Input
                      value={form.private_email}
                      onChange={(e) => update("private_email", e.target.value)}
                    />
                  </Field.Root>
                  <Field.Root>
                    <Field.Label>Adresa</Field.Label>
                    <Input
                      value={form.street}
                      onChange={(e) => update("street", e.target.value)}
                    />
                  </Field.Root>
                  <Field.Root>
                    <Field.Label>Datum rođenja</Field.Label>
                    <Input
                      type="date"
                      value={form.date_of_birth}
                      onChange={(e) => update("date_of_birth", e.target.value)}
                    />
                  </Field.Root>
                  <Field.Root>
                    <Field.Label>Datum zaposlenja</Field.Label>
                    <Input
                      type="date"
                      value={form.hire_date}
                      onChange={(e) => update("hire_date", e.target.value)}
                    />
                  </Field.Root>
                </SimpleGrid>

                {error && (
                  <Text color="red.500" fontSize="sm">
                    {error}
                  </Text>
                )}

                <Button
                  type="submit"
                  colorPalette="brand"
                  loading={saving}
                  disabled={!canSave}
                >
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
