import { runMigrationsIfNeeded } from "@/lib/db/migrate";
import { getDashboardStats } from "@/lib/db/queries";

type DashboardStats = Awaited<ReturnType<typeof getDashboardStats>>;

type CacheEntry = {
  expiresAt: number;
  data: DashboardStats;
};

const CACHE_TTL_MS = 15_000;
const cache = new Map<string, CacheEntry>();

export async function getCachedDashboardStats(locale: "en" | "fi") {
  const now = Date.now();
  const key = `stats:${locale}`;
  const hit = cache.get(key);

  if (hit && hit.expiresAt > now) {
    return hit.data;
  }

  await runMigrationsIfNeeded();
  const fresh = await getDashboardStats(locale);

  cache.set(key, {
    expiresAt: now + CACHE_TTL_MS,
    data: fresh,
  });

  return fresh;
}
