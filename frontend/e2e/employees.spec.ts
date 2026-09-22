import { test, expect } from "@playwright/test";
import { PASSWORD, USERS, login } from "./helpers";

test("hr creates an employee and it appears in the grid with its supervisor", async ({
  page,
}) => {
  const stamp = Date.now();
  const email = `e2e-created-${stamp}@hr-sistem.com`;

  await login(page, USERS.hr);
  await page.goto("/employees/new");

  await page.locator('input[type="email"]').nth(0).fill(email);
  await page.locator('input[type="password"]').fill(PASSWORD);

  const textInputs = page.locator(
    'input:not([type="email"]):not([type="password"]):not([type="date"])',
  );
  await textInputs.nth(0).fill("E2E");
  await textInputs.nth(1).fill("Kreirani");

  // country (required) — first select on the form
  await page.getByText("Izaberi državu", { exact: true }).click();
  await page.getByText("Serbia", { exact: true }).first().click();

  // position (required)
  await page.getByText("Izaberi poziciju", { exact: true }).click();
  await page.getByText(/Backend Engineer/).first().click();

  // obavezne datume (rođenje ≥16 godina, zaposlenje)
  const dates = page.locator('input[type="date"]');
  await dates.nth(0).fill("1995-05-05");
  await dates.nth(1).fill("2026-01-15");

  await page.getByRole("button", { name: "Kreiraj zaposlenog" }).click();
  await page.waitForURL(/\/employees\/\d+/, { timeout: 15_000 });

  // the new employee shows up in the grid
  await page.goto("/employees");
  await expect(page.getByText("E2E Kreirani")).toBeVisible();

  // the seeded employee reports to E2E HR (supervisor column)
  const row = page.locator("tr", { hasText: "E2E Employee" });
  await expect(row).toContainText("E2E HR");
});
