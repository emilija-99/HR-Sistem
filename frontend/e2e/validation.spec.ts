import { test, expect } from "@playwright/test";
import { USERS, login } from "./helpers";

const REQUIRED = "Ovo polje je obavezno.";

test("register: empty form marks every field as required", async ({ page }) => {
  await page.goto("/register");
  await page.getByRole("button", { name: "Registruj se" }).click();

  // email, lozinka, potvrda
  await expect(page.getByText(REQUIRED)).toHaveCount(3);
  await expect(
    page.getByText("Polja moraju da budu popunjena. Označena polja su obavezna."),
  ).toBeVisible();
});

test("register: email without a TLD is rejected", async ({ page }) => {
  await page.goto("/register");
  await page.locator('input[type="email"]').fill("novivijdfslfjsdlfjsdlfkj@dsadas");
  await page.locator('input[type="password"]').nth(0).fill("TestTest1!");
  await page.locator('input[type="password"]').nth(1).fill("TestTest1!");
  await page.getByRole("button", { name: "Registruj se" }).click();

  await expect(page.getByText("Unesite ispravnu email adresu.")).toBeVisible();
  await expect(page).toHaveURL(/\/register/);
});

test("register: mismatched passwords mark the confirm field", async ({
  page,
}) => {
  await page.goto("/register");
  await page.locator('input[type="email"]').fill("pera@hr-sistem.com");
  await page.locator('input[type="password"]').nth(0).fill("TestTest1!");
  await page.locator('input[type="password"]').nth(1).fill("DrugaLozinka1!");
  await page.getByRole("button", { name: "Registruj se" }).click();

  await expect(page.getByText("Lozinke se ne poklapaju.")).toBeVisible();
  await expect(page).toHaveURL(/\/register/);
});

test("login: empty form marks both fields as required", async ({ page }) => {
  await page.goto("/login");
  await page.getByRole("button", { name: "Prijavi se" }).click();

  await expect(page.getByText(REQUIRED)).toHaveCount(2);
  await expect(page).toHaveURL(/\/login/);
});

test("hr new employee: empty form is rejected with a message", async ({
  page,
}) => {
  await login(page, USERS.hr);
  await page.goto("/employees/new");

  await page.getByRole("button", { name: "Kreiraj zaposlenog" }).click();

  await expect(
    page.getByText("Molimo vas popunite polja kako bi napravili nalog za vas."),
  ).toBeVisible();
  await expect(page.getByText(REQUIRED).first()).toBeVisible();
  await expect(page).toHaveURL(/\/employees\/new/);
});

test("register: invalid fields are coloured red", async ({ page }) => {
  await page.goto("/register");
  await page.getByRole("button", { name: "Registruj se" }).click();

  const email = page.locator('input[type="email"]');
  await expect(email).toHaveAttribute("aria-invalid", "true");
  // Chakra `border.error` = red.500 (#ef4444)
  await expect(email).toHaveCSS("border-color", "rgb(239, 68, 68)");
});

test("hr new employee: empty submit colours invalid fields red", async ({
  page,
}) => {
  await login(page, USERS.hr);
  await page.goto("/employees/new");
  await page.getByRole("button", { name: "Kreiraj zaposlenog" }).click();

  const email = page.locator('input[type="email"]').nth(0);
  await expect(email).toHaveCSS("border-color", "rgb(239, 68, 68)");

  // i Select (combobox) dobija crvenu ivicu preko aria-invalid
  const country = page.getByRole("combobox").first();
  await expect(country).toHaveAttribute("aria-invalid", "true");
  await expect(country).toHaveCSS("border-color", "rgb(239, 68, 68)");
});

test("register: the incomplete message disappears once the form is valid", async ({
  page,
}) => {
  await page.goto("/register");
  await page.getByRole("button", { name: "Registruj se" }).click();

  const message = page.getByText(
    "Polja moraju da budu popunjena. Označena polja su obavezna.",
  );
  await expect(message).toBeVisible();

  await page.locator('input[type="email"]').fill("pera@hr-sistem.com");
  await page.locator('input[type="password"]').nth(0).fill("TestTest1!");
  await page.locator('input[type="password"]').nth(1).fill("TestTest1!");

  // poruka se sklanja sama, bez ponovnog klika
  await expect(message).toHaveCount(0);
});

test("employee edit: phone takes digits only (max 10) and email must be valid", async ({
  page,
}) => {
  await login(page, USERS.hr);
  await page.goto("/profile");
  await page.getByRole("button", { name: "Izmeni" }).click();

  const phone = page.getByLabel("Telefon");
  await phone.fill("abc123def");
  await expect(phone).toHaveValue("123");

  await phone.fill("123456789012345");
  await expect(phone).toHaveValue("1234567890");

  await page.getByLabel("Privatni email").fill("nije-email");
  await page.getByRole("button", { name: "Sačuvaj" }).click();

  await expect(page.getByText("Unesite ispravnu email adresu.")).toBeVisible();
});

test("hr new employee: a birth date under 16 is rejected", async ({ page }) => {
  await login(page, USERS.hr);
  await page.goto("/employees/new");

  // 2015 → sigurno mlađe od 16 godina
  await page.locator('input[type="date"]').nth(0).fill("2015-01-01");
  await page.getByRole("button", { name: "Kreiraj zaposlenog" }).click();

  await expect(
    page.getByText("Zaposleni mora imati najmanje 16 godina."),
  ).toBeVisible();
  await expect(page).toHaveURL(/\/employees\/new/);
});
