let authToken: string | null = null;
let refreshPromise: Promise<string> | null = null;

/**
 * Greška sa HTTP statusom, da stranice mogu da razlikuju npr. 409 (konflikt —
 * „email već postoji") od 400/401 i da vežu poruku za konkretno polje.
 */
export class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

export const setToken = (t: string | null) => {
  authToken = t;
};
export const getToken = () => authToken;

async function readError(res: Response): Promise<string> {
  // Middleware vraća običan tekst (npr. "Zabranjen pristup: nedostaje dozvola
  // users.manage"), a handleri JSON. Podržavamo oba formata.
  const text = await res.text();
  if (!text) return `Zahtev nije uspeo (${res.status})`;
  try {
    const data = JSON.parse(text);
    return data.message || data.error || text;
  } catch {
    return text;
  }
}

/** Extract the access token from an enveloped response body. */
const accessTokenOf = (body: any): string | undefined => body?.data?.accessToken;

export async function api(path: string, options: RequestInit = {}) {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };
  if (authToken) {
    headers["Authorization"] = `Bearer ${authToken}`;
  }

  let res = await fetch(path, { ...options, headers, credentials: "include" });

  if (res.status === 401 && authToken) {
    try {
      if (!refreshPromise) {
        refreshPromise = fetch("/api/v1/refresh", {
          method: "POST",
          credentials: "include",
        })
          .then((r) => r.json())
          .then((data) => {
            const fresh = accessTokenOf(data);
            if (!fresh) throw new Error("osvežavanje nije uspelo");
            authToken = fresh;
            return fresh;
          })
          .finally(() => {
            refreshPromise = null;
          });
      }

      const newToken = await refreshPromise;
      headers["Authorization"] = `Bearer ${newToken}`;
      res = await fetch(path, { ...options, headers, credentials: "include" });
    } catch {
      authToken = null;
      throw new Error("Sesija je istekla. Prijavite se ponovo.");
    }
  }

  if (!res.ok) {
    throw new ApiError(await readError(res), res.status);
  }

  const body = await res.json();

  // The API always answers with an envelope: { status, message, data }.
  // Unwrap it so pages work with the payload directly.
  if (body && typeof body === "object" && "status" in body && "data" in body) {
    return body.data;
  }
  return body;
}
