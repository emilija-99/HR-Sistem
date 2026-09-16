import { defineConfig, devices } from "@playwright/test";

// Default target is the containerized app (nginx on :8080).
// Set E2E_TARGET=dev to run against the Vite dev server (:5173), which Playwright
// will start automatically via webServer.
const useDevServer = process.env.E2E_TARGET === "dev";

const baseURL =
  process.env.PLAYWRIGHT_BASE_URL ||
  (useDevServer ? "http://localhost:5173" : "http://localhost:8080");

export default defineConfig({
  testDir: "./e2e",
  timeout: 30_000,
  expect: { timeout: 7_000 },
  // Tests share one backend/database, so run serially.
  fullyParallel: false,
  workers: 1,
  retries: process.env.CI ? 1 : 0,
  reporter: [["list"]],
  globalSetup: "./e2e/global-setup.ts",
  globalTeardown: "./e2e/global-teardown.ts",
  use: {
    baseURL,
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
  },
  webServer: useDevServer
    ? {
        command: "npm run dev",
        url: "http://localhost:5173",
        reuseExistingServer: true,
        timeout: 120_000,
      }
    : undefined,
  projects: [{ name: "chromium", use: { ...devices["Desktop Chrome"] } }],
});
