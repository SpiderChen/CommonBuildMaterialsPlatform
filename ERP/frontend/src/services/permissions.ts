import type { Role } from "./types";

export function hasPermission(grants: string[] | undefined, permission: string): boolean {
  if (!permission) return true;
  if (permission.includes("&")) {
    return permission.split("&").map((item) => item.trim()).filter(Boolean).every((item) => hasPermission(grants, item));
  }
  if (permission.includes("|")) {
    return permission.split("|").map((item) => item.trim()).filter(Boolean).some((item) => hasPermission(grants, item));
  }
  return (grants || []).some((grant) => {
    if (grant === "*" || grant === permission) return true;
    if (grant.endsWith(":*") && permission.startsWith(grant.slice(0, -1))) return true;
    if (grant === "reports:read" && permission === "report:read") return true;
    return false;
  });
}

export function permissionsForRole(roles: Role[] | undefined, roleCode: string | undefined) {
  return (roles || []).find((role) => role.code === roleCode)?.permissions || [];
}
