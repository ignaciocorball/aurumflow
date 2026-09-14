import { describe, expect, it, beforeEach } from "vitest";
import { initialState, persistTheme, reduce, sortMarkets } from "../store/reducer";
import { THEME_KEY } from "../store/types";
import type { Market, Snapshot } from "../store/types";

beforeEach(() => {
  localStorage.clear();
});

describe("theme persistence", () => {
  it("stores preference", () => {
    persistTheme("dark");
    expect(localStorage.getItem(THEME_KEY)).toBe("dark");
    const s = reduce(initialState(), { type: "THEME", theme: "light" });
    expect(s.theme).toBe("light");
    expect(localStorage.getItem(THEME_KEY)).toBe("light");
  });
});

describe("snapshot hydration and SSE reducer", () => {
  it("hydrates and marks live", () => {
    const snap = baseSnap();
    const s = reduce(initialState(), { type: "HYDRATE", snapshot: snap });
    expect(s.snapshot?.portfolio.equity).toBe(999.96);
    expect(s.stale).toBe(false);
    expect(s.connected).toBe(true);
  });
  it("marks stale then recovers", () => {
    let s = reduce(initialState(), { type: "HYDRATE", snapshot: baseSnap() });
    s = reduce(s, { type: "STALE" });
    expect(s.stale).toBe(true);
    s = reduce(s, { type: "HYDRATE", snapshot: baseSnap() });
    expect(s.stale).toBe(false);
  });
  it("applies health and dislocation events", () => {
    let s = reduce(initialState(), { type: "HYDRATE", snapshot: baseSnap() });
    s = reduce(s, { type: "EVENT", event: { type: "DATA_HEALTH_CHANGED", at: "", data: "DEGRADED" } });
    expect(s.snapshot?.environment.data_health).toBe("DEGRADED");
    s = reduce(s, { type: "EVENT", event: { type: "DISLOCATION_CHANGED", at: "", data: "ELEVATED" } });
    expect(s.snapshot?.world.dislocation).toBe("ELEVATED");
  });
});

describe("DEMO vs SHADOW", () => {
  it("keeps kinds distinct", () => {
    const snap = baseSnap();
    expect(snap.execution.positions[0].kind).toBe("DEMO");
    expect(snap.execution.positions[1].kind).toBe("SHADOW");
    expect(snap.portfolio.scope).toBe("BROKER_DEMO");
  });
});

describe("market sorting", () => {
  it("sorts by salience then attention", () => {
    const rows = sortMarkets(baseSnap().markets, "salience", "desc");
    expect(rows[0].id).toBe("J225");
    const att = sortMarkets(baseSnap().markets, "attention", "desc");
    expect(att[0].id).toBe("GOLD");
  });
});

describe("pnl calendar aggregation", () => {
  it("does not invent days", () => {
    expect(baseSnap().portfolio.has_broker_trades).toBe(false);
    expect(baseSnap().portfolio.calendar || []).toEqual([]);
  });
});

function m(partial: Partial<Market> & Pick<Market, "id" | "label">): Market {
  return {
    region: "",
    region_label: "",
    asset_class: "",
    bid: 0,
    ask: 0,
    mid: 0,
    change_pct: 0,
    spark: [],
    attention: 0,
    salience: 0,
    coverage: 0,
    trend: "",
    trend_display: "",
    regime: "",
    setup: "",
    setup_display: "",
    session: "",
    session_display: "",
    research: "",
    research_display: "",
    demo_status: "",
    demo_display: "",
    quality: "",
    quality_display: "",
    market_status: "",
    market_status_display: "",
    eligibility: "",
    positioning: "",
    flow_context: "",
    salience_reason: "",
    evidence: [],
    pilot: false,
    ...partial,
  };
}

