import { execSync } from "node:child_process";

// (Re)create the fixed E2E accounts before the suite runs.
export default function globalSetup() {
  execSync("sh ../scripts/seed-test-data.sh", { stdio: "inherit" });
}
