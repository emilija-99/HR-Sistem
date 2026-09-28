/**
 * Ko sme biti nečiji nadređeni: SENIOR ili LEAD po poziciji, ili
 * PLATFORM_ADMIN / MANAGER_PORTAL_ACCESS po ulozi.
 *
 * Isto pravilo se primenjuje i na serveru (`employee.IsEligibleSupervisor`), pa
 * je ovo samo UX filter za padajuću listu — server i dalje sam odbija nevažeći
 * izbor.
 */

export const SUPERVISOR_LEVELS = ["SENIOR", "LEAD"];
export const SUPERVISOR_ROLES = ["PLATFORM_ADMIN", "MANAGER_PORTAL_ACCESS"];

type Candidate = {
  first_name: string;
  last_name: string;
  position_level?: string | null;
  role?: string | null;
};

export function isEligibleSupervisor(e: {
  position_level?: string | null;
  role?: string | null;
}): boolean {
  return (
    SUPERVISOR_LEVELS.includes(e.position_level ?? "") ||
    SUPERVISOR_ROLES.includes(e.role ?? "")
  );
}

/** Prikaz u listi: ime + nivo, odnosno uloga kada je ona razlog podobnosti. */
export const supervisorLabel = (e: Candidate): string => {
  const extra = SUPERVISOR_ROLES.includes(e.role ?? "")
    ? e.role
    : (e.position_level ?? "");
  return extra ? `${e.first_name} ${e.last_name} (${extra})` : `${e.first_name} ${e.last_name}`;
};
