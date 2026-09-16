import { test, expect } from "@playwright/test";
import { PASSWORD, USERS, login } from "./helpers";

test("hr creates an employee and it appears in the grid with its supervisor", async ({
  page,
}) => {
  const stamp = Date.now();
  const email = `e2e-created-${stamp}@hr-sistem.com`;

  await login(page, USERS.hr);
  await page.goto("/employees/new");

  await page.locator('input[type="email"]').fill(email);
  await page.locator('input[type="password"]').fill(PASSWORD);

  const textInputs = page.locator(
    'input:not([type="email"]):not([type="password"]):not([type="date"])',
  );
  await textInputs.nth(0).fill("E2E");
  await textInputs.nth(1).fill("Kreirani");

  // country (required) — first select on the form
  await page.getByText("Izaberi državu", { exact: true }).click();
  await page.getByText("Serbia", { exact: true }).first().click();

  await page.getByRole("button", { name: "Kreiraj zaposlenog" }).click();
  await page.waitForURL(/\/employees\/\d+/, { timeout: 15_000 });

  // the new employee shows up in the grid
  await page.goto("/employees");
  await expect(page.getByText("E2E Kreirani")).toBeVisible();

  // the seeded employee reports to E2E HR (supervisor column)
  const row = page.locator("tr", { hasText: "E2E Employee" });
  await expect(row).toContainText("E2E HR");
});
