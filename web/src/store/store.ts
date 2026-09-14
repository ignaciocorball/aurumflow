import { useSyncExternalStore } from "react";
import { initialState, needsImmediateRefetch, reduce } from "./reducer";
import type { Action, Snapshot, UIState, ViewId } from "./types";

let state: UIState = initialState();
const listeners = new Set<() => void>();
let es: EventSource | null = null;
let refetchTimer = 0;
let backoff = 1000;

function emit() {
  listeners.forEach((l) => l());
}

export function dispatch(action: Action) {
  state = reduce(state, action);
  emit();
}

export function getState() {
  return state;
}

function subscribe(fn: () => void) {
  listeners.add(fn);
  return () => listeners.delete(fn);
}

export function useUI(): UIState {
  return useSyncExternalStore(subscribe, getState, getState);
}

export async function hydrate() {
  try {
    const r = await fetch("/ui/snapshot");
    if (!r.ok) throw new Error("snapshot " + r.status);
    const snap = (await r.json()) as Snapshot;
    dispatch({ type: "HYDRATE", snapshot: snap });
    backoff = 1000;
  } catch (e) {
    dispatch({ type: "ERROR", error: e instanceof Error ? e.message : "snapshot failed" });
  }
}

function scheduleRefetch(immediate: boolean) {
  if (immediate) {
    void hydrate();
    return;
  }
  if (refetchTimer) return;
  refetchTimer = window.setTimeout(() => {
    refetchTimer = 0;
    void hydrate();
  }, 400);
}

export function connect() {
  if (es) es.close();
  es = new EventSource("/ui/events");
  es.onopen = () => dispatch({ type: "CONNECTED", value: true });
  es.onerror = () => {
    dispatch({ type: "STALE" });
    es?.close();
    es = null;
    window.setTimeout(() => {
      void hydrate();
      connect();
    }, backoff);
    backoff = Math.min(backoff * 2, 15000);
  };
  const types = [
    "WORLD_UPDATED",
    "MARKET_UPDATED",
    "OPPORTUNITY_UPDATED",
    "POSITION_UPDATED",
    "TRADE_COMPLETED",
    "PNL_UPDATED",
    "NEWS_CLUSTER_UPDATED",
    "ECON_EVENT_UPDATED",
    "DATA_HEALTH_CHANGED",
    "DISLOCATION_CHANGED",
  ];
  for (const t of types) {
    es.addEventListener(t, (ev) => {
      try {
        const data = JSON.parse((ev as MessageEvent).data);
        dispatch({ type: "EVENT", event: data });
        scheduleRefetch(needsImmediateRefetch(data));
      } catch {
        scheduleRefetch(false);
      }
    });
  }
}

export function setView(view: ViewId) {
  dispatch({ type: "VIEW", view });
  const url = view === "overview" ? "#" : `#/${view}`;
  if (location.hash !== url) history.replaceState(null, "", url);
}

export function viewFromHash(): ViewId {
  const h = location.hash.replace(/^#\/?/, "");
  const ok: ViewId[] = ["overview", "world", "markets", "intelligence", "execution", "micro", "research"];
  return (ok as string[]).includes(h) ? (h as ViewId) : "overview";
}
