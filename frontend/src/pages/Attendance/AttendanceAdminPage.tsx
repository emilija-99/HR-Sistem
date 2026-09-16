import { useState, useEffect, useMemo } from "react";
import { api } from "@/api/client";
import {
  Container, Heading, VStack, HStack, Button, Text,
  Badge, Spinner, Table, Box, Select, createListCollection,
} from "@chakra-ui/react";

interface AttendanceRecord {
  id: number;
  employee_id: number;
  clock_in: string;
  clock_out: string | null;
  status: string;
  worked_minutes: number | null;
  first_name: string;
  last_name: string;
  position_title: string;
  position_level: string;
  department_id: number | null;
  department_name: string;
}

interface DepartmentGroup {
  id: number;
  name: string;
  records: AttendanceRecord[];
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

export default function AttendanceAdminPage() {
  const [records, setRecords] = useState<AttendanceRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [statusFilter, setStatusFilter] = useState<string[]>([]); // [] = all
  const [deptFilter, setDeptFilter] = useState<string[]>([]); // [] = all

  const statusCollection = createListCollection({
    items: [
      { id: "WORKING", name: "Na poslu (WORKING)" },
      { id: "DONE", name: "Odjavljeni (DONE)" },
    ],
    itemToString: (i) => i.name,
    itemToValue: (i) => i.id,
  });

  useEffect(() => {
    api("/api/v1/attendance")
      .then(setRecords)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, []);

  const departments = useMemo<DepartmentGroup[]>(() => {
    const filtered = records.filter(
      (r) =>
        (statusFilter.length === 0 || statusFilter.includes(r.status)) &&
        (deptFilter.length === 0 ||
          r.department_id === null ||
          deptFilter.includes(String(r.department_id))),
    );

    const map = new Map<number, DepartmentGroup>();
    for (const r of filtered) {
      const id = r.department_id ?? 0;
      const name = r.department_name || "Bez departmana";
      if (!map.has(id)) map.set(id, { id, name, records: [] });
      map.get(id)!.records.push(r);
    }
    return Array.from(map.values()).sort((a, b) => a.name.localeCompare(b.name));
  }, [records, statusFilter, deptFilter]);

  const deptOptions = useMemo(
    () =>
      Array.from(
        new Map(
          records.map((r) => [
            r.department_id ?? 0,
            { id: String(r.department_id ?? 0), name: r.department_name || "Bez departmana" },
          ]),
        ).values(),
      ),
    [records],
  );

  const deptCollection = createListCollection({
    items: deptOptions,
    itemToString: (i) => i.name,
    itemToValue: (i) => i.id,
  });

  if (loading)
    return (
      <Container py={10} textAlign="center">
        <Spinner size="xl" />
      </Container>
    );

  const workingCount = records.filter((r) => r.status === "WORKING").length;

  return (
    <Container maxW="container.xl" py={6}>
      <Heading mb={2}>Prisustvo — svi zaposleni</Heading>
      <Text color="gray.600" mb={6}>
        Trenutno na poslu: <Badge colorPalette="green">{workingCount}</Badge>{" "}
        · Ukupno zapisa: {records.length}
      </Text>

      {error && (
        <Text color="red.500" fontSize="sm" mb={4}>
          {error}
        </Text>
      )}

      <HStack gap={4} mb={6} wrap="wrap">
        <Box>
          <Text fontSize="sm" fontWeight="bold" mb={1}>
            Status
          </Text>
          <Select.Root
            collection={statusCollection}
            multiple
            value={statusFilter}
            onValueChange={(e: any) => setStatusFilter(e.value)}
          >
            <Select.Trigger width={56}>
              <Select.ValueText placeholder="Svi statusi" />
            </Select.Trigger>
            <Select.Content>
              {statusCollection.items.map((s) => (
                <Select.Item key={s.id} item={s}>
                  {s.name}
                </Select.Item>
              ))}
            </Select.Content>
          </Select.Root>
        </Box>

        <Box>
          <Text fontSize="sm" fontWeight="bold" mb={1}>
            Departman
          </Text>
          <Select.Root
            collection={deptCollection}
            multiple
            value={deptFilter}
            onValueChange={(e: any) => setDeptFilter(e.value)}
          >
            <Select.Trigger width={72}>
              <Select.ValueText placeholder="Svi departmani" />
            </Select.Trigger>
            <Select.Content>
              {deptCollection.items.map((d) => (
                <Select.Item key={d.id} item={d}>
                  {d.name}
                </Select.Item>
              ))}
            </Select.Content>
          </Select.Root>
        </Box>

        <Button
          size="sm"
          variant="outline"
          mt={5}
          onClick={() => {
            setStatusFilter([]);
            setDeptFilter([]);
          }}
        >
          Očisti filtere
        </Button>
      </HStack>

      {departments.length === 0 ? (
        <Text color="gray.500">Nema zapisa za izabrane filtere.</Text>
      ) : (
        <VStack gap={6} align="stretch">
          {departments.map((dept) => (
            <Box key={dept.id}>
              <Heading size="md" mb={2}>
                {dept.name}{" "}
                <Badge colorPalette="blue" ml={2}>
                  {dept.records.length}
                </Badge>
              </Heading>
              <Box overflowX="auto">
                <Table.Root variant="outline">
                  <Table.Header>
                    <Table.Row>
                      <Table.ColumnHeader>Zaposleni</Table.ColumnHeader>
                      <Table.ColumnHeader>Pozicija</Table.ColumnHeader>
                      <Table.ColumnHeader>Prijava</Table.ColumnHeader>
                      <Table.ColumnHeader>Odjava</Table.ColumnHeader>
                      <Table.ColumnHeader>Radni vek</Table.ColumnHeader>
                      <Table.ColumnHeader>Status</Table.ColumnHeader>
                    </Table.Row>
                  </Table.Header>
                  <Table.Body>
                    {dept.records.map((rec) => (
                      <Table.Row key={rec.id}>
                        <Table.Cell>
                          <Text fontWeight="medium">
                            {rec.first_name} {rec.last_name}
                          </Text>
                        </Table.Cell>
                        <Table.Cell>
                          {rec.position_title ? (
                            <Text fontSize="sm">
                              {rec.position_title}{" "}
                              <Badge ml={1} size="xs">
                                {rec.position_level}
                              </Badge>
                            </Text>
                          ) : (
                            "-"
                          )}
                        </Table.Cell>
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
                          {rec.status === "WORKING" ? (
                            <Badge colorPalette="green">● Na poslu</Badge>
                          ) : (
                            <Badge colorPalette="gray">Odjavljen</Badge>
                          )}
                        </Table.Cell>
                      </Table.Row>
                    ))}
                  </Table.Body>
                </Table.Root>
              </Box>
            </Box>
          ))}
        </VStack>
      )}
    </Container>
  );
}