function baseSnap(): Snapshot {
  return {
    generated_at: "2026-09-14T03:00:00Z",
    fixture_mode: false,
    environment: {
      name: "DEMO",
      demo: true,
      live_possible: "IMPOSSIBLE",
      live_possible_display: "Impossible",
      account_masked: "****",
      broker_status: "DEMO READY",
      world_valid: "CURRENT_WORLD_STATE_VALID",
      world_valid_display: "World current",
      data_health: "HEALTHY",
      utc: "03:00:00 UTC",
    },
    world: {
      as_of: "",
      valid: "CURRENT_WORLD_STATE_VALID",
      valid_display: "World current",
      hash: "x",
      session: "ASIA",
      regime: "MIXED",
      regime_display: "Mixed",
      dislocation: "NORMAL",
      dislocation_display: "Normal",
      confidence: 0.8,
      liquidity: "Unknown",
      usd: "Unknown",
    },
    markets: [
      m({ id: "J225", label: "J225", salience: 83, attention: 45 }),
      m({ id: "GOLD", label: "GOLD", salience: 36, attention: 60 }),
      m({ id: "US100", label: "US100", salience: 71, attention: 45, demo_display: "DEMO eligible" }),
    ],
    portfolio: {
      equity: 999.96,
      currency: "USD",
      today_pnl: 0,
      today_r: 0,
      unrealized: 0,
      realized: 0,
      net: 0,
      return_pct: 0,
      drawdown_pct: 0,
      open_risk: 0,
      open_risk_cap: 10,
      pilot_cap: 5,
      aggregate_cap: 10,
      precious_risk: 0,
      precious_cap: 5,
      us_equity_risk: 0,
      us_equity_cap: 5,
      scope: "BROKER_DEMO",
      pilots_armed: 2,
      calendar: [],
      weekly_total: 0,
      monthly_total: 0,
      all_time: 0,
      all_time_r: 0,
      trade_count: 0,
      equity_curve: [],
      has_broker_trades: false,
    },
    execution: {
      banner: "CAPITAL DEMO",
      positions: [
        {
          market: "GOLD",
          origin: "GOLD_STRATEGY",
          origin_display: "GOLD strategy",
          kind: "DEMO",
          strategy: "legacy",
          direction: "",
          entry: 0,
          current: 0,
          sl: 0,
          tp: 0,
          upl: 0,
          return_r: 0,
          planned_risk: 0,
          actual_risk: 0,
          account: "****",
          open: false,
          lifecycle: [],
          lifecycle_label: "Idle",
          broker_reconcile: "Matched",
          session: "ASIA",
          waiting: "STRATEGY WAITING",
          last_execution: "",
          deal_ref: "",
          size: 0,
          eligible: "PARTIAL",
          armed: "ON",
        },
        {
          market: "US100",
          origin: "DEMO_MIRROR",
          origin_display: "DEMO mirror",
          kind: "SHADOW",
          strategy: "LEGACY_NORMALIZED_V0",
          direction: "",
          entry: 0,
          current: 0,
          sl: 0,
          tp: 0,
          upl: 0,
          return_r: 0,
          planned_risk: 0,
          actual_risk: 0,
          account: "****",
          open: false,
          lifecycle: [],
          lifecycle_label: "Idle",
          broker_reconcile: "Matched",
          session: "ASIA",
          waiting: "",
          last_execution: "",
          deal_ref: "",
          size: 0,
          eligible: "YES",
          armed: "ON",
        },
      ],
      risk_groups: [],
      halt: false,
      trust: [],
    },
    research: { rows: [] },
    news: [],
    events: [],
    health: [],
    narratives: { world: "", matters: [], execution: "", health: "" },
    activity: [],
    micro: {
      price: 0,
      microprice: 0,
      spread: 0,
      cvd: 0,
      pressure: 0,
      imbalance: 0,
      imb1: 0,
      imb5: 0,
      imb10: 0,
      book_synced: true,
      book_age_ms: 0,
      gaps: 0,
      resyncs: 0,
      drops: 0,
      latency_p50: 0,
      latency_p95: 0,
      latency_p99: 0,
      provider: "",
      capability: "",
      absorption: "",
      exhaustion: "",
      integrity: "",
      event_rate: 0,
      series: [],
      bids: [],
      asks: [],
      resync_reason: "",
    },
    regions: [],
    flows: [],
  };
}
