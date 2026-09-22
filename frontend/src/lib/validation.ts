import { formatISO } from "./dates";

/**
 * Klijentska pravila validacije formi.
 *
 * Ovo je **UX sloj** — server i dalje nezavisno validira (`validator.V.Struct`
 * u handlerima), pa se zaobilaženje klijenta ne isplati. Poruke su na srpskom
 * jer ih korisnik vidi direktno ispod polja.
 */

/** Maksimalna dužina imena/prezimena (kolona u bazi je 50). */
export const MAX_NAME = 20;
/** Telefon: samo cifre, najviše 10. */
export const MAX_PHONE = 10;
/** Minimalna starost zaposlenog. */
export const MIN_AGE = 16;
/** Lozinka: 8–25 znakova (isto kao `validate:"min=8,max=25"` na serveru). */
export const MIN_PASSWORD = 8;
export const MAX_PASSWORD = 25;

/**
 * Email mora da ima domen sa TLD-om, pa `pera@dsadas` NE prolazi
 * (a `pera@firma.rs` prolazi).
 */
export const EMAIL_RE =
  /^[A-Za-z0-9._%+-]+@[A-Za-z0-9-]+(\.[A-Za-z0-9-]+)*\.[A-Za-z]{2,}$/;

export const REQUIRED = "Ovo polje je obavezno.";

export const isBlank = (v: unknown): boolean =>
  v === null || v === undefined || String(v).trim() === "";

/** Za numeričke select-e gde je `0` = „nije izabrano". */
export const isUnset = (n: number): boolean => !n || n <= 0;

export const isEmail = (v: string): boolean => EMAIL_RE.test(v.trim());

/** Uklanja sve što nije cifra (za unos telefona). */
export const onlyDigits = (v: string): string => v.replace(/\D/g, "");

export const tooLong = (v: string, max: number): boolean =>
  v.trim().length > max;

/** Najstariji dozvoljeni datum rođenja (danas − `minAge` godina). */
export function birthDateLimit(minAge = MIN_AGE, now = new Date()): string {
  return formatISO(
    new Date(now.getFullYear() - minAge, now.getMonth(), now.getDate()),
  );
}

/** Da li je osoba rođena na `dob` napunila bar `minAge` godina. */
export function isOldEnough(dob: string, minAge = MIN_AGE): boolean {
  if (!dob) return false;
  return dob <= birthDateLimit(minAge);
}

export type Errors = Record<string, string>;

export const hasErrors = (errors: Errors): boolean =>
  Object.keys(errors).length > 0;

// ── pojedinačna pravila (deljena između formi) ──────────────────

/** Prazno → obavezno; predugačko → poruka sa limitom. */
export function nameError(label: string, value: string): string {
  if (isBlank(value)) return REQUIRED;
  if (tooLong(value, MAX_NAME))
    return `${label} može imati najviše ${MAX_NAME} karaktera.`;
  return "";
}

/** Opciono polje: greška samo ako je nešto uneto a nije ispravno. */
export function phoneError(value: string): string {
  if (isBlank(value)) return "";
  const trimmed = value.trim();
  if (onlyDigits(trimmed) !== trimmed)
    return "Telefon može sadržati samo cifre.";
  if (trimmed.length > MAX_PHONE)
    return `Telefon može imati najviše ${MAX_PHONE} cifara.`;
  return "";
}

/** Opciono polje: greška samo ako je nešto uneto a nije ispravno. */
export function privateEmailError(value: string): string {
  if (isBlank(value)) return "";
  return isEmail(value) ? "" : "Unesite ispravnu email adresu.";
}

// ── validatori po formi ─────────────────────────────────────────

/** Validacija naloga: email + lozinka (+ potvrda na registraciji). */
export function validateAccount(
  email: string,
  password: string,
  confirm?: string,
): Errors {
  const errors: Errors = {};

  if (isBlank(email)) errors.email = REQUIRED;
  else if (!isEmail(email)) errors.email = "Unesite ispravnu email adresu.";

  if (isBlank(password)) errors.password = REQUIRED;
  else if (password.length < MIN_PASSWORD || password.length > MAX_PASSWORD)
    errors.password = `Lozinka mora imati između ${MIN_PASSWORD} i ${MAX_PASSWORD} znakova.`;

  if (confirm !== undefined) {
    if (isBlank(confirm)) errors.confirm = REQUIRED;
    else if (confirm !== password) errors.confirm = "Lozinke se ne poklapaju.";
  }

  return errors;
}

export type ProfileForm = {
  first_name: string;
  last_name: string;
  phone_number: string;
  private_email: string;
  country: number;
  date_of_birth: string;
  hire_date: string;
  position_id: number;
};

/**
 * Validacija profila zaposlenog — deli je onboarding (samousluživanje) i
 * HR forma za novog zaposlenog.
 *
 * Obavezno: ime, prezime, država, datum rođenja, datum zaposlenja, pozicija.
 * Opciono, ali se validira ako je uneto: telefon, privatni email, adresa, grad.
 */
export function validateProfile(form: ProfileForm): Errors {
  const errors: Errors = {};

  const first = nameError("Ime", form.first_name);
  if (first) errors.first_name = first;

  const last = nameError("Prezime", form.last_name);
  if (last) errors.last_name = last;

  if (isUnset(form.country)) errors.country = "Država je obavezna.";

  if (isBlank(form.date_of_birth)) errors.date_of_birth = REQUIRED;
  else if (!isOldEnough(form.date_of_birth))
    errors.date_of_birth = `Zaposleni mora imati najmanje ${MIN_AGE} godina.`;

  if (isBlank(form.hire_date)) errors.hire_date = REQUIRED;

  if (isUnset(form.position_id)) errors.position_id = "Pozicija je obavezna.";

  const phone = phoneError(form.phone_number);
  if (phone) errors.phone_number = phone;

  const privateEmail = privateEmailError(form.private_email);
  if (privateEmail) errors.private_email = privateEmail;

  return errors;
}

export type ContactForm = {
  first_name: string;
  last_name: string;
  phone_number: string;
  private_email: string;
};

/**
 * Validacija forme za **izmenu** zaposlenog: ime/prezime + kontakt polja.
 *
 * Namerno ne zahteva državu/datume/poziciju — postojeći zaposleni ih može
 * imati prazne (kolone to dozvoljavaju), pa izmena ne sme da bude blokirana.
 */
export function validateContact(form: ContactForm): Errors {
  const errors: Errors = {};

  const first = nameError("Ime", form.first_name);
  if (first) errors.first_name = first;

  const last = nameError("Prezime", form.last_name);
  if (last) errors.last_name = last;

  const phone = phoneError(form.phone_number);
  if (phone) errors.phone_number = phone;

  const privateEmail = privateEmailError(form.private_email);
  if (privateEmail) errors.private_email = privateEmail;

  return errors;
}

/** Zajednička poruka kada forma nije popunjena. */
export const FORM_INCOMPLETE =
  "Polja moraju da budu popunjena. Označena polja su obavezna.";

/** Poruka za formu profila (onboarding / HR kreiranje naloga). */
export const PROFILE_INCOMPLETE =
  "Molimo vas popunite polja kako bi napravili nalog za vas.";
