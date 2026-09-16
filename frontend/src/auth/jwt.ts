import type { User } from "./types";

/**
 * Decode the user information embedded in an access token.
 *
 * The access token is short-lived and its payload only carries display/UI
 * claims (id, email, role). The server remains the source of truth for
 * authorization — this is only used to restore the UI after a silent refresh
 * (the raw `user` object is not persisted anywhere).
 */
export function userFromToken(token: string): User | null {
  try {
    const payload = token.split(".")[1];
    if (!payload) return null;

    // base64url → base64 with padding
    const base64 = payload.replace(/-/g, "+").replace(/_/g, "/");
    const padded = base64 + "=".repeat((4 - (base64.length % 4)) % 4);

    const claims = JSON.parse(atob(padded));
    if (!claims.user_id) return null;

    return {
      id: Number(claims.user_id),
      email: String(claims.email ?? ""),
      role: String(claims.role ?? ""),
    };
  } catch {
    return null;
  }
}
