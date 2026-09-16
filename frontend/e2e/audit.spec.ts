import { test, expect } from "@playwright/test";
import { USERS, login } from "./helpers";

test("hr can open the audit log and see entries", async ({ page }) => {
  await login(page, USERS.hr);
  await page.goto("/admin/audit");

  await expect(page.getByRole("heading", { name: "Audit log" })).toBeVisible();
  await expect(page.getByRole("table")).toBeVisible();
  // the seeded admin actions (e.g. employee.create) guarantee at least one row
  await expect(page.locator("tbody tr").first()).toBeVisible();
});

test("employee cannot open the audit log", async ({ page }) => {
  await login(page, USERS.employee);
  await page.goto("/admin/audit");
  await expect(page).toHaveURL(/\/unauthorized/);
});
