import { test, expect } from "@playwright/test";
import { USERS, login } from "./helpers";

// Privremena provera brendiranja: boje i font iz dizajn specifikacije.
// #123B68 -> rgb(18, 59, 104)   (navy)
// #1674D1 -> rgb(22, 116, 209)  (primarna plava)
// #5D7190 -> rgb(93, 113, 144)  (sekundarni tekst)
// #F4F8FC -> rgb(244, 248, 252) (soft blue pozadina)
// #E4EEFA -> rgb(228, 238, 250) (brand.subtle)

test("login strana koristi brend paletu i Inter", async ({ page }) => {
  await page.goto("/login");
  await expect(page).toHaveTitle("HR Sistem");

  const font = await page.evaluate(
    () => getComputedStyle(document.body).fontFamily,
  );
  expect(font).toContain("Inter");

  // Pozadina aplikacije je soft blue (#F4F8FC)
  await expect(page.locator("body")).toHaveCSS(
    "background-color",
    "rgb(244, 248, 252)",
  );

  await expect(page.getByAltText("HR Sistem")).toBeVisible();

  // Primarno dugme je primary blue (#1674D1)
  await expect(page.getByRole("button", { name: "Prijavi se" })).toHaveCSS(
    "background-color",
    "rgb(22, 116, 209)",
  );
});

test("navigacija koristi brend boje", async ({ page }) => {
  await login(page, USERS.employee);

  const brandLink = page.locator('a[href="/home"]').first();
  await expect(brandLink.getByText("HR", { exact: true })).toHaveCSS(
    "color",
    "rgb(18, 59, 104)",
  );
  await expect(brandLink.getByText("Sistem", { exact: true })).toHaveCSS(
    "color",
    "rgb(22, 116, 209)",
  );

  // aktivna ruta ("Početna") je brand.600 na brand.subtle pozadini
  const active = page.getByRole("link", { name: "Početna" }).locator("p");
  await expect(active).toHaveCSS("color", "rgb(22, 116, 209)");

  // neaktivna ruta je u sekundarnoj boji teksta (#5D7190)
  const inactive = page.getByRole("link", { name: "Moj profil" }).locator("p");
  await expect(inactive).toHaveCSS("color", "rgb(93, 113, 144)");
});
