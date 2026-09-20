import { useState, useEffect } from "react";
import { api } from "@/api/client";
import {
  Container, Heading, Card, VStack, HStack, Button, Text,
  Badge, Spinner, Table, Box,
} from "@chakra-ui/react";

interface AttendanceRecord {
  id: number;
  employee_id: number;
  clock_in: string;
  clock_out: string | null;
  status: string;
  worked_minutes: number | null;
  first_name?: string;
  last_name?: string;
}

function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString("sr-RS", {
    dateStyle: "short",
    timeStyle: "medium",
  });
}

function formatDuration(minutes: number): string {
  if (minutes < 1) return "< 1 min";
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  return h > 0 ? `${h}h ${m}m` : `${m} min`;
}

export default function AttendancePage() {
  const [status, setStatus] = useState<{ is_working: boolean; record?: AttendanceRecord } | null>(null);
  const [records, setRecords] = useState<AttendanceRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  const refresh = async () => {
    const [st, recs] = await Promise.all([
      api("/api/v1/attendance/status"),
      api("/api/v1/attendance/me"),
    ]);
    setStatus(st);
    setRecords(recs);
  };

  useEffect(() => {
    refresh()
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, []);

  const clockIn = async () => {
    setBusy(true);
    setError("");
    try {
      await api("/api/v1/attendance/clock-in", { method: "POST" });
      await refresh();
    } catch (err: any) {
      setError(err.message || "Clock-in failed");
    } finally {
      setBusy(false);
    }
  };

  const clockOut = async () => {
    setBusy(true);
    setError("");
    try {
      await api("/api/v1/attendance/clock-out", { method: "POST" });
      await refresh();
    } catch (err: any) {
      setError(err.message || "Clock-out failed");
    } finally {
      setBusy(false);
    }
  };

  if (loading)
    return (
      <Container py={10} textAlign="center">
        <Spinner size="xl" />
      </Container>
    );

  const isWorking = status?.is_working ?? false;

  return (
    <Container maxW="container.lg" py={6}>
      <Heading mb={6}>Prisustvo na poslu</Heading>

      {error && (
        <Text color="red.500" fontSize="sm" mb={4}>
          {error}
        </Text>
      )}

      <Card.Root mb={8}>
        <Card.Body>
          <HStack justify="space-between" wrap="wrap" gap={4}>
            <VStack align="start" gap={2}>
              <HStack>
                <Text fontWeight="bold">Status:</Text>
                {isWorking ? (
                  <Badge colorPalette="green" fontSize="md">
                    ● Na poslu
                  </Badge>
                ) : (
                  <Badge colorPalette="gray" fontSize="md">
                    ○ Van posla
                  </Badge>
                )}
              </HStack>
              {isWorking && status?.record && (
                <Text fontSize="sm" color="fg.muted">
                  Prijava: {formatDateTime(status.record.clock_in)}
                </Text>
              )}
            </VStack>
            {isWorking ? (
              <Button colorPalette="red" size="lg" loading={busy} onClick={clockOut}>
                ⏱ Odjavi se (clock out)
              </Button>
            ) : (
              <Button colorPalette="green" size="lg" loading={busy} onClick={clockIn}>
                ▶ Prijavi se (clock in)
              </Button>
            )}
          </HStack>
        </Card.Body>
      </Card.Root>

      <Heading size="md" mb={4}>
        Istorija prisustva
      </Heading>
      {records.length === 0 ? (
        <Text color="fg.muted">Nema evidentiranog prisustva.</Text>
      ) : (
        <Box overflowX="auto">
          <Table.Root variant="outline">
            <Table.Header>
              <Table.Row>
                <Table.ColumnHeader>Prijava</Table.ColumnHeader>
                <Table.ColumnHeader>Odjava</Table.ColumnHeader>
                <Table.ColumnHeader>Radni vek</Table.ColumnHeader>
                <Table.ColumnHeader>Status</Table.ColumnHeader>
              </Table.Row>
            </Table.Header>
            <Table.Body>
              {records.map((rec) => (
                <Table.Row key={rec.id}>
                  <Table.Cell>{formatDateTime(rec.clock_in)}</Table.Cell>
                  <Table.Cell>
                    {rec.clock_out ? formatDateTime(rec.clock_out) : "-"}
                  </Table.Cell>
                  <Table.Cell>
                    {rec.worked_minutes != null
                      ? formatDuration(rec.worked_minutes)
                      : "-"}
                  </Table.Cell>
                  <Table.Cell>
                    <Badge colorPalette={rec.status === "DONE" ? "gray" : "green"}>
                      {rec.status}
                    </Badge>
                  </Table.Cell>
                </Table.Row>
              ))}
            </Table.Body>
          </Table.Root>
        </Box>
      )}
    </Container>
  );
}
