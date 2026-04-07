import { createHmac, timingSafeEqual } from "node:crypto";

type UnlockPayload = {
  sessionId: string;
  locale: string;
  unlockedAt: number;
};

function toBase64Url(input: string) {
  return Buffer.from(input, "utf8").toString("base64url");
}

function signPayload(payload: string, secret: string) {
  return createHmac("sha256", secret).update(payload).digest("base64url");
}

export function createUnlockCookieValue(payload: UnlockPayload, secret: string) {
  const encodedPayload = toBase64Url(JSON.stringify(payload));
  const signature = signPayload(encodedPayload, secret);

  return `${encodedPayload}.${signature}`;
}

export function verifyUnlockCookieValue(value: string, secret: string) {
  const [encodedPayload, signature] = value.split(".");
  if (!encodedPayload || !signature) return null;

  const expected = signPayload(encodedPayload, secret);
  const a = Buffer.from(signature);
  const b = Buffer.from(expected);

  if (a.length !== b.length || !timingSafeEqual(a, b)) {
    return null;
  }

  try {
    const json = Buffer.from(encodedPayload, "base64url").toString("utf8");
    const payload = JSON.parse(json) as UnlockPayload;

    if (
      typeof payload.sessionId !== "string" ||
      typeof payload.locale !== "string" ||
      typeof payload.unlockedAt !== "number"
    ) {
      return null;
    }

    return payload;
  } catch {
    return null;
  }
}
