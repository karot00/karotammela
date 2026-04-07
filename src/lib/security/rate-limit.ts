type RateLimitConfig = {
  scope: string;
  key: string;
  limit: number;
  windowMs: number;
};

type RateLimitEntry = {
  count: number;
  resetAt: number;
};

const STORE_KEY = "__karot_rate_limit_store__";

function getStore() {
  const globalState = globalThis as typeof globalThis & {
    [STORE_KEY]?: Map<string, RateLimitEntry>;
  };

  if (!globalState[STORE_KEY]) {
    globalState[STORE_KEY] = new Map<string, RateLimitEntry>();
  }

  return globalState[STORE_KEY];
}

function cleanupExpiredEntries(store: Map<string, RateLimitEntry>, now: number) {
  for (const [key, entry] of store.entries()) {
    if (entry.resetAt <= now) {
      store.delete(key);
    }
  }
}

export function getClientIp(request: Request) {
  const forwardedFor = request.headers.get("x-forwarded-for");
  if (forwardedFor) {
    return forwardedFor.split(",")[0]?.trim() || "unknown";
  }

  const realIp = request.headers.get("x-real-ip");
  if (realIp) {
    return realIp;
  }

  return "unknown";
}

export function enforceRateLimit(config: RateLimitConfig) {
  const now = Date.now();
  const store = getStore();
  const bucketKey = `${config.scope}:${config.key}`;

  if (store.size > 5000) {
    cleanupExpiredEntries(store, now);
  }

  const existing = store.get(bucketKey);

  if (!existing || existing.resetAt <= now) {
    store.set(bucketKey, {
      count: 1,
      resetAt: now + config.windowMs,
    });

    return {
      allowed: true,
      remaining: Math.max(config.limit - 1, 0),
      retryAfterMs: config.windowMs,
    };
  }

  existing.count += 1;

  if (existing.count > config.limit) {
    return {
      allowed: false,
      remaining: 0,
      retryAfterMs: Math.max(existing.resetAt - now, 1000),
    };
  }

  return {
    allowed: true,
    remaining: Math.max(config.limit - existing.count, 0),
    retryAfterMs: Math.max(existing.resetAt - now, 1000),
  };
}
