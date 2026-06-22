let authToken: string | null = null;
let refreshPromise: Promise<string> | null = null;

export const setToken = (t: string | null) => {
  authToken = t;
};
export const getToken = () => authToken;

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
    if (!refreshPromise) {
      refreshPromise = fetch("/api/v1/refresh", {
        method: "POST",
        credentials: "include",
      })
        .then((r) => r.json())
        .then((data) => {
          authToken = data.accessToken;
          return data.accessToken;
        })
        .finally(() => {
          refreshPromise = null;
        });
    }

    const newToken = await refreshPromise;
    headers["Authorization"] = `Bearer ${newToken}`;
    res = await fetch(path, { ...options, headers, credentials: "include" });
  }

  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.message || "Request failed");
  }

  return res.json();
}
