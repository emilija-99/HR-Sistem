import { useState, useEffect } from "react";
import { Link as RouterLink } from "react-router-dom";
import { api } from "@/api/client";
import {
  Container,
  Heading,
  Table,
  Badge,
  Link as ChakraLink,
  Spinner,
  Box,
  HStack,
  Text,
  Button,
} from "@chakra-ui/react";

interface Employee {
  id: number;
  user_id: number;
  first_name: string;
  last_name: string;
  city: string | null;
  hire_date: string | null;
  position_id: number | null;
  position_title: string | null;
  position_level: string | null;
  department_id: number | null;
  department_name: string | null;
  supervisor_id: number | null;
  supervisor_name: string | null;
}

export default function EmployeesPage() {
  const [employees, setEmployees] = useState<Employee[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api("/api/v1/employees")
      .then(setEmployees)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  if (loading)
    return (
      <>
        <Container py={10} textAlign="center">
          <Spinner size="xl" />
        </Container>
      </>
    );

  return (
    <>
      <Container maxW="container.xl" py={6}>
        <HStack justify="space-between" mb={6}>
          <Heading size="lg">Zaposleni</Heading>
          <Button asChild colorPalette="brand" size="sm">
            <RouterLink to="/employees/new">+ Novi zaposleni</RouterLink>
          </Button>
        </HStack>
        <Box overflowX="auto">
          <Table.Root variant="outline">
            <Table.Header>
              <Table.Row>
                <Table.ColumnHeader>Ime i prezime</Table.ColumnHeader>
                <Table.ColumnHeader>Departman</Table.ColumnHeader>
                <Table.ColumnHeader>Pozicija</Table.ColumnHeader>
                <Table.ColumnHeader>Nivo</Table.ColumnHeader>
                <Table.ColumnHeader>Nadređeni</Table.ColumnHeader>
                <Table.ColumnHeader>Grad</Table.ColumnHeader>
                <Table.ColumnHeader>Datum zaposlenja</Table.ColumnHeader>
                <Table.ColumnHeader></Table.ColumnHeader>
              </Table.Row>
            </Table.Header>
            <Table.Body>
              {employees.map((emp) => (
                <Table.Row key={emp.id}>
                  <Table.Cell>
                    <Text fontWeight="medium">
                      {emp.first_name} {emp.last_name}
                    </Text>
                  </Table.Cell>
                  <Table.Cell>
                    {emp.department_name ? (
                      <Badge colorPalette="brand">{emp.department_name}</Badge>
                    ) : (
                      "-"
                    )}
                  </Table.Cell>
                  <Table.Cell>{emp.position_title || "-"}</Table.Cell>
                  <Table.Cell>
                    {emp.position_level ? (
                      <Badge
                        colorPalette={
                          emp.position_level === "LEAD"
                            ? "orange"
                            : emp.position_level === "SENIOR"
                              ? "red"
                              : emp.position_level === "MEDIOR"
                                ? "blue"
                                : "gray"
                        }
                      >
                        {emp.position_level}
                      </Badge>
                    ) : (
                      "-"
                    )}
                  </Table.Cell>
                  <Table.Cell>{emp.supervisor_name || "-"}</Table.Cell>
                  <Table.Cell>{emp.city || "-"}</Table.Cell>
                  <Table.Cell>{emp.hire_date || "-"}</Table.Cell>
                  <Table.Cell>
                    <ChakraLink asChild colorPalette="brand">
                      <RouterLink to={`/employees/${emp.id}`}>
                        Detalji
                      </RouterLink>
                    </ChakraLink>
                  </Table.Cell>
                </Table.Row>
              ))}
            </Table.Body>
          </Table.Root>
          {employees.length === 0 && (
            <Text textAlign="center" py={10} color="fg.muted">
              Nema zaposlenih.
            </Text>
          )}
        </Box>
      </Container>
    </>
  );
}
