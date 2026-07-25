/**
 * Shared constants for the Go vs Next.js comparison experiment.
 *
 * The apex `karotammela.fi` serves the self-hosted Go + HTMX build (Hetzner
 * CX23); the `next.karotammela.fi` subdomain serves this Next.js build natively
 * on Vercel. The Tech Switcher redirects between the two origins and the
 * performance widget pings each `/api/ping` directly from the browser.
 *
 * Both are overridable via NEXT_PUBLIC_* env vars for local development.
 */
export const GO_ORIGIN = (
  process.env.NEXT_PUBLIC_GO_ORIGIN ?? "https://karotammela.fi"
).replace(/\/$/, "");

export const NEXT_ORIGIN = (
  process.env.NEXT_PUBLIC_NEXT_ORIGIN ?? "https://next.karotammela.fi"
).replace(/\/$/, "");

/** Origins allowed to read the comparison endpoints cross-origin. */
export const COMPARISON_ORIGINS = [GO_ORIGIN, NEXT_ORIGIN] as const;

/**
 * Returns the CORS headers for a comparison endpoint. The request Origin is
 * reflected only when it matches a known comparison origin; unknown origins
 * receive no `Access-Control-Allow-Origin` header.
 */
export function comparisonCorsHeaders(origin: string | null): HeadersInit {
  const headers: Record<string, string> = { Vary: "Origin" };

  if (origin && (COMPARISON_ORIGINS as readonly string[]).includes(origin)) {
    headers["Access-Control-Allow-Origin"] = origin;
    headers["Access-Control-Allow-Methods"] = "GET, OPTIONS";
    headers["Access-Control-Max-Age"] = "86400";
  }

  return headers;
}
