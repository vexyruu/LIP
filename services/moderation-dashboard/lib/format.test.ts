import { describe, it, expect } from "vitest";
import {
  shortUUID,
  formatPrice,
  tierBadgeClass,
  listingShortId,
  conditionLabel,
  riskTierLabel,
  policyViolationLabel,
  priceMarkerPercent,
} from "./format";

describe("shortUUID", () => {
  it("abbreviates a full UUID", () => {
    expect(shortUUID("22222222-2222-2222-2222-222222222222")).toBe("2222…2222");
  });

  it("returns short strings unchanged", () => {
    expect(shortUUID("abc")).toBe("abc");
  });
});

describe("formatPrice", () => {
  it("formats USD with two decimals and grouping", () => {
    expect(formatPrice(1234.5)).toBe("$1,234.50");
    expect(formatPrice(0)).toBe("$0.00");
  });
});

describe("conditionLabel", () => {
  it("maps known conditions", () => {
    expect(conditionLabel(1)).toBe("1/5 (New)");
    expect(conditionLabel(5)).toBe("5/5 (Poor)");
  });

  it("falls back to Unknown for out-of-range values", () => {
    expect(conditionLabel(9)).toBe("9/5 (Unknown)");
  });
});

describe("riskTierLabel", () => {
  it("labels each tier", () => {
    expect(riskTierLabel("HIGH")).toBe("CRITICAL RISK TIER");
    expect(riskTierLabel("MEDIUM")).toBe("ELEVATED RISK TIER");
    expect(riskTierLabel("LOW")).toBe("LOW RISK TIER");
    expect(riskTierLabel(null)).toBe("UNSCORED RISK TIER");
  });
});

describe("policyViolationLabel", () => {
  it("maps booleans and null", () => {
    expect(policyViolationLabel(true)).toBe("DETECTED");
    expect(policyViolationLabel(false)).toBe("NONE DETECTED");
    expect(policyViolationLabel(null)).toBe("UNKNOWN");
  });
});

describe("priceMarkerPercent", () => {
  it("returns 50 when bounds are missing or invalid", () => {
    expect(priceMarkerPercent(10, null, null)).toBe(50);
    expect(priceMarkerPercent(10, 100, 100)).toBe(50);
  });

  it("positions the ask within the band", () => {
    expect(priceMarkerPercent(50, 0, 100)).toBe(50);
    expect(priceMarkerPercent(25, 0, 100)).toBe(25);
  });

  it("clamps to 0–100", () => {
    expect(priceMarkerPercent(-50, 0, 100)).toBe(0);
    expect(priceMarkerPercent(500, 0, 100)).toBe(100);
  });
});

describe("listingShortId", () => {
  it("formats the first 8 hex chars as XXXX-XXXX", () => {
    expect(listingShortId("abcdef12-3456-7890-0000-000000000000")).toBe(
      "ABCD-EF12"
    );
  });
});

describe("tierBadgeClass", () => {
  it("uses the error color for HIGH", () => {
    expect(tierBadgeClass("HIGH")).toContain("text-error");
  });

  it("is case-insensitive", () => {
    expect(tierBadgeClass("low")).toContain("text-primary-container");
  });
});
