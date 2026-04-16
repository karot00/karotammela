const FALLBACK_SITE_URL = "https://karotammela.fi";

function normalizeUrl(input: string) {
  return input.endsWith("/") ? input.slice(0, -1) : input;
}

export function getSiteUrl() {
  const envUrl = process.env.APP_URL?.trim();

  if (!envUrl) {
    return FALLBACK_SITE_URL;
  }

  try {
    const parsed = new URL(envUrl);
    return normalizeUrl(parsed.toString());
  } catch {
    return FALLBACK_SITE_URL;
  }
}

export function toAbsoluteUrl(pathname: string) {
  const siteUrl = getSiteUrl();
  const safePath = pathname.startsWith("/") ? pathname : `/${pathname}`;
  return `${siteUrl}${safePath}`;
}
