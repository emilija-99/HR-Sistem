import { Page, expect } from "@playwright/test";

export const PASSWORD = process.env.E2E_PASSWORD || "E2eTest1!";

export const USERS = {
  admin: "e2e-admin@hr-sistem.com",
  hr: "e2e-hr@hr-sistem.com",
  employee: "e2e-employee@hr-sistem.com",
};

/** Log in through the UI and wait until we leave the login page. */
export async function login(page: Page, email: string, password = PASSWORD) {
  await page.goto("/login");
  await page.getByPlaceholder("email@hr-sistem.com").fill(email);
  await page.getByPlaceholder("••••••••").fill(password);
  await page.getByRole("button", { name: "Prijavi se" }).click();
  await page.waitForURL(/\/(home|onboarding)/, { timeout: 15_000 });
  await expect(page).not.toHaveURL(/\/login/);
}

/**
 * Chakra v3 selects are not native elements: open the trigger by its
 * placeholder text and click the option by its label.
 */
export async function chooseByPlaceholder(
  page: Page,
  placeholder: string,
  option: string | RegExp,
) {
  await page.getByText(placeholder, { exact: true }).first().click();
  await page.getByText(option).first().click();
}
