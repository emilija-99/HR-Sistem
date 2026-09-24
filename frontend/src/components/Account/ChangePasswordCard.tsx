import { useState } from "react";
import { api } from "@/api/client";
import {
  MAX_PASSWORD,
  MIN_PASSWORD,
  REQUIRED,
  isBlank,
} from "@/lib/validation";
import { Card, Heading, VStack, Field, Input, Button, Text } from "@chakra-ui/react";

export default function ChangePasswordCard() {
  const [current, setCurrent] = useState("");
  const [next, setNext] = useState("");
  const [confirm, setConfirm] = useState("");
  const [saving, setSaving] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [formError, setFormError] = useState("");
  const [success, setSuccess] = useState("");

  const clearError = (field: string) => {
    setSuccess("");
    setErrors((prev) => {
      if (!prev[field]) return prev;
      const copy = { ...prev };
      delete copy[field];
      return copy;
    });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError("");
    setSuccess("");
    setErrors({});

    // ── klijentska validacija (ista pravila kao na serveru: 8–25 znakova) ──
    const found: Record<string, string> = {};
    if (isBlank(current)) found.current = REQUIRED;
    if (isBlank(next)) found.next = REQUIRED;
    else if (next.length < MIN_PASSWORD || next.length > MAX_PASSWORD)
      found.next = `Lozinka mora imati između ${MIN_PASSWORD} i ${MAX_PASSWORD} znakova.`;
    if (isBlank(confirm)) found.confirm = REQUIRED;
    // Ne znamo koje je polje pogrešno → obeležavamo oba.
    else if (next !== confirm) {
      found.next = "Lozinke se ne poklapaju.";
      found.confirm = "Lozinke se ne poklapaju.";
    }

    if (Object.keys(found).length > 0) {
      setErrors(found);
      return;
    }

    setSaving(true);
    try {
      await api("/api/v1/change-password", {
        method: "PUT",
        body: JSON.stringify({
          current_password: current,
          new_password: next,
        }),
      });
      setCurrent("");
      setNext("");
      setConfirm("");
      setSuccess("Lozinka je uspešno promenjena.");
    } catch (err: any) {
      // 400 = trenutna lozinka nije ispravna → greška ide na to polje.
      if (err?.status === 400) {
        setErrors({ current: err.message });
      } else {
        setFormError(err?.message || "Promena lozinke nije uspela.");
      }
    } finally {
      setSaving(false);
    }
  };

  // Greške postoje samo posle neuspele akcije (klijentska validacija ili server),
  // pa se prikazuju direktno — nema potrebe za „touched" stanjem.
  const shown = (field: string) => errors[field] || "";

  return (
    <Card.Root mt={6} borderColor="brand.200" boxShadow="md">
      <Card.Header>
        <Heading size="md" color="brand.800">
          Promena lozinke
        </Heading>
      </Card.Header>
      <Card.Body>
        <form onSubmit={handleSubmit} noValidate>
          <VStack gap={4} align="stretch">
            <Field.Root required invalid={!!shown("current")}>
              <Field.Label>Trenutna lozinka</Field.Label>
              <Input
                type="password"
                maxLength={MAX_PASSWORD}
                value={current}
                onChange={(e) => {
                  setCurrent(e.target.value);
                  clearError("current");
                }}
                autoComplete="current-password"
              />
              <Field.ErrorText>{shown("current")}</Field.ErrorText>
            </Field.Root>

            <Field.Root required invalid={!!shown("next")}>
              <Field.Label>Nova lozinka</Field.Label>
              <Input
                type="password"
                maxLength={MAX_PASSWORD}
                value={next}
                onChange={(e) => {
                  setNext(e.target.value);
                  clearError("next");
                  clearError("confirm");
                }}
                autoComplete="new-password"
                placeholder={`${MIN_PASSWORD}–${MAX_PASSWORD} znakova`}
              />
              <Field.ErrorText>{shown("next")}</Field.ErrorText>
            </Field.Root>

            <Field.Root required invalid={!!shown("confirm")}>
              <Field.Label>Potvrdi novu lozinku</Field.Label>
              <Input
                type="password"
                maxLength={MAX_PASSWORD}
                value={confirm}
                onChange={(e) => {
                  setConfirm(e.target.value);
                  clearError("next");
                  clearError("confirm");
                }}
                autoComplete="new-password"
              />
              <Field.ErrorText>{shown("confirm")}</Field.ErrorText>
            </Field.Root>

            {formError && (
              <Text color="red.500" fontSize="sm">
                {formError}
              </Text>
            )}
            {success && (
              <Text color="green.500" fontSize="sm">
                {success}
              </Text>
            )}

            <Button type="submit" colorPalette="brand" loading={saving}>
              Promeni lozinku
            </Button>
          </VStack>
        </form>
      </Card.Body>
    </Card.Root>
  );
}
