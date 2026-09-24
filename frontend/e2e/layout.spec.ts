import { test, expect } from "@playwright/test";
import { USERS, login } from "./helpers";

/**
 * Zahtev: navigacija mora da bude vidljiva na svakoj stranici i da označi
 * stranicu na kojoj se korisnik nalazi (aria-current="page").
 */
test("navigacija je vidljiva na svakoj stranici i označava aktivnu", async ({
  page,
}) => {
  await login(page, USERS.hr);

  const pages: [string, string][] = [
    ["/home", "Početna"],
    ["/absences", "Odsustva"],
    ["/balance", "Dostupni dani"],
    ["/employees", "Zaposleni"],
    ["/profile", "Moj profil"],
    ["/attendance/admin", "Prisustvo"],
    ["/admin/audit", "Audit log"],
  ];

  for (const [path, label] of pages) {
    await page.goto(path);
    const link = page.getByRole("link", { name: label, exact: true });
    await expect(link).toBeVisible();
    await expect(link).toHaveAttribute("aria-current", "page");
  }
});

test("navigacija označava samo jednu stavku", async ({ page }) => {
  await login(page, USERS.hr);
  await page.goto("/absences/approvals");

  await expect(
    page.getByRole("link", { name: "Odobravanja", exact: true }),
  ).toHaveAttribute("aria-current", "page");
  // /absences ne sme biti označen kada smo na /absences/approvals
  await expect(
    page.getByRole("link", { name: "Odsustva", exact: true }),
  ).not.toHaveAttribute("aria-current", "page");
});
