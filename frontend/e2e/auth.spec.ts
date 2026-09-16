import { test, expect } from "@playwright/test";
import { USERS, login } from "./helpers";

test("employee can log in and lands on home", async ({ page }) => {
  await login(page, USERS.employee);
  await expect(page).toHaveURL(/\/home/);
  await expect(
    page.getByRole("heading", { name: /Dobrodošli/ }),
  ).toBeVisible();
});

test("invalid credentials show an error and stay on login", async ({
  page,
}) => {
  await page.goto("/login");
  await page.getByPlaceholder("email@hr-sistem.com").fill(USERS.employee);
  await page.getByPlaceholder("••••••••").fill("WrongPass1!");
  await page.getByRole("button", { name: "Prijavi se" }).click();

  await expect(page.getByText("Invalid credentials")).toBeVisible();
  await expect(page).toHaveURL(/\/login/);
});
