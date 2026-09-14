import type { DayPnL, Snapshot } from "../store/types";
import { dispatch } from "../store/store";
import { money, num, pct, rMult, signedClass } from "../lib/format";
import { Empty, LineChart, Metric, Panel, ScoreBar } from "../components/ui";
import { WorldMap } from "../map/WorldMap";

export function Overview({ snap }: { snap: Snapshot }) {
  const p = snap.portfolio;
  const top = [...snap.markets].sort((a, b) => b.salience - a.salience).slice(0, 6);
  return (
    <div>
      <div className="kpi-strip">
        <Metric label="Portfolio equity" value={money(p.equity)} sub="BROKER DEMO" />
        <Metric label="Today" value={money(p.today_pnl)} sub={rMult(p.today_r)} tone={signedClass(p.today_pnl)} />
        <Metric label="Drawdown" value={pct(p.drawdown_pct)} />
        <Metric label="Open risk" value={money(p.open_risk)} sub={`cap ${money(p.open_risk_cap)}`} />
        <Metric label="Regime" value={snap.world.regime_display || "—"} />
        <Metric label="Dislocation" value={snap.world.dislocation_display || "Normal"} />
        <Metric label="Data" value={snap.environment.data_health} />
        <Metric label="Pilots" value={`${p.pilots_armed} armed`} sub="GOLD · US100" />
      </div>
      <div className="overview-grid">
        <div className="overview-map">
          <Panel title="Global capital map">
            <WorldMap
              compact
              regions={snap.regions || []}
              markets={snap.markets}
              flows={snap.flows || []}
              onRegion={(id) => dispatch({ type: "DRAWER", drawer: { kind: "region", id } })}
              onMarket={(id) => dispatch({ type: "DRAWER", drawer: { kind: "market", id } })}
            />
          </Panel>
        </div>
        <div className="overview-side">
          <Panel title="World summary">
            <p className="world-copy">{snap.narratives.world}</p>
          </Panel>
          <Panel title="Market pulse">
            {top.map((m) => (
              <div key={m.id} className="pulse-row" onClick={() => dispatch({ type: "DRAWER", drawer: { kind: "market", id: m.id } })}>
                <strong>{m.label}</strong>
                <div>
                  <div className="attn-row">
                    <span className="muted">Salience</span>
                    <ScoreBar kind="sal" value={m.salience} />
                    <span className="num">{num(m.salience, 0)}</span>
                  </div>
                  <div className="attn-row">
                    <span className="muted">Attention</span>
                    <ScoreBar kind="cov" value={m.attention} />
                    <span className="num">{num(m.attention, 0)}</span>
                  </div>
                </div>
                <span className="secondary">{m.trend_display}</span>
              </div>
            ))}
          </Panel>
          <Panel title="What matters now">
            <ul className="matters">
              {(snap.narratives.matters || []).map((s) => (
                <li key={s}>{s}</li>
              ))}
            </ul>
          </Panel>
        </div>
        <div className="overview-perf">
          <Panel title="Bot performance · BROKER DEMO" action={<span className="muted">Realized / Unrealized / R</span>}>
            <div className="kpi-strip" style={{ marginBottom: 12 }}>
              <Metric label="Realized" value={money(p.realized)} />
              <Metric label="Unrealized" value={money(p.unrealized)} />
              <Metric label="Net" value={money(p.net)} />
              <Metric label="Return" value={`${pct(p.return_pct)} · ${rMult(p.all_time_r)}`} />
            </div>
            {p.has_broker_trades ? (
              <>
                <LineChart points={p.equity_curve || []} />
                <PnLCalendar days={p.calendar || []} />
                <div className="secondary" style={{ marginTop: 8 }}>
                  Week {money(p.weekly_total)} · Month {money(p.monthly_total)}
                </div>
              </>
            ) : (
              <Empty>No broker-backed strategy trades yet.</Empty>
            )}
          </Panel>
        </div>
        <div className="overview-news">
          <Panel title="News intelligence">
            {(snap.news || []).length === 0 ? (
              <Empty>No news available.</Empty>
            ) : (
              snap.news.slice(0, 5).map((n) => (
                <div key={n.id} className="news-row" onClick={() => dispatch({ type: "DRAWER", drawer: { kind: "news", id: n.id } })}>
                  <div>
                    <div>{n.title}</div>
                    <div className="muted">
                      {n.source} · {n.age} · {(n.markets || []).join(" · ")} · {n.count} sources
                    </div>
                  </div>
                  <span className="badge">{n.importance}</span>
                </div>
              ))
            )}
          </Panel>
          <Panel title="Next events">
            {(snap.events || []).length === 0 ? (
              <Empty>No high-impact events in the next window.</Empty>
            ) : (
              snap.events.slice(0, 4).map((e) => (
                <div key={e.id} className="event-row" onClick={() => dispatch({ type: "DRAWER", drawer: { kind: "event", id: e.id } })}>
                  <div>
                    <div>{e.name}</div>
                    <div className="muted">
                      {e.importance} · {(e.markets || []).join(" · ")}
                    </div>
                  </div>
                  <span className="num">{e.countdown}</span>
                </div>
              ))
            )}
          </Panel>
        </div>
        <div className="overview-activity">
          <Panel title="Live AurumFlow activity" action={<span className="muted">{snap.narratives.execution}</span>}>
            {(snap.activity || []).length === 0 ? (
              <Empty>No recent observational events.</Empty>
            ) : (
              snap.activity.slice(-8).reverse().map((a, i) => (
                <div key={i} className="activity-row">
                  <span className="muted">{(a.at || "").slice(11, 19)}</span>
                  <span>{a.kind}</span>
                  <span>{a.text}</span>
                </div>
              ))
            )}
          </Panel>
        </div>
      </div>
    </div>
  );
}

