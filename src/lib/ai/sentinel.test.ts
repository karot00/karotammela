import assert from "node:assert/strict";
import { describe, it } from "node:test";

import { extractLevelTag, resolveNextLevel, stripLevelTag } from "@/lib/ai/sentinel";

describe("sentinel parser", () => {
  it("extracts a valid level tag", () => {
    assert.equal(extractLevelTag("Hello [LEVEL:42]"), 42);
  });

  it("clamps extracted values beyond range", () => {
    assert.equal(extractLevelTag("[LEVEL:999]"), 100);
  });

  it("strips level tags from visible text", () => {
    assert.equal(stripLevelTag("Response [LEVEL:66]"), "Response");
  });

  it("applies fallback progression when tag is missing", () => {
    assert.equal(resolveNextLevel(10, null), 12);
  });

  it("caps positive delta per turn", () => {
    assert.equal(resolveNextLevel(10, 90), 28);
  });

  it("caps negative delta per turn", () => {
    assert.equal(resolveNextLevel(60, 0), 42);
  });
});
