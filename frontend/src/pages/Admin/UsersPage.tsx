import { useState, useEffect } from "react";
import Menu from "@/components/Menu/Menu";
import { api } from "@/api/client";
import { useAuth } from "@/providers/AuthProvider";
import {
  Container,
  Heading,
  HStack,
  Text,
  Badge,
  Spinner,
  Button,
  Table,
  Box,
  Select,
  createListCollection,
} from "@chakra-ui/react";

const ROLES = [
  "EMPLOYEE",
  "MANAGER_PORTAL_ACCESS",
  "HR_ADMIN",
  "PLATFORM_ADMIN",
];

const roleCollection = createListCollection({
  items: ROLES.map((r) => ({ label: r, value: r })),
  itemToString: (item) => item.label,
  itemToValue: (item) => item.value,
});

interface UserRow {
  id: number;
  email: string;
  isActive: boolean;
  role: string;
  draftRole: string;
  saving: boolean;
}

export default function UsersPage() {
  const { user: currentUser } = useAuth();
  const [rows, setRows] = useState<UserRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");

  const load = async () => {
    const users: any[] = await api("/api/v1/users");
    // GET /users has no role column, so fetch each user's role individually.
    const details = await Promise.all(
      users.map((u) =>
        api(`/api/v1/users/${u.id}`).then((d) => ({ id: u.id, role: d.role })),
      ),
    );
    const roleById = new Map(details.map((d) => [d.id, d.role]));
    setRows(
      users.map((u) => ({
        id: u.id,
        email: u.email,
        isActive: u.is_active,
        role: roleById.get(u.id) || "",
        draftRole: roleById.get(u.id) || "",
        saving: false,
      })),
    );
  };

  useEffect(() => {
    load()
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, []);

  const setRow = (id: number, patch: Partial<UserRow>) =>
    setRows((prev) => prev.map((r) => (r.id === id ? { ...r, ...patch } : r)));

  const saveRole = async (row: UserRow) => {
    setError("");
    setNotice("");
    setRow(row.id, { saving: true });
    try {
      await api(`/api/v1/users/${row.id}/role`, {
        method: "PUT",
        body: JSON.stringify({ roleName: row.draftRole }),
      });
      setRow(row.id, { role: row.draftRole, saving: false });
      setNotice(`Uloga za ${row.email} je postavljena na ${row.draftRole}.`);
    } catch (err: any) {
      setRow(row.id, { saving: false });
      setError(err.message || "Promena uloge nije uspela.");
    }
  };

  const toggleStatus = async (row: UserRow) => {
    setError("");
    setNotice("");
    setRow(row.id, { saving: true });
    try {
      await api("/api/v1/change-status", {
        method: "PUT",
        body: JSON.stringify({ id: row.id, isActive: !row.isActive }),
      });
      setRow(row.id, { isActive: !row.isActive, saving: false });
      setNotice(
        `${row.email} je ${row.isActive ? "deaktiviran" : "aktiviran"}.`,
      );
    } catch (err: any) {
      setRow(row.id, { saving: false });
      setError(err.message || "Promena statusa nije uspela.");
    }
  };

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
      <Container maxW="container.xl" py={6}>
        <Heading size="lg" mb={2}>
          Korisnici i uloge
        </Heading>
        <Text color="fg.muted" fontSize="sm" mb={6}>
          Upravljanje nalozima i dodela uloga. Promena uloge važi od sledeće
          prijave korisnika.
        </Text>

        {error && (
          <Text color="red.500" fontSize="sm" mb={4}>
            {error}
          </Text>
        )}
        {notice && (
          <Text color="green.500" fontSize="sm" mb={4}>
            {notice}
          </Text>
        )}

        <Box overflowX="auto">
          <Table.Root variant="outline">
            <Table.Header>
              <Table.Row>
                <Table.ColumnHeader>Email</Table.ColumnHeader>
                <Table.ColumnHeader>Status</Table.ColumnHeader>
                <Table.ColumnHeader>Uloga</Table.ColumnHeader>
                <Table.ColumnHeader></Table.ColumnHeader>
              </Table.Row>
            </Table.Header>
            <Table.Body>
              {rows.map((row) => {
                const isSelf = currentUser?.id === row.id;
                const changed =
                  row.draftRole !== row.role && row.draftRole !== "";
                return (
                  <Table.Row key={row.id}>
                    <Table.Cell>
                      <HStack>
                        <Text fontWeight="medium">{row.email}</Text>
                        {isSelf && <Badge colorPalette="brand">vi</Badge>}
                      </HStack>
                    </Table.Cell>
                    <Table.Cell>
                      <Badge colorPalette={row.isActive ? "green" : "red"}>
                        {row.isActive ? "AKTIVAN" : "NEAKTIVAN"}
                      </Badge>
                    </Table.Cell>
                    <Table.Cell>
                      {isSelf ? (
                        <Badge colorPalette="gray">{row.role}</Badge>
                      ) : (
                        <Select.Root
                          collection={roleCollection}
                          size="sm"
                          width="220px"
                          value={row.draftRole ? [row.draftRole] : []}
                          onValueChange={(e: any) =>
                            setRow(row.id, { draftRole: e.value[0] })
                          }
                          disabled={row.saving}
                        >
                          <Select.Trigger>
                            <Select.ValueText placeholder="Izaberi ulogu" />
                          </Select.Trigger>
                          <Select.Content>
                            {roleCollection.items.map((r) => (
                              <Select.Item key={r.value} item={r}>
                                {r.label}
                              </Select.Item>
                            ))}
                          </Select.Content>
                        </Select.Root>
                      )}
                    </Table.Cell>
                    <Table.Cell>
                      {!isSelf && (
                        <HStack>
                          <Button
                            size="sm"
                            colorPalette="brand"
                            disabled={!changed}
                            loading={row.saving}
                            onClick={() => saveRole(row)}
                          >
                            Sačuvaj ulogu
                          </Button>
                          <Button
                            size="sm"
                            variant="outline"
                            colorPalette={row.isActive ? "red" : "green"}
                            loading={row.saving}
                            onClick={() => toggleStatus(row)}
                          >
                            {row.isActive ? "Deaktiviraj" : "Aktiviraj"}
                          </Button>
                        </HStack>
                      )}
                      {isSelf && (
                        <Text fontSize="xs" color="fg.muted">
                          Ne možete menjati sopstveni nalog.
                        </Text>
                      )}
                    </Table.Cell>
                  </Table.Row>
                );
              })}
            </Table.Body>
          </Table.Root>
        </Box>
      </Container>
    </>
  );
}