function PnLCalendar({ days }: { days: DayPnL[] }) {
  const by = Object.fromEntries(days.map((d) => [d.date, d]));
  const max = Math.max(1, ...days.map((d) => Math.abs(d.realized)));
  const start = new Date();
  start.setUTCDate(start.getUTCDate() - 34);
  const cells = [];
  for (let i = 0; i < 35; i++) {
    const d = new Date(start);
    d.setUTCDate(start.getUTCDate() + i);
    const key = d.toISOString().slice(0, 10);
    const row = by[key];
    const mag = row ? Math.abs(row.realized) / max : 0;
    const bg = !row
      ? undefined
      : row.realized >= 0
        ? `color-mix(in srgb, var(--positive) ${Math.round(mag * 80)}%, var(--surface-2))`
        : `color-mix(in srgb, var(--negative) ${Math.round(mag * 80)}%, var(--surface-2))`;
    cells.push(
      <div key={key} className={`cal-cell ${row ? "has" : ""}`} style={{ background: bg }} title={row ? `${key} ${row.realized} ${row.r}R ${row.trades} trades` : key}>
        {d.getUTCDate()}
      </div>,
    );
  }
  return <div className="calendar">{cells}</div>;
}

export function WorldPage({ snap }: { snap: Snapshot }) {
  return (
    <div>
      <p className="world-copy">{snap.narratives.world}</p>
      <div className="secondary" style={{ marginBottom: 16 }}>
        {snap.world.valid_display} · Regime {snap.world.regime_display} · Dislocation {snap.world.dislocation_display} · Liquidity {snap.world.liquidity}
      </div>
      <Panel title="World map">
        <WorldMap
          regions={snap.regions || []}
          markets={snap.markets}
          flows={snap.flows || []}
          onRegion={(id) => dispatch({ type: "DRAWER", drawer: { kind: "region", id } })}
          onMarket={(id) => dispatch({ type: "DRAWER", drawer: { kind: "market", id } })}
        />
      </Panel>
      <div className="region-cards" style={{ marginTop: 16 }}>
        {(snap.regions || []).map((r) => (
          <div key={r.id} className="panel" style={{ padding: 12 }}>
            <div className="label muted">{r.label}</div>
            <div className="num" style={{ fontSize: 22 }}>{num(r.salience, 0)}</div>
            <div className="secondary">{r.narrative}</div>
          </div>
        ))}
      </div>
    </div>
  );
}
