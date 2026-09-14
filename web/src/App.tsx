import { useEffect, useMemo, useState, type ReactNode } from "react";
import { Monitor, Moon, Sun } from "lucide-react";
import { applyTheme } from "./lib/theme";
import { healthClass, utcClock } from "./lib/format";
import { connect, dispatch, hydrate, setView, useUI, viewFromHash } from "./store/store";
import { Badge, Empty } from "./components/ui";
import { Overview, WorldPage } from "./pages/Overview";
import { ExecutionPage, IntelligencePage, MarketsPage, MicroPage, ResearchPage } from "./pages/Rest";
import type { ThemePref } from "./store/types";

const VIEWS = [
  ["overview", "Overview"],
  ["world", "World"],
  ["markets", "Markets"],
  ["intelligence", "Intelligence"],
  ["execution", "Execution"],
  ["micro", "Micro"],
  ["research", "Research"],
] as const;

export function App() {
  const ui = useUI();
  const [clock, setClock] = useState(() => utcClock());

  useEffect(() => {
    applyTheme(ui.theme);
    const mq = window.matchMedia("(prefers-color-scheme: dark)");
    const on = () => applyTheme(ui.theme);
    mq.addEventListener("change", on);
    return () => mq.removeEventListener("change", on);
  }, [ui.theme]);

  useEffect(() => {
    dispatch({ type: "VIEW", view: viewFromHash() });
    void hydrate();
    connect();
    const onHash = () => setView(viewFromHash());
    window.addEventListener("hashchange", onHash);
    const t = window.setInterval(() => setClock(utcClock()), 1000);
    return () => {
      window.removeEventListener("hashchange", onHash);
      window.clearInterval(t);
    };
  }, []);

  const snap = ui.snapshot;
  const page = useMemo(() => {
    if (!snap) return ui.error ? <Empty>Unable to load snapshot. {ui.error}</Empty> : <Empty>Connecting to the observatory…</Empty>;
    switch (ui.view) {
      case "world":
        return <WorldPage snap={snap} />;
      case "markets":
        return <MarketsPage snap={snap} sortKey={ui.sortKey} sortDir={ui.sortDir} />;
      case "intelligence":
        return <IntelligencePage snap={snap} />;
      case "execution":
        return <ExecutionPage snap={snap} />;
      case "micro":
        return <MicroPage snap={snap} />;
      case "research":
        return <ResearchPage snap={snap} />;
      default:
        return <Overview snap={snap} />;
    }
  }, [snap, ui.view, ui.sortKey, ui.sortDir, ui.error]);

  const env = snap?.environment;
  const health = env?.data_health || "UNKNOWN";
  const world = env?.world_valid_display || "World unknown";

  return (
    <div className="app">
      <header className="header">
        <div className="brand">AURUMFLOW</div>
        <div className="header-meta">
          <Badge tone="healthy">{world}</Badge>
          <Badge tone={healthClass(health)}>{health === "DEGRADED" ? "DATA DEGRADED" : health === "OFFLINE" ? "DATA OFFLINE" : "DATA HEALTHY"}</Badge>
          <Badge tone="demo">DEMO</Badge>
          {env?.broker_status ? <Badge>{env.broker_status}</Badge> : null}
          {ui.stale ? <Badge tone="degraded">STALE</Badge> : null}
          {snap?.fixture_mode ? <Badge tone="critical">FIXTURE MODE</Badge> : null}
        </div>
        <div className="header-right">
          <div className="clock" aria-label="UTC clock">
            {clock}
          </div>
          <ThemeControl value={ui.theme} />
        </div>
      </header>
      <nav className="nav" aria-label="Primary">
        {VIEWS.map(([id, lab]) => (
          <button key={id} aria-current={ui.view === id ? "page" : undefined} onClick={() => setView(id)}>
            {lab}
          </button>
        ))}
      </nav>
      <main className="page">{page}</main>
      {ui.drawer && snap ? <Drawer /> : null}
    </div>
  );
}

