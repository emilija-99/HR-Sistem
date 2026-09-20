import { useState, useEffect } from "react";
import { api } from "@/api/client";
import { workingDays, isWeekend, lastDateWithin, todayISO, daysLabel } from "@/lib/dates";
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

interface BalanceRow {
  absence_type_id: number;
  available_days: number;
}

interface MyPolicyRow {
  absence_type_id: number;
  policy: { requires_balance: boolean } | null;
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
  const [balances, setBalances] = useState<Record<number, number>>({});
  const [requiresBalance, setRequiresBalance] = useState<
    Record<number, boolean>
  >({});
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

  const fetchBalances = () =>
    Promise.all([
      api("/api/v1/absences/balance/me").then((rows: BalanceRow[] = []) => {
        const map: Record<number, number> = {};
        rows.forEach((r) => {
          map[r.absence_type_id] = Number(r.available_days) || 0;
        });
        setBalances(map);
      }),
      api("/api/v1/absences/balance/my-policies").then(
        (rows: MyPolicyRow[] = []) => {
          const map: Record<number, boolean> = {};
          rows.forEach((r) => {
            if (r.policy) map[r.absence_type_id] = r.policy.requires_balance;
          });
          setRequiresBalance(map);
        },
      ),
    ]);

  useEffect(() => {
    Promise.all([
      api("/api/v1/absences/types").then((d) => setTypes(d || [])),
      fetchRequests(),
      fetchBalances(),
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
      await Promise.all([fetchRequests(), fetchBalances()]);
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
      await Promise.all([fetchRequests(), fetchBalances()]);
    } catch (err: any) {
      setError(err.message || "Slanje zahteva nije uspelo");
    }
  };

  const handleCancel = async (id: number) => {
    setError("");
    try {
      await api(`/api/v1/absences/requests/${id}/cancel`, { method: "PUT" });
      if (editId === id) resetForm();
      await Promise.all([fetchRequests(), fetchBalances()]);
    } catch (err: any) {
      setError(err.message || "Otkazivanje nije uspelo");
    }
  };

  // ── validacija perioda i balansa ────────────────────────────
  const today = todayISO();
  const selected = form.absence_type_id;
  // Fail-closed, kao na serveru: tip bez politike takođe zahteva pokriće.
  const needsBalance = selected ? requiresBalance[selected] ?? true : true;
  const available = selected ? balances[selected] ?? 0 : 0;
  const requested = workingDays(form.start_date, form.end_date);

  const dateError = (() => {
    const { start_date: start, end_date: end } = form;
    if (!start || !end) return "";
    if (end < start) return "Krajnji datum ne može biti pre početnog.";
    if (start < today) return "Datum ne može biti u prošlosti.";
    if (isWeekend(start)) return "Početni datum ne može biti subota ili nedelja.";
    if (isWeekend(end)) return "Krajnji datum ne može biti subota ili nedelja.";
    if (requested === 0) return "Izabrani period ne sadrži nijedan radni dan.";
    return "";
  })();

  const balanceError =
    needsBalance && !dateError && requested > available
      ? `Za izabrani period treba ${requested} ${daysLabel(requested)}, a na raspolaganju je ${available}.`
      : "";

  const canSaveDates =
    !!selected && !!form.start_date && !!form.end_date && !dateError;
  const canSubmit = canSaveDates && !balanceError;

  // „Do" ne može preko raspoloživih dana.
  const maxEnd =
    needsBalance && form.start_date
      ? lastDateWithin(form.start_date, Math.max(1, Math.floor(available)))
      : undefined;

  const validationMessage = dateError || balanceError || error;

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
                        {balances[t.id] !== undefined
                          ? ` · ${balances[t.id]} dana`
                          : ""}
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
                    min={today}
                    value={form.start_date}
                    onChange={(e) => update("start_date", e.target.value)}
                  />
                </Field.Root>
                <Field.Root required width="full">
                  <Field.Label>Do</Field.Label>
                  <Input
                    type="date"
                    min={form.start_date || today}
                    max={maxEnd}
                    value={form.end_date}
                    onChange={(e) => update("end_date", e.target.value)}
                  />
                </Field.Root>
              </HStack>
              <Text fontSize="xs" color="fg.muted">
                Jedan dan se bira tako što su „Od“ i „Do“ isti datum. Računaju se
                samo radni dani — vikendi i praznici se ne obračunavaju. Datum ne
                može biti u prošlosti ni vikend.
              </Text>
              {selected > 0 && (
                <Text fontSize="sm" color={balanceError ? "red.500" : "fg.muted"}>
                  {needsBalance ? (
                    <>
                      Na raspolaganju: <b>{available}</b> dana
                      {form.start_date && form.end_date
                        ? ` · za izabrani period: ${requested} ${daysLabel(requested)}`
                        : ""}
                    </>
                  ) : (
                    "Ovaj tip odsustva se ne ograničava balansom (neograničeno)."
                  )}
                </Text>
              )}
              <Field.Root>
                <Field.Label>Razlog</Field.Label>
                <Textarea
                  value={form.reason}
                  onChange={(e) => update("reason", e.target.value)}
                  placeholder="Opciono"
                />
              </Field.Root>
              {validationMessage && (
                <Text color="red.500" fontSize="sm">
                  {validationMessage}
                </Text>
              )}
              <HStack>
                <Button
                  type="submit"
                  colorPalette="brand"
                  loading={submitting}
                  disabled={!canSubmit}
                >
                  {editId ? "Sačuvaj izmene" : "Podnesi zahtev"}
                </Button>
                {!editId && (
                  <Button
                    type="button"
                    variant="outline"
                    loading={submitting}
                    disabled={!canSaveDates}
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
        <Text color="fg.muted">Nema zahteva.</Text>
      ) : (
        <VStack gap={3} align="stretch">
          {requests.map((req) => (
            <Card.Root key={req.id}>
              <Card.Body>
                <HStack justify="space-between" wrap="wrap">
                  <VStack align="start" gap={1}>
                    <HStack>
                      <Badge colorPalette="brand">{req.type_name}</Badge>
                      <Badge colorPalette={statusColors[req.status] || "gray"}>
                        {req.status}
                      </Badge>
                      {req.is_paid ? (
                        <Badge colorPalette="green">plaćeno</Badge>
                      ) : (
                        <Badge colorPalette="orange">neplaćeno</Badge>
                      )}
                    </HStack>
                    <Text fontSize="sm" color="fg.muted">
                      {req.start_date} → {req.end_date} ({req.total_days} dana)
                    </Text>
                    {req.reason && (
                      <Text fontSize="sm" color="fg.muted">
                        {req.reason}
                      </Text>
                    )}
                  </VStack>
                  {(req.status === "DRAFT" || req.status === "PENDING") && (
                    <HStack>
                      {req.status === "DRAFT" && (
                        <Button
                          size="sm"
                          colorPalette="brand"
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
