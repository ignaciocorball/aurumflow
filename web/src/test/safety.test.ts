import { describe, expect, it } from "vitest";
import { hydrate } from "../store/store";
import { Forbidden } from "./forbidden";

describe("no execution mutation path", () => {
  it("snapshot client is GET-only", () => {
    expect(hydrate.toString()).not.toMatch(/POST/);
    expect(hydrate.toString()).not.toMatch(/BUY/);
    expect(hydrate.toString()).not.toMatch(/SELL/);
  });
  it("UI copy has no order verbs as controls", () => {
    expect(Forbidden).not.toContain("BUY_BUTTON");
  });
});
