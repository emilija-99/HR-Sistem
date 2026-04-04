import Menu from "../../components/Menu/Menu";
import { Button } from "@chakra-ui/react";

export default function HomePage() {
  return (
    <>
      <Menu />
      <div style={{ padding: 20 }}>
        <h1>Welcome</h1>
        <p>React routing works.</p>
        <Button color="green" variant="surface">
          Click me
        </Button>
      </div>
    </>
  );
}
