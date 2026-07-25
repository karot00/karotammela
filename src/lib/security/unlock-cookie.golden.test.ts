/**
 * Phase 12.5f golden-vector suite: consumes the shared vector file generated
 * by scripts/generate-unlock-cookie-vectors.ts and goth/test/genvectors, and
 * proves the reference implementation reproduces every create-parity value
 * byte-for-byte (including the Go-minted cookies) and accepts/rejects every
 * value per the shared expectation. The Go twin suite lives in
 * goth/internal/security/unlock_cookie_golden_test.go.
 */
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { join } from "node:path";
import { describe, it } from "node:test";

import { createUnlockCookieValue, verifyUnlockCookieValue } from "@/lib/security/unlock-cookie";

type GoldenVector = {
  name: string;
  createdBy: "node" | "go" | "manual";
  secret: string;
  verifySecret?: string;
  payload: { sessionId: string; locale: string; unlockedAt: number } | null;
  value: string;
  expectVerify: boolean;
  note?: string;
};

const doc = JSON.parse(
  readFileSync(join(process.cwd(), "shared/security/unlock-cookie-vectors.json"), "utf8"),
) as { vectors: GoldenVector[] };

describe("unlock cookie golden vectors (12.5f)", () => {
  assert.ok(doc.vectors.length > 0, "no vectors loaded");
  const creators = new Set(doc.vectors.map((v) => v.createdBy));
  assert.ok(creators.has("node") && creators.has("go"), "vectors must cover both stacks");

  for (const vector of doc.vectors) {
    it(`verifies ${vector.name} per expectation`, () => {
      const secret = vector.verifySecret ?? vector.secret;
      const parsed = verifyUnlockCookieValue(vector.value, secret);
      if (!vector.expectVerify) {
        assert.equal(parsed, null, `expected rejection (${vector.note ?? ""})`);
        return;
      }
      assert.ok(vector.payload, "accepting vector must carry its expected payload");
      assert.deepEqual(parsed, vector.payload);
    });

    // Only stack-minted vectors are create-parity cases; "manual" vectors
    // (e.g. extra trailing segments) are verify-only by construction.
    if (vector.payload && (vector.createdBy === "node" || vector.createdBy === "go")) {
      it(`reproduces ${vector.name} byte-for-byte (created by ${vector.createdBy})`, () => {
        assert.equal(createUnlockCookieValue(vector.payload!, vector.secret), vector.value);
      });
    }
  }
});
