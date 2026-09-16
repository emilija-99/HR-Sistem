import { test, expect } from "@playwright/test";
import { USERS, login } from "./helpers";

test("employee can save a draft and submit it", async ({ page }) => {
  await login(page, USERS.employee);
  await page.goto("/absences");

  // choose absence type
  await page.getByText("Izaberi tip", { exact: true }).click();
  await page.getByText(/Vacation/).first().click();

  // Mon 2026-11-09 .. Fri 2026-11-13 (5 business days, within the 15 granted)
  const dates = page.locator('input[type="date"]');
  await dates.nth(0).fill("2026-11-09");
  await dates.nth(1).fill("2026-11-13");

  await page.getByRole("button", { name: "Sačuvaj kao nacrt" }).click();
  await expect(page.getByText("2026-11-09 → 2026-11-13")).toBeVisible();
  await expect(page.getByText("DRAFT").first()).toBeVisible();

  // submit the draft
  await page.getByRole("button", { name: "Podnesi", exact: true }).first().click();
  await expect(page.getByText("PENDING").first()).toBeVisible();
});

test("submitting a request with no working days is rejected", async ({
  page,
}) => {
  await login(page, USERS.employee);
  await page.goto("/absences");

  await page.getByText("Izaberi tip", { exact: true }).click();
  await page.getByText(/Vacation/).first().click();

  // a weekend only
  const dates = page.locator('input[type="date"]');
  await dates.nth(0).fill("2026-11-14");
  await dates.nth(1).fill("2026-11-15");

  await page.getByRole("button", { name: "Podnesi zahtev" }).click();
  await expect(page.getByText("Invalid dates")).toBeVisible();
});
