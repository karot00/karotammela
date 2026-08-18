const VIP_STATUS_TIMEOUT_MS = 2500;
const DEFAULT_VIP_ORIGIN = "https://karotammela.fi";

// The VIP portal is owned by the Go build on the apex host. The Next.js
// dashboard only links to it; it never renders or proxies VIP UI (plan §4.4).
function vipStatusOrigin(): string {
  const raw = process.env.VIP_STATUS_ORIGIN?.trim();
  if (!raw) {
    return DEFAULT_VIP_ORIGIN;
  }
  try {
    return new URL(raw).origin;
  } catch {
    return DEFAULT_VIP_ORIGIN;
  }
}

// parseVipStatus validates the Go /api/vip/status contract and fails closed:
// anything but an enabled:true response carrying an absolute URL on the
// expected origin yields null (plan §5.2). The client only ever receives the
// resulting vipUrl string, never the raw status or any feature flag.
export function parseVipStatus(
  raw: string,
  allowedOrigin: string,
): string | null {
  let body: unknown;
  try {
    body = JSON.parse(raw);
  } catch {
    return null;
  }
  if (typeof body !== "object" || body === null || Array.isArray(body)) {
    return null;
  }
  const record = body as Record<string, unknown>;
  if (record.enabled !== true) {
    return null;
  }
  if (typeof record.url !== "string" || record.url.length === 0) {
    return null;
  }
  let parsed: URL;
  try {
    parsed = new URL(record.url);
  } catch {
    return null;
  }
  if (parsed.origin !== allowedOrigin) {
    return null;
  }
  return parsed.toString();
}

// getVipUrl resolves the VIP link for the unlocked dashboard. It must never
// throw and must never block the dashboard: timeout, network failure, non-200,
// malformed JSON or an unexpected origin all hide the link (fail closed).
export async function getVipUrl(): Promise<string | null> {
  const origin = vipStatusOrigin();
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), VIP_STATUS_TIMEOUT_MS);
  try {
    const response = await fetch(`${origin}/api/vip/status`, {
      cache: "no-store",
      headers: { accept: "application/json" },
      signal: controller.signal,
    });
    if (!response.ok) {
      return null;
    }
    return parseVipStatus(await response.text(), origin);
  } catch {
    return null;
  } finally {
    clearTimeout(timer);
  }
}
