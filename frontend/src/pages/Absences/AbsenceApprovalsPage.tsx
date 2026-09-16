import { useState, useEffect } from "react";
import { api } from "@/api/client";
import {
  Container, Heading, Card, VStack, HStack, Button, Text,
  Badge, Spinner, Table, Box,
} from "@chakra-ui/react";

interface AbsenceRequest {
  id: number;
  employee_id: number;
  first_name?: string;
  last_name?: string;
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

export default function AbsenceApprovalsPage() {
  const [requests, setRequests] = useState<AbsenceRequest[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const fetchRequests = () =>
    api("/api/v1/absences/requests").then(setRequests);

  useEffect(() => {
    fetchRequests()
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, []);

  const decide = async (id: number, action: "approve" | "reject") => {
    try {
      await api(`/api/v1/absences/requests/${id}/${action}`, { method: "PUT" });
      await fetchRequests();
    } catch (err: any) {
      setError(err.message || "Failed to update request");
    }
  };

  if (loading)
    return (
      <Container py={10} textAlign="center">
        <Spinner size="xl" />
      </Container>
    );

  const pending = requests.filter((r) => r.status === "PENDING");
  const decided = requests.filter((r) => r.status !== "PENDING");

  return (
    <Container maxW="container.xl" py={6}>
      <Heading mb={6}>Odobravanje odsustava</Heading>

      {error && (
        <Text color="red.500" fontSize="sm" mb={4}>
          {error}
        </Text>
      )}

      <Heading size="md" mb={4}>
        Na čekanju ({pending.length})
      </Heading>
      {pending.length === 0 ? (
        <Text color="gray.500" mb={6}>Nema zahteva na čekanju.</Text>
      ) : (
        <VStack gap={3} align="stretch" mb={8}>
          {pending.map((req) => (
            <Card.Root key={req.id}>
              <Card.Body>
                <HStack justify="space-between" wrap="wrap">
                  <VStack align="start" gap={1}>
                    <HStack>
                      <Text fontWeight="bold">
                        {req.first_name} {req.last_name}
                      </Text>
                      <Badge colorPalette="purple">{req.type_name}</Badge>
                      <Badge colorPalette="yellow">{req.status}</Badge>
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
                  <HStack>
                    <Button
                      size="sm"
                      colorPalette="green"
                      onClick={() => decide(req.id, "approve")}
                    >
                      ✓ Odobri
                    </Button>
                    <Button
                      size="sm"
                      colorPalette="red"
                      variant="outline"
                      onClick={() => decide(req.id, "reject")}
                    >
                      ✕ Odbij
                    </Button>
                  </HStack>
                </HStack>
              </Card.Body>
            </Card.Root>
          ))}
        </VStack>
      )}

      <Heading size="md" mb={4}>
        Rešeni zahtevi ({decided.length})
      </Heading>
      {decided.length === 0 ? (
        <Text color="gray.500">Nema rešenih zahteva.</Text>
      ) : (
        <Box overflowX="auto">
          <Table.Root variant="outline">
            <Table.Header>
              <Table.Row>
                <Table.ColumnHeader>Zaposleni</Table.ColumnHeader>
                <Table.ColumnHeader>Tip</Table.ColumnHeader>
                <Table.ColumnHeader>Period</Table.ColumnHeader>
                <Table.ColumnHeader>Dana</Table.ColumnHeader>
                <Table.ColumnHeader>Status</Table.ColumnHeader>
              </Table.Row>
            </Table.Header>
            <Table.Body>
              {decided.map((req) => (
                <Table.Row key={req.id}>
                  <Table.Cell>
                    {req.first_name} {req.last_name}
                  </Table.Cell>
                  <Table.Cell>{req.type_name}</Table.Cell>
                  <Table.Cell>
                    {req.start_date} → {req.end_date}
                  </Table.Cell>
                  <Table.Cell>{req.total_days}</Table.Cell>
                  <Table.Cell>
                    <Badge colorPalette={statusColors[req.status] || "gray"}>
                      {req.status}
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
