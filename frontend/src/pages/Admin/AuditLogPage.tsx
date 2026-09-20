import { useState, useEffect } from "react";
import Menu from "@/components/Menu/Menu";
import { api } from "@/api/client";
import {
  Container,
  Heading,
  HStack,
  Text,
  Badge,
  Spinner,
  Table,
  Box,
  Select,
  createListCollection,
} from "@chakra-ui/react";

interface AuditEntry {
  id: string;
  timestamp: string;
  action: string;
  entity: string;
  entity_id: number;
  actor_id: number | null;
  details: Record<string, unknown> | null;
  ip: string;
  user_agent: string;
}

const ENTITIES = [
  { label: "Svi entiteti", value: "" },
  { label: "user", value: "user" },
  { label: "employee", value: "employee" },
  { label: "absence_request", value: "absence_request" },
  { label: "leave_balance", value: "leave_balance" },
  { label: "employee_leave_policy", value: "employee_leave_policy" },
];

const entityCollection = createListCollection({
  items: ENTITIES,
  itemToString: (i) => i.label,
  itemToValue: (i) => i.value,
});

const entityColors: Record<string, string> = {
  user: "blue",
  employee: "purple",
  absence_request: "orange",
  leave_balance: "green",
  employee_leave_policy: "teal",
};

export default function AuditLogPage() {
  const [entries, setEntries] = useState<AuditEntry[]>([]);
  const [actions, setActions] = useState<string[]>([]);
  const [entity, setEntity] = useState("");
  const [action, setAction] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const actionCollection = createListCollection({
    items: [
      { label: "Sve akcije", value: "" },
      ...actions.map((a) => ({ label: a, value: a })),
    ],
    itemToString: (i) => i.label,
    itemToValue: (i) => i.value,
  });

  useEffect(() => {
    api("/api/v1/audit/actions")
      .then((d) => setActions(Array.isArray(d) ? d : []))
      .catch(() => {
        /* filter stays empty */
      });
  }, []);

  useEffect(() => {
    setLoading(true);
    const params = new URLSearchParams({ limit: "100" });
    if (entity) params.set("entity", entity);
    if (action) params.set("action", action);

    api(`/api/v1/audit/recent?${params.toString()}`)
      .then((d) => {
        setEntries(d);
        setError("");
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, [entity, action]);

  return (
    <>
      <Menu />
      <Container maxW="container.xl" py={6}>
        <Heading size="lg" mb={2}>
          Audit log
        </Heading>
        <Text color="fg.muted" fontSize="sm" mb={6}>
          Evidencija akcija (ko je šta menjao i kada), poslednjih 100 zapisa.
        </Text>

        <HStack mb={6} gap={4} wrap="wrap">
          <Select.Root
            collection={entityCollection}
            size="sm"
            width="220px"
            value={[entity]}
            onValueChange={(e: any) => setEntity(e.value[0] ?? "")}
          >
            <Select.Trigger>
              <Select.ValueText placeholder="Svi entiteti" />
            </Select.Trigger>
            <Select.Content>
              {entityCollection.items.map((i) => (
                <Select.Item key={i.value} item={i}>
                  {i.label}
                </Select.Item>
              ))}
            </Select.Content>
          </Select.Root>

          <Select.Root
            collection={actionCollection}
            size="sm"
            width="260px"
            value={[action]}
            onValueChange={(e: any) => setAction(e.value[0] ?? "")}
          >
            <Select.Trigger>
              <Select.ValueText placeholder="Sve akcije" />
            </Select.Trigger>
            <Select.Content>
              {actionCollection.items.map((i) => (
                <Select.Item key={i.value} item={i}>
                  {i.label}
                </Select.Item>
              ))}
            </Select.Content>
          </Select.Root>
        </HStack>

        {error && (
          <Text color="red.500" fontSize="sm" mb={4}>
            {error}
          </Text>
        )}

        {loading ? (
          <Container py={10} textAlign="center">
            <Spinner size="xl" />
          </Container>
        ) : entries.length === 0 ? (
          <Text color="fg.muted">Nema zapisa za izabrane filtere.</Text>
        ) : (
          <Box overflowX="auto">
            <Table.Root variant="outline" size="sm">
              <Table.Header>
                <Table.Row>
                  <Table.ColumnHeader>Vreme</Table.ColumnHeader>
                  <Table.ColumnHeader>Akcija</Table.ColumnHeader>
                  <Table.ColumnHeader>Entitet</Table.ColumnHeader>
                  <Table.ColumnHeader>Aktor</Table.ColumnHeader>
                  <Table.ColumnHeader>Detalji</Table.ColumnHeader>
                </Table.Row>
              </Table.Header>
              <Table.Body>
                {entries.map((e) => (
                  <Table.Row key={e.id}>
                    <Table.Cell whiteSpace="nowrap">
                      {new Date(e.timestamp).toLocaleString()}
                    </Table.Cell>
                    <Table.Cell>
                      <Text fontWeight="medium">{e.action}</Text>
                    </Table.Cell>
                    <Table.Cell>
                      <Badge colorPalette={entityColors[e.entity] || "gray"}>
                        {e.entity} #{e.entity_id}
                      </Badge>
                    </Table.Cell>
                    <Table.Cell>{e.actor_id ?? "-"}</Table.Cell>
                    <Table.Cell maxW="360px">
                      <Text fontSize="xs" color="fg.muted" truncate>
                        {e.details ? JSON.stringify(e.details) : "-"}
                      </Text>
                    </Table.Cell>
                  </Table.Row>
                ))}
              </Table.Body>
            </Table.Root>
          </Box>
        )}
      </Container>
    </>
  );
}
