import { test, expect } from "@playwright/test";
import { USERS, login } from "./helpers";

// The manual rollover is time-boxed to the last week of December / first week of
// January, so outside that window the control must be disabled (the server
// rejects it too). Tests are expected to run outside that window.
test("hr cannot run manual rollover outside the December/January window", async ({
  page,
}) => {
  await login(page, USERS.hr);
  await page.goto("/balance");

  await expect(
    page.getByRole("heading", { name: "Godišnji prenos (rollover)" }),
  ).toBeVisible();

  const button = page.getByRole("button", { name: "Pokreni prenos" });
  await expect(button).toBeDisabled();

  await expect(page.getByText("Trenutno nije dostupno")).toBeVisible();
  await expect(
    page.getByText(/poslednjoj nedelji decembra/i),
  ).toBeVisible();
});
