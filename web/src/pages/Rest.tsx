import type { Snapshot } from "../store/types";
import { dispatch } from "../store/store";
import { sortMarkets } from "../store/reducer";
import { kindClass, money, num, pct, rMult, signedClass, statusTone } from "../lib/format";
import { AttentionRing, Empty, Lifecycle, Panel, RiskGauge, ScoreBar, Sparkline } from "../components/ui";
import { Badge } from "../components/ui";

export function MarketsPage({ snap, sortKey, sortDir }: { snap: Snapshot; sortKey: string; sortDir: "asc" | "desc" }) {
  const rows = sortMarkets(snap.markets, sortKey, sortDir);
  const cols: [string, string][] = [
    ["market", "Market"],
    ["price", "Price"],
    ["move", "Change"],
    ["spark", "Spark"],
    ["attention", "Attention"],
    ["salience", "Salience"],
    ["coverage", "Coverage"],
    ["trend", "Regime"],
    ["setup", "Setup"],
    ["session", "Session"],
    ["research", "Research"],
    ["demo", "DEMO"],
    ["quality", "Data"],
  ];
  const sortable = new Set(["salience", "attention", "move", "research", "demo", "market"]);
  return (
    <Panel title="Market radar">
      <div className="table-wrap">
        <table className="radar">
          <thead>
            <tr>
              {cols.map(([k, lab]) => (
                <th key={k} onClick={() => sortable.has(k) && dispatch({ type: "SORT", key: k })}>
                  {lab}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {rows.map((m) => (
              <tr key={m.id} onClick={() => dispatch({ type: "DRAWER", drawer: { kind: "market", id: m.id } })}>
                <td>
                  <strong>{m.label}</strong>
                </td>
                <td className="num">{num(m.mid, m.mid > 1000 ? 2 : 3)}</td>
                <td className={`num ${signedClass(m.change_pct)}`}>{pct(m.change_pct)}</td>
                <td>
                  <Sparkline values={m.spark || []} />
                </td>
                <td>
                  <AttentionRing value={m.attention} />
                </td>
                <td>
                  <ScoreBar kind="sal" value={m.salience} />
                  <span className="num">{num(m.salience, 0)}</span>
                </td>
                <td>
                  <ScoreBar kind="cov" value={m.coverage} />
                  <span className="num">{num(m.coverage, 1)}%</span>
                </td>
                <td>{m.trend_display}</td>
                <td>{m.setup_display}</td>
                <td>{m.session_display}</td>
                <td>{m.research_display}</td>
                <td>{m.demo_display}</td>
                <td>{m.quality_display}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </Panel>
  );
}

export function IntelligencePage({ snap }: { snap: Snapshot }) {
  const ranked = [...snap.markets].sort((a, b) => b.salience - a.salience);
  return (
    <div className="overview-grid">
      <Panel title="Attention vs salience vs coverage">
        {snap.markets.map((m) => (
          <div key={m.id} className="market-card">
            <div>
              <strong>{m.label}</strong>
              <div className="muted">Structural attention is not unusual activity.</div>
            </div>
            <div style={{ width: 220 }}>
              <div className="attn-row">
                <span>Attention</span>
                <ScoreBar kind="cov" value={m.attention} />
                <span className="num">{num(m.attention, 0)}</span>
              </div>
              <div className="attn-row">
                <span>Salience</span>
                <ScoreBar kind="sal" value={m.salience} />
                <span className="num">{num(m.salience, 0)}</span>
              </div>
              <div className="attn-row">
                <span>Coverage</span>
                <ScoreBar kind="cov" value={m.coverage} />
                <span className="num">{num(m.coverage, 0)}</span>
              </div>
            </div>
          </div>
        ))}
      </Panel>
      <div>
        <Panel title="Dislocation">
          <p className="world-copy">{snap.world.dislocation_display}</p>
          <p className="secondary">{snap.narratives.world}</p>
        </Panel>
        <Panel title="Top anomalies">
          {ranked.slice(0, 6).map((m) => (
            <div key={m.id} className="pulse-row">
              <strong>{m.label}</strong>
              <span className="secondary">{m.trend_display}</span>
              <span className="num">{num(m.salience, 0)}</span>
            </div>
          ))}
        </Panel>
      </div>
    </div>
  );
}

export function ExecutionPage({ snap }: { snap: Snapshot }) {
  return (
    <div>
      <div className="banner">{snap.execution.banner}</div>
      <p className="secondary" style={{ marginBottom: 16 }}>
        {snap.narratives.execution}
      </p>
      <div className="overview-grid">
        <div>
          {snap.execution.positions.map((p) => (
            <Panel
              key={p.market}
              title={`${p.market} · ${p.origin_display}`}
              action={
                <span>
                  <Badge tone={kindClass(p.kind)}>{p.kind}</Badge>
                </span>
              }
            >
              {p.open ? (
                <div className="kpi-strip">
                  <div className="kpi">
                    <div className="label">Direction</div>
                    <div className="value">{p.direction || "—"}</div>
                  </div>
                  <div className="kpi">
                    <div className="label">Entry</div>
                    <div className="value num">{num(p.entry)}</div>
                  </div>
                  <div className="kpi">
                    <div className="label">Current</div>
                    <div className="value num">{num(p.current)}</div>
                  </div>
                  <div className="kpi">
                    <div className="label">UPL</div>
                    <div className={`value num ${signedClass(p.upl)}`}>{money(p.upl)}</div>
                  </div>
                </div>
              ) : (
                <Empty>No position open. {p.waiting || "Waiting for the next valid setup."}</Empty>
              )}
              <div className="muted" style={{ margin: "8px 0" }}>
                SL {num(p.sl)} · TP {num(p.tp)} · Planned {money(p.planned_risk)} · {rMult(p.return_r)} · Account {p.account || "masked"} · Broker {p.broker_reconcile}
              </div>
              <Lifecycle stages={p.lifecycle || []} />
            </Panel>
          ))}
        </div>
        <div>
          <Panel title="Risk">
            {(snap.execution.risk_groups || []).map((g) => (
              <RiskGauge key={g.id} label={g.label} used={g.used} cap={g.cap || 5} />
            ))}
          </Panel>
          <Panel title="Operational trust">
            {(snap.execution.trust || []).map((t) => (
              <div key={t.market} className="market-card">
                <div>
                  <strong>{t.market}</strong>
                  <div className="secondary">{t.hint}</div>
                </div>
                <Badge tone={statusTone(t.code)}>{t.display || "—"}</Badge>
              </div>
            ))}
          </Panel>
        </div>
      </div>
    </div>
  );
}

export function MicroPage({ snap }: { snap: Snapshot }) {
  const m = snap.micro;
  const cvd = (m.series || []).map((p) => ({ t: p.t, v: p.cvd }));
  const pr = (m.series || []).map((p) => ({ t: p.t, v: p.pressure }));
  return (
    <div>
      <div className="kpi-strip">
        <div className="kpi">
          <div className="label">BTC</div>
          <div className="value num">{num(m.price, 2)}</div>
        </div>
        <div className="kpi">
          <div className="label">Spread</div>
          <div className="value num">{num(m.spread, 2)}</div>
        </div>
        <div className="kpi">
          <div className="label">Book</div>
          <div className="value">{m.book_synced ? "Synced" : "Resync"}</div>
        </div>
        <div className="kpi">
          <div className="label">Latency p95</div>
          <div className="value num">{num(m.latency_p95, 1)} ms</div>
        </div>
        <div className="kpi">
          <div className="label">CVD</div>
          <div className="value num">{num(m.cvd, 1)}</div>
        </div>
        <div className="kpi">
          <div className="label">Pressure</div>
          <div className="value num">{num(m.pressure, 2)}</div>
        </div>
      </div>
      <div className="micro-grid">
        <Panel title="CVD">
          <Sparkline values={cvd.map((p) => p.v)} />
          {cvd.length < 2 ? <Empty>Waiting for microstructure series.</Empty> : null}
        </Panel>
        <Panel title="Pressure">
          <Sparkline values={pr.map((p) => p.v)} />
        </Panel>
        <Panel title="Book (L2, not MBO)">
          <div className="book">
            <div>
              {(m.bids || []).map((b, i) => (
                <div key={i}>
                  {num(b.price, 1)} · {num(b.qty, 3)}
                </div>
              ))}
            </div>
            <div>
              {(m.asks || []).map((b, i) => (
                <div key={i}>
                  {num(b.price, 1)} · {num(b.qty, 3)}
                </div>
              ))}
            </div>
          </div>
          <div className="muted" style={{ marginTop: 8 }}>
            Imbalance {num(m.imbalance, 3)} · 1 {num(m.imb1, 3)} · 5 {num(m.imb5, 3)} · gaps {m.gaps} · drops {m.drops} · resyncs {m.resyncs}
          </div>
        </Panel>
        <Panel title="Quality">
          <div>Integrity {m.integrity || "—"}</div>
          <div>Absorption {m.absorption || "—"}</div>
          <div>Classification {m.exhaustion || "—"}</div>
          <div>Provider {m.provider || "—"}</div>
          <div>Capability {m.capability || "—"}</div>
          <div>Age {m.book_age_ms} ms</div>
        </Panel>
      </div>
    </div>
  );
}

export function ResearchPage({ snap }: { snap: Snapshot }) {
  return (
    <Panel title="Validation matrix">
      <div className="table-wrap">
        <table className="radar">
          <thead>
            <tr>
              <th>Market</th>
              <th>Status</th>
              <th>Holdout n</th>
              <th>Discovery n</th>
              <th>Expectancy</th>
              <th>Profit factor</th>
              <th>Hit</th>
              <th>Spec</th>
              <th>Note</th>
            </tr>
          </thead>
          <tbody>
            {(snap.research.rows || []).map((r) => (
              <tr key={r.market}>
                <td>{r.market}</td>
                <td>
                  <Badge tone={statusTone(r.status)}>{r.status_display}</Badge>
                </td>
                <td className="num">{r.holdout_n || "—"}</td>
                <td className="num">{r.discovery_n || "—"}</td>
                <td className="num">{r.expectancy || "—"}</td>
                <td className="num">{r.profit_factor || "—"}</td>
                <td className="num">{r.hit ? pct(r.hit * 100, 0) : "—"}</td>
                <td>{r.spec || "—"}</td>
                <td className="secondary">{r.note}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <p className="muted" style={{ marginTop: 12 }}>
        Rejected means evidence did not support promotion. It is not a system fault.
      </p>
    </Panel>
  );
}
