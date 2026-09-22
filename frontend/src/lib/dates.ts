/**
 * Pomoćne funkcije za datume na klijentu (bez zavisnosti).
 *
 * Napomena: vikendi se računaju lokalno, a praznici ne (klijent ih ne dobija sa
 * servera). Zato klijentski broj radnih dana može biti **veći ili jednak**
 * serverskom, pa je provera „da li je previše dana" konzervativna — server je i
 * dalje jedini autoritet (vraća `409 Insufficient balance`).
 */

export const ISO_DATE = /^\d{4}-\d{2}-\d{2}$/;

/** Datum u `YYYY-MM-DD` formatu (lokalno). */
export function formatISO(d: Date): string {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}

/** Datum parsiran kao lokalna ponoć (bez UTC pomeranja). */
export function parseISO(iso: string): Date {
  const [y, m, d] = iso.split("-").map(Number);
  return new Date(y, m - 1, d);
}

/** Današnji datum u `YYYY-MM-DD` formatu (lokalno). */
export function todayISO(now: Date = new Date()): string {
  return formatISO(new Date(now.getFullYear(), now.getMonth(), now.getDate()));
}

export function isWeekend(iso: string): boolean {
  if (!ISO_DATE.test(iso)) return false;
  const day = parseISO(iso).getDay();
  return day === 0 || day === 6;
}

/**
 * Srpska množina za dane: 1 radni dan, 2 radna dana, 5 radnih dana.
 * Vraća samo deo sa imenicom („radni dan“ / „radna dana“ / „radnih dana“).
 */
export function daysLabel(n: number): string {
  const abs = Math.abs(Math.trunc(n)) % 100;
  const last = abs % 10;
  if (abs >= 11 && abs <= 14) return "radnih dana";
  if (last === 1) return "radni dan";
  if (last >= 2 && last <= 4) return "radna dana";
  return "radnih dana";
}

/** Broj radnih dana (bez vikenda) u inkluzivnom periodu [start, end]. */
export function workingDays(start: string, end: string): number {
  if (!ISO_DATE.test(start) || !ISO_DATE.test(end)) return 0;
  const s = parseISO(start);
  const e = parseISO(end);
  if (e < s) return 0;

  let count = 0;
  const cursor = new Date(s);
  while (cursor <= e) {
    const day = cursor.getDay();
    if (day !== 0 && day !== 6) count++;
    cursor.setDate(cursor.getDate() + 1);
  }
  return count;
}

/**
 * Poslednji datum (počev od `start`) koji sadrži najviše `maxDays` radnih dana —
 * koristi se kao `max` za polje „Do", tako da se ne može izabrati predugačak
 * period.
 */
export function lastDateWithin(start: string, maxDays: number): string {
  if (!ISO_DATE.test(start) || maxDays < 1) return start;

  const cursor = parseISO(start);
  let counted = 0;
  for (;;) {
    const day = cursor.getDay();
    if (day !== 0 && day !== 6) {
      counted++;
      if (counted === maxDays) break;
    }
    cursor.setDate(cursor.getDate() + 1);
  }
  return formatISO(cursor);
}
