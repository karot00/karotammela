const FALLBACK_SITE_URL = "https://www.karotammela.fi";

function normalizeUrl(input: string) {
  return input.endsWith("/") ? input.slice(0, -1) : input;
}

export function getSiteUrl() {
  const envUrl = process.env.APP_URL?.trim();

  if (envUrl) {
    try {
      const parsed = new URL(envUrl);
      const isLocalhost =
        parsed.hostname === "localhost" || parsed.hostname === "127.0.0.1";

      if (!(process.env.NODE_ENV === "production" && isLocalhost)) {
        return normalizeUrl(parsed.toString());
      }
    } catch {
      // Ignore invalid APP_URL and continue to platform-derived defaults.
    }
  }

  const vercelProductionDomain =
    process.env.VERCEL_PROJECT_PRODUCTION_URL?.trim() ||
    process.env.NEXT_PUBLIC_VERCEL_PROJECT_PRODUCTION_URL?.trim();

  if (vercelProductionDomain) {
    return normalizeUrl(`https://${vercelProductionDomain}`);
  }

  const vercelDomain = process.env.VERCEL_URL?.trim();

  if (vercelDomain) {
    return normalizeUrl(`https://${vercelDomain}`);
  }

  if (envUrl) {
    try {
      const parsed = new URL(envUrl);
      return normalizeUrl(parsed.toString());
    } catch {
      // Ignore invalid APP_URL and continue to fallback.
    }
  }

  return FALLBACK_SITE_URL;
}

export function toAbsoluteUrl(pathname: string) {
  const siteUrl = getSiteUrl();
  const safePath = pathname.startsWith("/") ? pathname : `/${pathname}`;
  return `${siteUrl}${safePath}`;
}
