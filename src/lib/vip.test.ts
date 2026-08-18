import assert from "node:assert/strict";
import { describe, it } from "node:test";

import { parseVipStatus } from "@/lib/vip";

const ORIGIN = "https://karotammela.fi";

describe("parseVipStatus (fail-closed VIP status contract)", () => {
  it("returns the canonical url for an enabled same-origin status", () => {
    const raw = JSON.stringify({ enabled: true, url: `${ORIGIN}/en/vip` });
    assert.equal(parseVipStatus(raw, ORIGIN), `${ORIGIN}/en/vip`);
  });

  it("returns null when the feature is disabled", () => {
    assert.equal(parseVipStatus(`{"enabled":false}`, ORIGIN), null);
    assert.equal(
      parseVipStatus(`{"enabled":false,"url":"${ORIGIN}/en/vip"}`, ORIGIN),
      null,
    );
  });

  it("returns null for malformed or empty bodies", () => {
    assert.equal(parseVipStatus("not json", ORIGIN), null);
    assert.equal(parseVipStatus("", ORIGIN), null);
    assert.equal(parseVipStatus("{", ORIGIN), null);
  });

  it("returns null when enabled is not exactly true", () => {
    assert.equal(
      parseVipStatus(`{"enabled":"true","url":"${ORIGIN}/en/vip"}`, ORIGIN),
      null,
    );
    assert.equal(
      parseVipStatus(`{"enabled":1,"url":"${ORIGIN}/en/vip"}`, ORIGIN),
      null,
    );
    assert.equal(parseVipStatus(`{"url":"${ORIGIN}/en/vip"}`, ORIGIN), null);
  });

  it("returns null when url is missing, empty or not a string", () => {
    assert.equal(parseVipStatus(`{"enabled":true}`, ORIGIN), null);
    assert.equal(parseVipStatus(`{"enabled":true,"url":""}`, ORIGIN), null);
    assert.equal(parseVipStatus(`{"enabled":true,"url":42}`, ORIGIN), null);
    assert.equal(parseVipStatus(`{"enabled":true,"url":null}`, ORIGIN), null);
  });

  it("returns null for an unparsable url", () => {
    const raw = JSON.stringify({ enabled: true, url: "not a url" });
    assert.equal(parseVipStatus(raw, ORIGIN), null);
  });

  it("returns null for a cross-origin url (allow-list, fail closed)", () => {
    const raw = JSON.stringify({
      enabled: true,
      url: "https://evil.example.com/en/vip",
    });
    assert.equal(parseVipStatus(raw, ORIGIN), null);
  });

  it("returns null for a relative url", () => {
    const raw = JSON.stringify({ enabled: true, url: "/en/vip" });
    assert.equal(parseVipStatus(raw, ORIGIN), null);
  });

  it("returns null for non-object JSON bodies", () => {
    assert.equal(parseVipStatus("null", ORIGIN), null);
    assert.equal(parseVipStatus("[]", ORIGIN), null);
    assert.equal(parseVipStatus(`"enabled"`, ORIGIN), null);
    assert.equal(parseVipStatus("42", ORIGIN), null);
  });
});
