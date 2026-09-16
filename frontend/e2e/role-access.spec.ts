import { test, expect } from "@playwright/test";
import { USERS, login } from "./helpers";

test("employee sees no admin links and is blocked from /employees", async ({
  page,
}) => {
  await login(page, USERS.employee);

  await expect(
    page.getByRole("link", { name: "Zaposleni", exact: true }),
  ).toHaveCount(0);
  await expect(
    page.getByRole("link", { name: "Korisnici", exact: true }),
  ).toHaveCount(0);
  await expect(
    page.getByRole("link", { name: "Audit log", exact: true }),
  ).toHaveCount(0);

  await page.goto("/employees");
  await expect(page).toHaveURL(/\/unauthorized/);
});

test("hr sees employee and audit links but not platform-only users", async ({
  page,
}) => {
  await login(page, USERS.hr);

  await expect(
    page.getByRole("link", { name: "Zaposleni", exact: true }),
  ).toBeVisible();
  await expect(
    page.getByRole("link", { name: "Audit log", exact: true }),
  ).toBeVisible();
  await expect(
    page.getByRole("link", { name: "Korisnici", exact: true }),
  ).toHaveCount(0);
});

test("platform admin sees the users link", async ({ page }) => {
  await login(page, USERS.admin);
  await expect(
    page.getByRole("link", { name: "Korisnici", exact: true }),
  ).toBeVisible();
});
