import { useState, useEffect, useMemo } from "react";
import { api } from "@/api/client";
import { useAuth } from "@/providers/AuthProvider";
import {
  Container, Heading, Card, VStack, HStack, Field, Input, Button,
  Text, Badge, Spinner, Table, Box, Select, createListCollection,
} from "@chakra-ui/react";

interface BalanceSummary {
  absence_type_id: number;
  type_name: string;
  code: string;
  is_paid: boolean;
  granted_days: number;
  used_days: number;
  available_days: number;
}

interface AbsenceType {
  id: number;
  type_name: string;
}

interface MyPolicy {
  absence_type_id: number;
  policy: {
    id: number;
    name: string;
    grant_policy: string;
    days_per_period: number | null;
    allow_carry_over: boolean;
    carry_over_max_days: number | null;
    carry_over_expiry_month: number | null;
    carry_over_expiry_day: number | null;
    requires_balance: boolean;
  };
}

interface Employee {
  id: number;
  first_name: string;
  last_name: string;
  department_id: number | null;
  department_name: string | null;
}

interface DepartmentGroup {
  name: string;
  employees: Employee[];
}

const ADMIN_ROLES = ["PLATFORM_ADMIN", "HR_ADMIN"];

// Manual rollover is only offered in the last week of December (targeting the
// current year) and the first week of January (targeting the previous year).
// The server enforces the same window.
function rolloverTargetYear(now = new Date()): number | null {
  const month = now.getMonth() + 1;
  const day = now.getDate();
  if (month === 12 && day >= 25) return now.getFullYear();
  if (month === 1 && day <= 7) return now.getFullYear() - 1;
  return null;
}

