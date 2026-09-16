import { useState, useEffect } from "react";
import { api } from "@/api/client";
import {
  Container, Heading, Card, VStack, HStack, Field, Input,
  Button, Text, Badge, Spinner, Select, Textarea, createListCollection,
} from "@chakra-ui/react";

interface AbsenceType {
  id: number;
  type_name: string;
  code: string;
  is_paid: boolean;
}

interface AbsenceRequest {
  id: number;
  absence_type_id: number;
  type_name: string;
  is_paid: boolean;
  start_date: string;
  end_date: string;
  total_days: number;
  reason: string | null;
  status: string;
  created_at: string;
}

const statusColors: Record<string, string> = {
  PENDING: "yellow",
  APPROVED: "green",
  REJECTED: "red",
  CANCELLED: "gray",
  DRAFT: "blue",
};

const emptyForm = {
  absence_type_id: 0,
  start_date: "",
  end_date: "",
  reason: "",
};

export default function AbsencesPage() {
  const [types, setTypes] = useState<AbsenceType[]>([]);
  const [requests, setRequests] = useState<AbsenceRequest[]>([]);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");
  const [editId, setEditId] = useState<number | null>(null);

  const typeCollection = createListCollection({
    items: types,
    itemToString: (item) => item.type_name,
    itemToValue: (item) => String(item.id),
  });

  const [form, setForm] = useState(emptyForm);

  const fetchRequests = () =>
    api("/api/v1/absences/requests/me").then(setRequests);

  useEffect(() => {
    Promise.all([
      api("/api/v1/absences/types").then((d) => setTypes(d.data || [])),
      fetchRequests(),
    ])
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, []);

  const update = (field: string, value: any) =>
    setForm((prev) => ({ ...prev, [field]: value }));

  const resetForm = () => {
    setEditId(null);
    setForm(emptyForm);
  };

  const save = async (asDraft: boolean) => {
    setSubmitting(true);
    setError("");
    try {
      if (editId) {
        await api(`/api/v1/absences/requests/${editId}`, {
          method: "PATCH",
          body: JSON.stringify({
            absence_type_id: form.absence_type_id,
            start_date: form.start_date,
            end_date: form.end_date,
            reason: form.reason || undefined,
          }),
        });
      } else {
        await api("/api/v1/absences/requests", {
          method: "POST",
          body: JSON.stringify({
            absence_type_id: form.absence_type_id,
            start_date: form.start_date,
            end_date: form.end_date,
            reason: form.reason || undefined,
            status: asDraft ? "DRAFT" : undefined,
          }),
        });
      }
      resetForm();
      await fetchRequests();
    } catch (err: any) {
      setError(err.message || "Greška pri čuvanju zahteva");
    } finally {
      setSubmitting(false);
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    save(false);
  };

  const handleEdit = (req: AbsenceRequest) => {
    setEditId(req.id);
    setError("");
    setForm({
      absence_type_id: req.absence_type_id,
      start_date: req.start_date,
      end_date: req.end_date,
      reason: req.reason || "",
    });
  };

  const handleSubmitDraft = async (id: number) => {
    setError("");
    try {
      await api(`/api/v1/absences/requests/${id}/submit`, { method: "POST" });
      await fetchRequests();
    } catch (err: any) {
      setError(err.message || "Slanje zahteva nije uspelo");
    }
  };

  const handleCancel = async (id: number) => {
    setError("");
    try {
      await api(`/api/v1/absences/requests/${id}/cancel`, { method: "PUT" });
      if (editId === id) resetForm();
      await fetchRequests();
    } catch (err: any) {
      setError(err.message || "Otkazivanje nije uspelo");
    }
  };

  const canSave = !!form.absence_type_id && !!form.start_date && !!form.end_date;

  if (loading)
    return (
      <Container py={10} textAlign="center">
        <Spinner size="xl" />
      </Container>
    );

  return (
    <Container maxW="container.lg" py={6}>
      <Heading mb={6}>Moja odsustva</Heading>

      <Card.Root mb={8}>
        <Card.Header>
          <Heading size="md">
            {editId ? `Izmena zahteva #${editId}` : "Podnesi zahtev"}
          </Heading>
        </Card.Header>
        <Card.Body>
          <form onSubmit={handleSubmit}>
            <VStack gap={4} align="stretch">
              <Field.Root required>
                <Field.Label>Tip odsustva</Field.Label>
                <Select.Root
                  collection={typeCollection}
                  value={form.absence_type_id ? [String(form.absence_type_id)] : []}
                  onValueChange={(e: any) =>
                    update("absence_type_id", parseInt(e.value[0]) || 0)
                  }
                >
                  <Select.Trigger>
                    <Select.ValueText placeholder="Izaberi tip" />
                  </Select.Trigger>
                  <Select.Content>
                    {typeCollection.items.map((t) => (
                      <Select.Item key={t.id} item={t}>
                        {t.type_name} {t.is_paid ? "(plaćeno)" : "(neplaćeno)"}
                      </Select.Item>
                    ))}
                  </Select.Content>
                </Select.Root>
              </Field.Root>
              <HStack gap={4}>
                <Field.Root required width="full">
                  <Field.Label>Od</Field.Label>
                  <Input
                    type="date"
                    value={form.start_date}
                    onChange={(e) => update("start_date", e.target.value)}
                  />
                </Field.Root>
                <Field.Root required width="full">
                  <Field.Label>Do</Field.Label>
                  <Input
                    type="date"
                    value={form.end_date}
                    onChange={(e) => update("end_date", e.target.value)}
                  />
                </Field.Root>
              </HStack>
              <Text fontSize="xs" color="gray.500">
                Računaju se samo radni dani — vikendi i praznici se ne
                obračunavaju.
              </Text>
              <Field.Root>
                <Field.Label>Razlog</Field.Label>
                <Textarea
                  value={form.reason}
                  onChange={(e) => update("reason", e.target.value)}
                  placeholder="Opciono"
                />
              </Field.Root>
              {error && (
                <Text color="red.500" fontSize="sm">
                  {error}
                </Text>
              )}
              <HStack>
                <Button
                  type="submit"
                  colorPalette="blue"
                  loading={submitting}
                  disabled={!canSave}
                >
                  {editId ? "Sačuvaj izmene" : "Podnesi zahtev"}
                </Button>
                {!editId && (
                  <Button
                    type="button"
                    variant="outline"
                    loading={submitting}
                    disabled={!canSave}
                    onClick={() => save(true)}
                  >
                    Sačuvaj kao nacrt
                  </Button>
                )}
                {editId && (
                  <Button type="button" variant="ghost" onClick={resetForm}>
                    Otkaži izmenu
                  </Button>
                )}
              </HStack>
            </VStack>
          </form>
        </Card.Body>
      </Card.Root>

      <Heading size="md" mb={4}>
        Istorija zahteva
      </Heading>
      {requests.length === 0 ? (
        <Text color="gray.500">Nema zahteva.</Text>
      ) : (
        <VStack gap={3} align="stretch">
          {requests.map((req) => (
            <Card.Root key={req.id}>
              <Card.Body>
                <HStack justify="space-between" wrap="wrap">
                  <VStack align="start" gap={1}>
                    <HStack>
                      <Badge colorPalette="purple">{req.type_name}</Badge>
                      <Badge colorPalette={statusColors[req.status] || "gray"}>
                        {req.status}
                      </Badge>
                      {req.is_paid ? (
                        <Badge colorPalette="green">plaćeno</Badge>
                      ) : (
                        <Badge colorPalette="orange">neplaćeno</Badge>
                      )}
                    </HStack>
                    <Text fontSize="sm" color="gray.600">
                      {req.start_date} → {req.end_date} ({req.total_days} dana)
                    </Text>
                    {req.reason && (
                      <Text fontSize="sm" color="gray.500">
                        {req.reason}
                      </Text>
                    )}
                  </VStack>
                  {(req.status === "DRAFT" || req.status === "PENDING") && (
                    <HStack>
                      {req.status === "DRAFT" && (
                        <Button
                          size="sm"
                          colorPalette="blue"
                          onClick={() => handleSubmitDraft(req.id)}
                        >
                          Podnesi
                        </Button>
                      )}
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => handleEdit(req)}
                      >
                        Izmeni
                      </Button>
                      <Button
                        size="sm"
                        variant="outline"
                        colorPalette="red"
                        onClick={() => handleCancel(req.id)}
                      >
                        Otkaži
                      </Button>
                    </HStack>
                  )}
                </HStack>
              </Card.Body>
            </Card.Root>
          ))}
        </VStack>
      )}
    </Container>
  );
}
