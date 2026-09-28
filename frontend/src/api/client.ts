import { toaster } from "@/components/ui/toaster";

let authToken: string | null = null;
let refreshPromise: Promise<string | null> | null = null;

/** Poruka kada access token istekne i tihi refresh ne uspe. */
export const SESSION_EXPIRED = "Vaša sesija je istekla. Prijavite se ponovo.";
/** Poruka za greške koje nisu drugačije pokrivene (5xx, mreža). */
export const SERVER_ERROR = "Problem na serveru.";

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
    // Ne prikazujemo sirovi HTML (proxy/render error stranu) kao poruku.
    if (/^\s*</.test(text)) return SERVER_ERROR;
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

  let res: Response;
  try {
    res = await fetch(path, { ...options, headers, credentials: "include" });
  } catch {
    toaster.create({ title: SERVER_ERROR, type: "error" });
    throw new ApiError(SERVER_ERROR, 0);
  }

  if (res.status === 401 && authToken) {
    let fresh: string | null = null;
    try {
      if (!refreshPromise) {
        refreshPromise = fetch("/api/v1/refresh", {
          method: "POST",
          credentials: "include",
        })
          .then(async (r) => {
            // 401/403 → the refresh cookie is gone or revoked: a real session end.
            if (r.status === 401 || r.status === 403) return null;
            // Anything else that is not ok is a server/network problem, not an
            // expired session — do not log the user out for it.
            if (!r.ok) throw new Error(`refresh failed: ${r.status}`);
            const data = await r.json();
            return accessTokenOf(data) ?? null;
          })
          .finally(() => {
            refreshPromise = null;
          });
      }
      fresh = await refreshPromise;
    } catch {
      toaster.create({ title: SERVER_ERROR, type: "error" });
      throw new ApiError(SERVER_ERROR, 0);
    }

    if (!fresh) {
      authToken = null;
      // Notify the app so it clears the session (the route guard then redirects
      // to /login) and tell the user why.
      window.dispatchEvent(new Event("auth:expired"));
      toaster.create({ title: SESSION_EXPIRED, type: "error" });
      throw new Error(SESSION_EXPIRED);
    }

    headers["Authorization"] = `Bearer ${fresh}`;
    try {
      res = await fetch(path, { ...options, headers, credentials: "include" });
    } catch {
      toaster.create({ title: SERVER_ERROR, type: "error" });
      throw new ApiError(SERVER_ERROR, 0);
    }
  }

  if (!res.ok) {
    // 5xx and anything else we cannot explain to the user → generic message.
    if (res.status >= 500) {
      toaster.create({ title: SERVER_ERROR, type: "error" });
      throw new ApiError(SERVER_ERROR, res.status);
    }
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
