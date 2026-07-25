/**
 * Phase 12.5f generator: shared unlock-cookie golden vectors for the
 * Node <-> Go interoperability proof. Uses the REAL reference implementation
 * (src/lib/security/unlock-cookie.ts) so the vectors are authoritative for
 * the Node side; the Go side adds its own created-by-Go vectors via
 * goth/test/genvectors and both stacks' test suites consume this file.
 *
 * Run: node --import tsx scripts/generate-unlock-cookie-vectors.ts
 * Output: shared/security/unlock-cookie-vectors.json
 */
import { createHmac } from "node:crypto";
import { writeFileSync, mkdirSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

import { createUnlockCookieValue } from "../src/lib/security/unlock-cookie";

type Payload = { sessionId: string; locale: string; unlockedAt: number };

type Vector = {
  name: string;
  createdBy: "node" | "go" | "manual";
  secret: string;
  verifySecret?: string;
  payload: Payload | null;
  value: string;
  expectVerify: boolean;
  note?: string;
};

const root = join(dirname(fileURLToPath(import.meta.url)), "..");
const outPath = join(root, "shared", "security", "unlock-cookie-vectors.json");

const secret = "golden-vector-secret-0123456789abcdef";

/** Sign an arbitrary JSON string like the reference does (for negative cases). */
function signRawJson(json: string, sec: string): string {
  const encoded = Buffer.from(json, "utf8").toString("base64url");
  const sig = createHmac("sha256", sec).update(encoded).digest("base64url");
  return `${encoded}.${sig}`;
}

const basic: Payload = { sessionId: "sess-abc123", locale: "fi", unlockedAt: 1780000000000 };
const unicode: Payload = {
  sessionId: "sess-<ä>&ö-🚀-\"quote\"",
  locale: "en",
  unlockedAt: 1780000123456,
};
const emptyFields: Payload = { sessionId: "", locale: "", unlockedAt: 0 };

const nodeBasic = createUnlockCookieValue(basic, secret);
const nodeUnicode = createUnlockCookieValue(unicode, secret);
const nodeEmpty = createUnlockCookieValue(emptyFields, secret);

// Tampered signature: flip the last character of a valid cookie.
const flip = (v: string) =>
  v.slice(0, -1) + (v.endsWith("A") ? "B" : "A");

const vectors: Vector[] = [
  {
    name: "node-basic",
    createdBy: "node",
    secret,
    payload: basic,
    value: nodeBasic,
    expectVerify: true,
    note: "create-parity: the other stack must reproduce this value byte-for-byte",
  },
  {
    name: "node-unicode",
    createdBy: "node",
    secret,
    payload: unicode,
    value: nodeUnicode,
    expectVerify: true,
    note: "multibyte + HTML-escape chars (<, &, quotes): JSON.stringify emits them raw; Go must too (SetEscapeHTML(false))",
  },
  {
    name: "node-empty-fields",
    createdBy: "node",
    secret,
    payload: emptyFields,
    value: nodeEmpty,
    expectVerify: true,
    note: "12.5f explicit resolution: empty strings and unlockedAt=0 PASS the reference typeof checks; both stacks accept",
  },
  {
    name: "signed-wrong-type-sessionid",
    createdBy: "manual",
    secret,
    payload: null,
    value: signRawJson(JSON.stringify({ sessionId: 123, locale: "fi", unlockedAt: 1780000000000 }), secret),
    expectVerify: false,
    note: "valid HMAC but sessionId is a number: rejected by typeof check (Node) / unmarshal type error (Go)",
  },
  {
    name: "signed-missing-locale",
    createdBy: "manual",
    secret,
    payload: null,
    value: signRawJson(JSON.stringify({ sessionId: "sess-x", unlockedAt: 1780000000000 }), secret),
    expectVerify: false,
    note: "valid HMAC but locale key absent: rejected by both stacks",
  },
  {
    name: "signed-null-fields",
    createdBy: "manual",
    secret,
    payload: null,
    value: signRawJson(JSON.stringify({ sessionId: null, locale: null, unlockedAt: null }), secret),
    expectVerify: false,
    note: "null is not string/number: rejected by both stacks",
  },
  {
    name: "tampered-signature",
    createdBy: "manual",
    secret,
    payload: null,
    value: flip(nodeBasic),
    expectVerify: false,
    note: "last signature char flipped: HMAC comparison fails",
  },
  {
    name: "tampered-payload",
    createdBy: "manual",
    secret,
    payload: null,
    value: `${nodeUnicode.split(".")[0]}.${nodeBasic.split(".")[1]}`,
    expectVerify: false,
    note: "payload swapped under another cookie's signature: HMAC mismatch",
  },
  {
    name: "wrong-secret",
    createdBy: "manual",
    secret,
    verifySecret: "not-the-right-secret",
    payload: null,
    value: nodeBasic,
    expectVerify: false,
    note: "valid cookie, wrong verification secret",
  },
  {
    name: "malformed-no-dot",
    createdBy: "manual",
    secret,
    payload: null,
    value: "bm90LWEtY29va2ll",
    expectVerify: false,
    note: "no separator",
  },
  {
    name: "malformed-empty-payload",
    createdBy: "manual",
    secret,
    payload: null,
    value: ".c2ln",
    expectVerify: false,
    note: "empty payload part",
  },
  {
    name: "malformed-empty-signature",
    createdBy: "manual",
    secret,
    payload: null,
    value: "cGF5bG9hZA.",
    expectVerify: false,
    note: "empty signature part",
  },
  {
    name: "extra-segments-ignored",
    createdBy: "manual",
    secret,
    payload: basic,
    value: `${nodeBasic}.extra`,
    expectVerify: true,
    note: "the reference split('.') keeps only segments 0 and 1, so trailing segments are ignored; both stacks accept (aligned in 12.5f)",
  },
];

const doc = {
  description:
    "Phase 12.5f golden vectors proving karot_unlock cookie interoperability between the Next.js reference " +
    "(src/lib/security/unlock-cookie.ts) and the Go port (goth/internal/security/security.go). " +
    "createdBy records which stack minted the value; both stacks must (a) reproduce every create-parity value " +
    "byte-for-byte from its payload and (b) accept/reject every value per expectVerify. " +
    "Created-by-Go vectors are appended by goth/test/genvectors, not by this script.",
  spec: {
    valueFormat: "base64url(JSON payload, no padding) + '.' + base64url(HMAC-SHA256(secret, payload-part), no padding)",
    payloadKeyOrder: ["sessionId", "locale", "unlockedAt"],
    validation:
      "HMAC compared first (constant time, length-checked); then each field must be present with the right JSON " +
      "type (string, string, number). Empty strings and unlockedAt=0 are ACCEPTED (reference typeof semantics, " +
      "explicitly resolved in 12.5f). Missing keys, null, and wrong JSON types are rejected.",
    cookieAttributes: {
      name: "karot_unlock",
      path: "/",
      httpOnly: true,
      sameSite: "Lax",
      secure: "production only",
      maxAgeSeconds: 1209600,
    },
  },
  vectors,
};

mkdirSync(dirname(outPath), { recursive: true });
writeFileSync(outPath, JSON.stringify(doc, null, 2) + "\n");
console.log(`wrote ${vectors.length} vectors to ${outPath}`);
