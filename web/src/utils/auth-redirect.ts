import { clearAccessToken } from "@/auth-state";
import { ROUTES } from "@/router/routes";

const AUTH_ROUTES = [ROUTES.AUTH, "/auth/sign-in", "/auth/sign-up", "/auth/admin"];

function isAuthRoute(path: string): boolean {
  return AUTH_ROUTES.some((route) => path.startsWith(route));
}

export function redirectOnAuthFailure(): void {
  const currentPath = window.location.pathname;

  // Don't redirect if already on auth pages (avoid redirect loop)
  if (isAuthRoute(currentPath)) {
    return;
  }

  // Clear invalid token and redirect to login
  clearAccessToken();
  window.location.replace(ROUTES.AUTH);
}
