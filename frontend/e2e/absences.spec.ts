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

test("a single day (start == end) is accepted", async ({ page }) => {
  await login(page, USERS.employee);
  await page.goto("/absences");

  await page.getByText("Izaberi tip", { exact: true }).click();
  await page.getByText(/Vacation/).first().click();

  // Monday 2026-12-07, both fields the same date → one day off
  const dates = page.locator('input[type="date"]');
  await dates.nth(0).fill("2026-12-07");
  await dates.nth(1).fill("2026-12-07");

  await expect(page.getByText(/za izabrani period: 1 radni dan/)).toBeVisible();
  await expect(page.getByRole("button", { name: "Podnesi zahtev" })).toBeEnabled();
});

test("weekend dates are rejected before the request is sent", async ({
  page,
}) => {
  await login(page, USERS.employee);
  await page.goto("/absences");

  await page.getByText("Izaberi tip", { exact: true }).click();
  await page.getByText(/Vacation/).first().click();

  // Saturday 2026-11-14 .. Sunday 2026-11-15
  const dates = page.locator('input[type="date"]');
  await dates.nth(0).fill("2026-11-14");
  await dates.nth(1).fill("2026-11-15");

  await expect(
    page.getByText("Početni datum ne može biti subota ili nedelja."),
  ).toBeVisible();
  await expect(
    page.getByRole("button", { name: "Podnesi zahtev" }),
  ).toBeDisabled();
});

test("a request cannot exceed the available days", async ({ page }) => {
  await login(page, USERS.employee);
  await page.goto("/absences");

  await page.getByText("Izaberi tip", { exact: true }).click();
  await page.getByText(/Vacation/).first().click();

  // ~165 working days — far beyond any balance this employee has
  const dates = page.locator('input[type="date"]');
  await dates.nth(0).fill("2026-11-09");
  await dates.nth(1).fill("2027-06-30");

  await expect(page.getByText(/na raspolaganju je/)).toBeVisible();
  await expect(
    page.getByRole("button", { name: "Podnesi zahtev" }),
  ).toBeDisabled();
});
