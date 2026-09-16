import { useState } from "react";
import { api } from "@/api/client";
import { Card, Heading, VStack, Field, Input, Button, Text } from "@chakra-ui/react";

export default function ChangePasswordCard() {
  const [current, setCurrent] = useState("");
  const [next, setNext] = useState("");
  const [confirm, setConfirm] = useState("");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setSuccess("");

    if (next.length < 8 || next.length > 25) {
      setError("Nova lozinka mora imati između 8 i 25 znakova.");
      return;
    }
    if (next !== confirm) {
      setError("Nova lozinka i potvrda se ne poklapaju.");
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
      setError(err.message || "Promena lozinke nije uspela.");
    } finally {
      setSaving(false);
    }
  };

  return (
    <Card.Root mt={6}>
      <Card.Header>
        <Heading size="md">Promena lozinke</Heading>
      </Card.Header>
      <Card.Body>
        <form onSubmit={handleSubmit}>
          <VStack gap={4} align="stretch">
            <Field.Root required>
              <Field.Label>Trenutna lozinka</Field.Label>
              <Input
                type="password"
                value={current}
                onChange={(e) => setCurrent(e.target.value)}
                autoComplete="current-password"
              />
            </Field.Root>
            <Field.Root required>
              <Field.Label>Nova lozinka</Field.Label>
              <Input
                type="password"
                value={next}
                onChange={(e) => setNext(e.target.value)}
                autoComplete="new-password"
                placeholder="8–25 znakova"
              />
            </Field.Root>
            <Field.Root required>
              <Field.Label>Potvrdi novu lozinku</Field.Label>
              <Input
                type="password"
                value={confirm}
                onChange={(e) => setConfirm(e.target.value)}
                autoComplete="new-password"
              />
            </Field.Root>
            {error && (
              <Text color="red.500" fontSize="sm">
                {error}
              </Text>
            )}
            {success && (
              <Text color="green.500" fontSize="sm">
                {success}
              </Text>
            )}
            <Button
              type="submit"
              colorPalette="blue"
              loading={saving}
              disabled={!current || !next || !confirm}
            >
              Promeni lozinku
            </Button>
          </VStack>
        </form>
      </Card.Body>
    </Card.Root>
  );
}
