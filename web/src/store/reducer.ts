import type { Action, StreamEvent, UIState, ViewId } from "./types";
import { THEME_KEY } from "./types";

export function loadTheme(): UIState["theme"] {
  const v = localStorage.getItem(THEME_KEY);
  if (v === "light" || v === "dark" || v === "system") return v;
  return "system";
}

export function persistTheme(theme: UIState["theme"]) {
  localStorage.setItem(THEME_KEY, theme);
}

export const initialState = (view: ViewId = "overview"): UIState => ({
  snapshot: null,
  connected: false,
  stale: false,
  lastOk: "",
  view,
  theme: typeof localStorage === "undefined" ? "system" : loadTheme(),
  drawer: null,
  period: "month",
  sortKey: "salience",
  sortDir: "desc",
  error: "",
});

const IMMEDIATE = new Set(["POSITION_UPDATED", "TRADE_COMPLETED", "DATA_HEALTH_CHANGED", "DISLOCATION_CHANGED"]);

export function reduce(state: UIState, action: Action): UIState {
  switch (action.type) {
    case "HYDRATE":
      return { ...state, snapshot: action.snapshot, stale: false, connected: true, lastOk: action.snapshot.generated_at, error: "" };
    case "STALE":
      return { ...state, stale: true, connected: false };
    case "CONNECTED":
      return { ...state, connected: action.value, stale: action.value ? state.stale : true };
    case "VIEW":
      return { ...state, view: action.view, drawer: null };
    case "THEME":
      persistTheme(action.theme);
      return { ...state, theme: action.theme };
    case "DRAWER":
      return { ...state, drawer: action.drawer };
    case "PERIOD":
      return { ...state, period: action.period };
    case "SORT":
      if (state.sortKey === action.key) {
        return { ...state, sortDir: state.sortDir === "asc" ? "desc" : "asc" };
      }
      return { ...state, sortKey: action.key, sortDir: "desc" };
    case "ERROR":
      return { ...state, error: action.error, stale: true };
    case "EVENT":
      return applyEvent(state, action.event);
    default:
      return state;
  }
}

export function needsImmediateRefetch(ev: StreamEvent): boolean {
  return IMMEDIATE.has(ev.type);
}

function applyEvent(state: UIState, ev: StreamEvent): UIState {
  if (!state.snapshot) return { ...state, connected: true };
  if (ev.type === "DATA_HEALTH_CHANGED" && typeof ev.data === "string") {
    return {
      ...state,
      connected: true,
      snapshot: {
        ...state.snapshot,
        environment: { ...state.snapshot.environment, data_health: ev.data },
      },
    };
  }
  if (ev.type === "DISLOCATION_CHANGED" && typeof ev.data === "string") {
    return {
      ...state,
      connected: true,
      snapshot: {
        ...state.snapshot,
        world: { ...state.snapshot.world, dislocation: ev.data },
      },
    };
  }
  return { ...state, connected: true };
}

export function sortMarkets<T extends { salience: number; attention: number; change_pct: number; research_display: string; demo_display: string; label: string }>(
  rows: T[],
  key: string,
  dir: "asc" | "desc",
): T[] {
  const copy = [...rows];
  const m = dir === "asc" ? 1 : -1;
  copy.sort((a, b) => {
    switch (key) {
      case "attention":
        return (a.attention - b.attention) * m;
      case "move":
        return (a.change_pct - b.change_pct) * m;
      case "research":
        return a.research_display.localeCompare(b.research_display) * m;
      case "demo":
        return a.demo_display.localeCompare(b.demo_display) * m;
      case "market":
        return a.label.localeCompare(b.label) * m;
      default:
        return (a.salience - b.salience) * m;
    }
  });
  return copy;
}
