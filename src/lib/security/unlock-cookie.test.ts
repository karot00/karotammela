import assert from "node:assert/strict";
import { describe, it } from "node:test";

import { createUnlockCookieValue, verifyUnlockCookieValue } from "@/lib/security/unlock-cookie";

describe("unlock cookie", () => {
  const secret = "test-secret";

  it("verifies a signed payload", () => {
    const value = createUnlockCookieValue(
      {
        sessionId: "abc123",
        locale: "fi",
        unlockedAt: 123,
      },
      secret,
    );

    const parsed = verifyUnlockCookieValue(value, secret);
    assert.deepEqual(parsed, {
      sessionId: "abc123",
      locale: "fi",
      unlockedAt: 123,
    });
  });

  it("rejects tampered payloads", () => {
    const value = createUnlockCookieValue(
      {
        sessionId: "abc123",
        locale: "fi",
        unlockedAt: 123,
      },
      secret,
    );

    const tampered = `${value}x`;
    const parsed = verifyUnlockCookieValue(tampered, secret);
    assert.equal(parsed, null);
  });
});
