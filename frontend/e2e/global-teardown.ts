import { execSync } from "node:child_process";

// Remove the E2E accounts and their data after the suite. Set E2E_KEEP_DATA=1
// to keep everything for inspection.
export default function globalTeardown() {
  if (process.env.E2E_KEEP_DATA === "1") return;
  execSync("sh ../scripts/seed-test-data.sh --clean", { stdio: "inherit" });
}