function ThemeControl({ value }: { value: ThemePref }) {
  const cycle: ThemePref[] = ["system", "light", "dark"];
  const next = cycle[(cycle.indexOf(value) + 1) % cycle.length];
  const Icon = value === "dark" ? Moon : value === "light" ? Sun : Monitor;
  return (
    <button className="icon-btn" aria-label={`Theme ${value}`} onClick={() => dispatch({ type: "THEME", theme: next })}>
      <Icon size={16} />
    </button>
  );
}

function Drawer() {
  const ui = useUI();
  const snap = ui.snapshot;
  if (!ui.drawer || !snap) return null;
  const close = () => dispatch({ type: "DRAWER", drawer: null });
  let title = "";
  let body: ReactNode = null;
  if (ui.drawer.kind === "news") {
    const n = snap.news.find((x) => x.id === ui.drawer!.id);
    title = n?.title || "News";
    body = n ? (
      <div>
        <p className="secondary">{n.summary || "Title and metadata only."}</p>
        <p>
          Source {n.source} · {n.age} · Relevance {Math.round(n.relevance)} (news relevance, not a trade score)
        </p>
        <p>Markets {(n.markets || []).join(" · ") || "—"}</p>
        <p>Themes {(n.themes || []).join(" · ") || "—"}</p>
        {n.url ? (
          <p>
            <a href={n.url} target="_blank" rel="noreferrer noopener">
              Open source
            </a>
          </p>
        ) : null}
      </div>
    ) : (
      <Empty>Cluster not found.</Empty>
    );
  } else if (ui.drawer.kind === "event") {
    const e = snap.events.find((x) => x.id === ui.drawer!.id);
    title = e?.name || "Event";
    body = e ? (
      <div>
        <p>
          {e.countdown} · {e.status} · {e.window}
        </p>
        <p>Source {e.source}</p>
        <p>Markets {(e.markets || []).join(" · ")}</p>
        {e.source_url ? (
          <p>
            <a href={e.source_url} target="_blank" rel="noreferrer noopener">
              Official schedule
            </a>
          </p>
        ) : null}
        <p className="muted">No forecast is implied. Actual/forecast shown only when the official source provides them.</p>
      </div>
    ) : (
      <Empty>Event not found.</Empty>
    );
  } else if (ui.drawer.kind === "region") {
    const r = snap.regions.find((x) => x.id === ui.drawer!.id);
    title = r?.label || "Region";
    body = r ? (
      <div>
        <p>{r.narrative}</p>
        <p>Attention {Math.round(r.attention)} · Salience {Math.round(r.salience)} · Coverage {Math.round(r.coverage)}</p>
        <p>Markets {r.markets.join(" · ")}</p>
        {r.next_event ? <p>Next official event: {r.next_event}</p> : <p>No mapped official event.</p>}
      </div>
    ) : (
      <Empty>Region not found.</Empty>
    );
  } else if (ui.drawer.kind === "market") {
    const m = snap.markets.find((x) => x.id === ui.drawer!.id);
    title = m?.label || "Market";
    body = m ? (
      <div>
        <p className="num" style={{ fontSize: 28 }}>
          {m.mid ? m.mid.toLocaleString() : "—"}
        </p>
        <p>{m.trend_display}</p>
        <p>Structural attention {Math.round(m.attention)}</p>
        <p>Market salience {Math.round(m.salience)}</p>
        <p>Evidence coverage {m.coverage.toFixed(1)}%</p>
        <p>State {m.setup_display} · Broker {m.market_status_display}</p>
        <p className="muted">{(m.evidence || []).join(" · ")}</p>
      </div>
    ) : (
      <Empty>Market not found.</Empty>
    );
  }
  return (
    <>
      <div className="drawer-back" onClick={close} />
      <aside className="drawer" role="dialog" aria-label={title}>
        <div className="panel-h">
          <h2>{title}</h2>
          <button className="icon-btn" onClick={close} aria-label="Close">
            ×
          </button>
        </div>
        {body}
      </aside>
    </>
  );
}
