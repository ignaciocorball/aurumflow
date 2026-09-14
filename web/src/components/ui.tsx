import type { CSSProperties, ReactNode } from "react";
import { healthClass, signedClass } from "../lib/format";

export function Panel({ title, action, children }: { title: string; action?: ReactNode; children: ReactNode }) {
  return (
    <section className="panel">
      <div className="panel-h">
        <h2>{title}</h2>
        {action}
      </div>
      <div className="panel-b">{children}</div>
    </section>
  );
}

export function Badge({ children, tone = "" }: { children: ReactNode; tone?: string }) {
  return <span className={`badge ${tone}`}>{children}</span>;
}

export function Empty({ children }: { children: ReactNode }) {
  return <div className="empty">{children}</div>;
}

export function Metric({ label, value, sub, tone }: { label: string; value: string; sub?: string; tone?: string }) {
  return (
    <div className="kpi">
      <div className="label">{label}</div>
      <div className={`value num ${tone || ""}`}>{value}</div>
      {sub ? <div className="sub">{sub}</div> : null}
    </div>
  );
}

export function HealthDot({ status }: { status: string }) {
  return <Badge tone={healthClass(status)}>{status}</Badge>;
}

export function Sparkline({ values, positive }: { values: number[]; positive?: boolean }) {
  if (!values || values.length < 2) {
    return <svg width="72" height="24" aria-hidden="true" />;
  }
  const min = Math.min(...values);
  const max = Math.max(...values);
  const span = max - min || 1;
  const d = values
    .map((v, i) => {
      const x = (i / (values.length - 1)) * 72;
      const y = 22 - ((v - min) / span) * 20;
      return `${i === 0 ? "M" : "L"}${x.toFixed(1)},${y.toFixed(1)}`;
    })
    .join(" ");
  const last = values[values.length - 1];
  const color = positive === undefined ? "currentColor" : last >= values[0] ? "var(--positive)" : "var(--negative)";
  return (
    <svg width="72" height="24" viewBox="0 0 72 24" aria-hidden="true">
      <path d={d} fill="none" stroke={color} strokeWidth="1.4" />
    </svg>
  );
}

export function AttentionRing({ value }: { value: number }) {
  const r = 14;
  const c = 2 * Math.PI * r;
  const off = c * (1 - Math.min(100, Math.max(0, value)) / 100);
  return (
    <svg width="36" height="36" viewBox="0 0 36 36" aria-label={`Attention ${Math.round(value)}`}>
      <circle cx="18" cy="18" r={r} fill="none" stroke="var(--surface-2)" strokeWidth="3" />
      <circle
        cx="18"
        cy="18"
        r={r}
        fill="none"
        stroke="var(--attention)"
        strokeWidth="3"
        strokeDasharray={`${c} ${c}`}
        strokeDashoffset={off}
        transform="rotate(-90 18 18)"
      />
      <text x="18" y="21" textAnchor="middle" fontSize="9" fill="currentColor">
        {Math.round(value)}
      </text>
    </svg>
  );
}

export function ScoreBar({ value, kind }: { value: number; kind: "sal" | "cov" }) {
  return (
    <div className={`bar ${kind}`} aria-label={`${kind} ${Math.round(value)}`}>
      <span style={{ width: `${Math.min(100, Math.max(0, value))}%` } as CSSProperties} />
    </div>
  );
}

export function RiskGauge({ used, cap, label }: { used: number; cap: number; label: string }) {
  const pct = cap > 0 ? Math.min(100, (used / cap) * 100) : 0;
  return (
    <div>
      <div className="panel-h" style={{ padding: "8px 0" }}>
        <h2>{label}</h2>
        <span className="num">
          ${used.toFixed(2)} / ${cap.toFixed(2)}
        </span>
      </div>
      <div className="riskbar">
        <span style={{ width: `${pct}%` }} />
      </div>
    </div>
  );
}

export function Lifecycle({ stages }: { stages: { id: string; label: string; state: string }[] }) {
  return (
    <div className="life" role="list">
      {stages.map((s) => (
        <div key={s.id} className={`st ${s.state}`} role="listitem" aria-label={`${s.label} ${s.state}`}>
          <div className="dot" />
          {s.label}
        </div>
      ))}
    </div>
  );
}

export function LineChart({
  points,
  color = "var(--attention)",
}: {
  points: { t: string | number; v: number }[];
  color?: string;
}) {
  if (!points || points.length < 2) return <Empty>No series yet.</Empty>;
  const vs = points.map((p) => p.v);
  const min = Math.min(...vs);
  const max = Math.max(...vs);
  const span = max - min || 1;
  const w = 320;
  const h = 72;
  const d = points
    .map((p, i) => {
      const x = (i / (points.length - 1)) * w;
      const y = h - 8 - ((p.v - min) / span) * (h - 16);
      return `${i === 0 ? "M" : "L"}${x.toFixed(1)},${y.toFixed(1)}`;
    })
    .join(" ");
  return (
    <svg width="100%" viewBox={`0 0 ${w} ${h}`} role="img" aria-label="equity curve">
      <path d={d} fill="none" stroke={color} strokeWidth="1.6" />
    </svg>
  );
}

export function Signed({ n, text }: { n: number; text: string }) {
  return <span className={`num ${signedClass(n)}`}>{text}</span>;
}
