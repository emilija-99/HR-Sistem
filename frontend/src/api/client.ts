let authToken: string | null = null;
let refreshPromise: Promise<string> | null = null;

export const setToken = (t: string | null) => {
  authToken = t;
};
export const getToken = () => authToken;

async function readError(res: Response): Promise<string> {
  // Permission middleware replies with plain text (e.g. "Forbidden: missing
  // permission users.manage"); handlers reply with JSON. Handle both.
  const text = await res.text();
  if (!text) return `Request failed (${res.status})`;
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
            if (!fresh) throw new Error("refresh failed");
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
      throw new Error("Session expired. Please log in again.");
    }
  }

  if (!res.ok) {
    throw new Error(await readError(res));
  }

  const body = await res.json();

  // The API always answers with an envelope: { status, message, data }.
  // Unwrap it so pages work with the payload directly.
  if (body && typeof body === "object" && "status" in body && "data" in body) {
    return body.data;
  }
  return body;
}
