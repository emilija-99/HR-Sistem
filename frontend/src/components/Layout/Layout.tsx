import { Outlet } from "react-router-dom";
import Menu from "@/components/Menu/Menu";
import { Box } from "@chakra-ui/react";

/**
 * Zajednički okvir za sve zaštićene stranice: navigacija je **uvek** vidljiva i
 * označava aktivnu stranicu, a stranica se renderuje ispod nje kroz `<Outlet />`.
 * Zato stranice više ne uključuju `<Menu />` svaka za sebe.
 */
export default function Layout() {
  return (
    <Box minH="100vh" bg="bg.subtle">
      <Menu />
      <Outlet />
    </Box>
  );
}
