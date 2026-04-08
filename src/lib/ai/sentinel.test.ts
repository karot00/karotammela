import assert from "node:assert/strict";
import { describe, it } from "node:test";

import {
  applyInputAdjustment,
  extractLevelTag,
  getInputLevelAdjustment,
  resolveNextLevel,
  stripLevelTag,
} from "@/lib/ai/sentinel";

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
    assert.equal(resolveNextLevel(10, null), 15);
  });

  it("caps positive delta per turn", () => {
    assert.equal(resolveNextLevel(10, 90), 40);
  });

  it("caps negative delta per turn", () => {
    assert.equal(resolveNextLevel(60, 0), 30);
  });

  it("penalizes insulting input", () => {
    assert.equal(getInputLevelAdjustment("you are stupid"), -18);
  });

  it("rewards convincing technical input", () => {
    assert.equal(
      getInputLevelAdjustment("Because your session cookie flow has a replay risk, let's review threat boundaries."),
      8,
    );
  });

  it("applies input adjustment with clamp", () => {
    assert.equal(applyInputAdjustment(95, "Because this architecture has clear risk boundaries."), 100);
  });
});