export default function BalancePage() {
  const { user } = useAuth();
  const isAdmin = user && ADMIN_ROLES.includes(user.role);

  const [balance, setBalance] = useState<BalanceSummary[]>([]);
  const [types, setTypes] = useState<AbsenceType[]>([]);
  const [employees, setEmployees] = useState<Employee[]>([]);
  const [myPolicies, setMyPolicies] = useState<MyPolicy[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [granting, setGranting] = useState(false);
  const [rolloverBusy, setRolloverBusy] = useState(false);
  const [rolloverResult, setRolloverResult] = useState("");
  const rolloverYear = rolloverTargetYear();
  const [grant, setGrant] = useState({
    employee_id: 0,
    absence_type_id: 0,
    days: 0,
  });

  const typeCollection = createListCollection({
    items: types,
    itemToString: (item) => item.type_name,
    itemToValue: (item) => String(item.id),
  });

  const employeeCollection = createListCollection({
    items: employees,
    itemToString: (item) => `${item.first_name} ${item.last_name}`,
    itemToValue: (item) => String(item.id),
  });

  // group employees by department for the grouped select
  const departmentGroups = useMemo<DepartmentGroup[]>(() => {
    const map = new Map<string, Employee[]>();
    for (const emp of employees) {
      const key = emp.department_name || "Bez departmana";
      if (!map.has(key)) map.set(key, []);
      map.get(key)!.push(emp);
    }
    return Array.from(map.entries())
      .map(([name, emps]) => ({
        name,
        employees: emps.sort((a, b) =>
          `${a.first_name} ${a.last_name}`.localeCompare(
            `${b.first_name} ${b.last_name}`,
          ),
        ),
      }))
      .sort((a, b) => a.name.localeCompare(b.name));
  }, [employees]);

  const fetchBalance = () =>
    api("/api/v1/absences/balance/me").then(setBalance);

  useEffect(() => {
    Promise.all([
      fetchBalance(),
      api("/api/v1/absences/types").then((d) => setTypes(d || [])),
      api("/api/v1/absences/balance/my-policies").then(setMyPolicies).catch(() => {}),
      isAdmin ? api("/api/v1/employees").then(setEmployees) : Promise.resolve(),
    ])
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, []);

  const handleRollover = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!rolloverYear) return;
    setRolloverBusy(true);
    setError("");
    setRolloverResult("");
    try {
      const res = await api("/api/v1/absences/balance/rollover", {
        method: "POST",
        body: JSON.stringify({ year: rolloverYear }),
      });
      setRolloverResult(
        `Godišnji prenos ${res.year}: dodeljeno ${res.accrued.length} zaposlenih, preneto ${res.carried.length}, isteklo ${res.expired.length} zapisa.`,
      );
      await fetchBalance();
    } catch (err: any) {
      setError(err.message || "Rollover failed");
    } finally {
      setRolloverBusy(false);
    }
  };

  const handleGrant = async (e: React.FormEvent) => {
    e.preventDefault();
    setGranting(true);
    setError("");
    try {
      await api("/api/v1/absences/balance/grant", {
        method: "POST",
        body: JSON.stringify({
          employee_id: grant.employee_id,
          absence_type_id: grant.absence_type_id,
          days: grant.days,
        }),
      });
      setGrant({ employee_id: 0, absence_type_id: 0, days: 0 });
      setError("Dani su dodeljeni ✓");
    } catch (err: any) {
      setError(err.message || "Failed to grant days");
    } finally {
      setGranting(false);
    }
  };

  if (loading)
    return (
      <Container py={10} textAlign="center">
        <Spinner size="xl" />
      </Container>
    );

  return (
    <Container maxW="container.lg" py={6}>
      <Heading mb={6}>Dostupni dani odsustva</Heading>

      {error && (
        <Text color={error.includes("✓") ? "green.500" : "red.500"} fontSize="sm" mb={4}>
          {error}
        </Text>
      )}

      {balance.length === 0 ? (
        <Text color="fg.muted" mb={6}>
          Još nema odobrenih dana. Administracija treba da ti dodeli godišnji odmor.
        </Text>
      ) : (
        <Box overflowX="auto" mb={8}>
          <Table.Root variant="outline">
            <Table.Header>
              <Table.Row>
                <Table.ColumnHeader>Tip</Table.ColumnHeader>
                <Table.ColumnHeader>Plaćeno</Table.ColumnHeader>
                <Table.ColumnHeader>Dodeljeno</Table.ColumnHeader>
                <Table.ColumnHeader>Iskorišćeno</Table.ColumnHeader>
                <Table.ColumnHeader>Dostupno</Table.ColumnHeader>
              </Table.Row>
            </Table.Header>
            <Table.Body>
              {balance.map((b) => (
                <Table.Row key={b.absence_type_id}>
                  <Table.Cell>
                    <Text fontWeight="medium">{b.type_name}</Text>
                  </Table.Cell>
                  <Table.Cell>
                    {b.is_paid ? (
                      <Badge colorPalette="green">plaćeno</Badge>
                    ) : (
                      <Badge colorPalette="orange">neplaćeno</Badge>
                    )}
                  </Table.Cell>
                  <Table.Cell>{b.granted_days}</Table.Cell>
                  <Table.Cell>{b.used_days}</Table.Cell>
                  <Table.Cell>
                    <Badge
                      colorPalette={
                        b.available_days <= 0
                          ? "red"
                          : b.available_days <= 3
                            ? "yellow"
                            : "green"
                      }
                    >
                      {b.available_days} dana
                    </Badge>
                  </Table.Cell>
                </Table.Row>
              ))}
            </Table.Body>
          </Table.Root>
        </Box>
      )}

      {myPolicies.length > 0 && (
        <Box overflowX="auto" mb={8}>
          <Heading size="md" mb={3}>
            Pravila (leave policies)
          </Heading>
          <Table.Root variant="outline">
            <Table.Header>
              <Table.Row>
                <Table.ColumnHeader>Tip</Table.ColumnHeader>
                <Table.ColumnHeader>Politika</Table.ColumnHeader>
                <Table.ColumnHeader>Dodeljeno godišnje</Table.ColumnHeader>
                <Table.ColumnHeader>Prenos</Table.ColumnHeader>
              </Table.Row>
            </Table.Header>
            <Table.Body>
              {myPolicies.map((mp) => (
                <Table.Row key={mp.absence_type_id}>
                  <Table.Cell>{typeName(mp.absence_type_id)}</Table.Cell>
                  <Table.Cell>
                    <Text fontWeight="medium">{mp.policy.name}</Text>
                  </Table.Cell>
                  <Table.Cell>
                    {mp.policy.days_per_period ? (
                      `${mp.policy.days_per_period} dana`
                    ) : (
                      <Badge colorPalette="brand">neograničeno</Badge>
                    )}
                  </Table.Cell>
                  <Table.Cell>
                    {mp.policy.allow_carry_over ? (
                      <Badge colorPalette="green">
                        do {mp.policy.carry_over_max_days} dana
                      </Badge>
                    ) : (
                      <Badge colorPalette="gray">nema</Badge>
                    )}
                  </Table.Cell>
                </Table.Row>
              ))}
            </Table.Body>
          </Table.Root>
        </Box>
      )}

      {isAdmin && (
        <>
        <Card.Root>
          <Card.Header>
            <Heading size="md">Dodeli dane (godišnji odmor)</Heading>
          </Card.Header>
          <Card.Body>
            <form onSubmit={handleGrant}>
              <VStack gap={4} align="stretch">
                <Field.Root required>
                  <Field.Label>Zaposleni</Field.Label>
                  <Select.Root
                    collection={employeeCollection}
                    value={grant.employee_id ? [String(grant.employee_id)] : []}
                    onValueChange={(e: any) =>
                      setGrant((p) => ({
                        ...p,
                        employee_id: parseInt(e.value[0]) || 0,
                      }))
                    }
                  >
                    <Select.Trigger>
                      <Select.ValueText placeholder="Izaberi zaposlenog" />
                    </Select.Trigger>
                    <Select.Content>
                      {departmentGroups.map((group) => (
                        <Select.ItemGroup key={group.name}>
                          <Select.ItemGroupLabel>
                            {group.name}
                          </Select.ItemGroupLabel>
                          {group.employees.map((emp) => (
                            <Select.Item key={emp.id} item={emp}>
                              {emp.first_name} {emp.last_name}
                            </Select.Item>
                          ))}
                        </Select.ItemGroup>
                      ))}
                    </Select.Content>
                  </Select.Root>
                </Field.Root>
                <Field.Root required>
                  <Field.Label>Tip odsustva</Field.Label>
                  <Select.Root
                    collection={typeCollection}
                    value={grant.absence_type_id ? [String(grant.absence_type_id)] : []}
                    onValueChange={(e: any) =>
                      setGrant((p) => ({
                        ...p,
                        absence_type_id: parseInt(e.value[0]) || 0,
                      }))
                    }
                  >
                    <Select.Trigger>
                      <Select.ValueText placeholder="Izaberi tip" />
                    </Select.Trigger>
                    <Select.Content>
                      {typeCollection.items.map((t) => (
                        <Select.Item key={t.id} item={t}>
                          {t.type_name}
                        </Select.Item>
                      ))}
                    </Select.Content>
                  </Select.Root>
                </Field.Root>
                <Field.Root required>
                  <Field.Label>Broj dana</Field.Label>
                  <Input
                    type="number"
                    step="0.5"
                    value={grant.days || ""}
                    onChange={(e) =>
                      setGrant((p) => ({
                        ...p,
                        days: parseFloat(e.target.value) || 0,
                      }))
                    }
                    placeholder="npr. 20"
                  />
                </Field.Root>
                <Button
                  type="submit"
                  colorPalette="brand"
                  loading={granting}
                  disabled={
                    !grant.employee_id || !grant.absence_type_id || !grant.days
                  }
                >
                  Dodeli dane
                </Button>
              </VStack>
            </form>
          </Card.Body>
        </Card.Root>

        <Card.Root mt={6}>
          <Card.Header>
            <Heading size="md">Godišnji prenos (rollover)</Heading>
          </Card.Header>
          <Card.Body>
            <Text fontSize="sm" color="fg.muted" mb={4}>
              Dodeljuje dane za izabranu godinu, prenosi neiskorišćene dane (do
              limita politike) i ističe prekoračenja.
            </Text>
            <form onSubmit={handleRollover}>
              <HStack gap={4} align="center">
                <Text fontWeight="medium">
                  {rolloverYear
                    ? `Godišnji prenos za ${rolloverYear}. godinu`
                    : "Trenutno nije dostupno"}
                </Text>
                <Button
                  type="submit"
                  colorPalette="brand"
                  loading={rolloverBusy}
                  disabled={!rolloverYear}
                >
                  Pokreni prenos
                </Button>
              </HStack>
            </form>
            <Text fontSize="xs" color="fg.muted" mt={3}>
              Dostupno samo u poslednjoj nedelji decembra (tekuća godina) i prvoj
              nedelji januara (prethodna godina).
            </Text>
            {rolloverResult && (
              <Text color="green.600" fontSize="sm" mt={3}>
                {rolloverResult}
              </Text>
            )}
          </Card.Body>
        </Card.Root>
        </>
      )}
    </Container>
  );
}

function typeName(id: number): string {
  const map: Record<number, string> = {
    1: "Vacation",
    2: "Parental Leave",
    3: "Sick Leave",
    4: "Training Leave",
    5: "Disability Leave",
    6: "Personal Leave",
  };
  return map[id] || `Tip ${id}`;
}
