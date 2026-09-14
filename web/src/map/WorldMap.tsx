import { useMemo, useState } from "react";
import type { FlowView, Market, RegionView } from "../store/types";
import { num } from "../lib/format";

const NODES: { id: string; x: number; y: number; r: number }[] = [
  { id: "US100", x: 232, y: 168, r: 7 },
  { id: "US500", x: 248, y: 186, r: 6 },
  { id: "US30", x: 218, y: 190, r: 5 },
  { id: "UK100", x: 468, y: 142, r: 5 },
  { id: "DE40", x: 498, y: 154, r: 6 },
  { id: "J225", x: 812, y: 178, r: 7 },
  { id: "CN50", x: 748, y: 206, r: 6 },
  { id: "GOLD", x: 430, y: 360, r: 7 },
  { id: "SILVER", x: 458, y: 378, r: 5 },
  { id: "OIL_CRUDE", x: 540, y: 248, r: 6 },
  { id: "BTC", x: 500, y: 72, r: 6 },
];

const REGION_SHAPES: { id: string; d: string }[] = [
  { id: "NORTH_AMERICA", d: "M118 96 C170 70 250 78 292 128 C320 168 300 230 248 252 C190 274 120 240 108 186 C98 142 92 116 118 96 Z" },
  { id: "UNITED_KINGDOM", d: "M452 118 C466 112 476 122 474 136 C468 148 454 146 450 134 C448 124 446 120 452 118 Z" },
  { id: "EUROPE", d: "M478 128 C530 108 572 128 586 168 C572 198 520 210 492 188 C474 170 466 146 478 128 Z" },
  { id: "JAPAN", d: "M792 150 C818 146 834 168 826 196 C808 214 786 200 782 176 C780 160 780 152 792 150 Z" },
  { id: "CHINA_HONG_KONG", d: "M700 168 C754 150 790 186 776 230 C748 252 700 240 688 208 C684 186 688 172 700 168 Z" },
  { id: "ASIA_PACIFIC", d: "M640 250 C720 230 800 260 820 310 C780 340 680 336 640 300 C628 276 628 258 640 250 Z" },
  { id: "COMMODITIES", d: "M360 330 C470 318 560 350 540 390 C470 412 380 400 360 360 C352 344 348 334 360 330 Z" },
];

function salienceFill(v: number): string {
  const t = Math.min(1, Math.max(0, v / 100));
  return `color-mix(in srgb, var(--map-active) ${Math.round(t * 70)}%, var(--map-land))`;
}

function qualityStroke(q: string): string {
  const u = (q || "").toUpperCase();
  if (u === "HEALTHY") return "var(--healthy)";
  if (u === "DEGRADED" || u === "STALE") return "var(--degraded)";
  if (u === "OFFLINE") return "var(--critical)";
  return "var(--border-strong)";
}

export function WorldMap({
  regions,
  markets,
  flows,
  onRegion,
  onMarket,
  compact,
}: {
  regions: RegionView[];
  markets: Market[];
  flows: FlowView[];
  onRegion: (id: string) => void;
  onMarket: (id: string) => void;
  compact?: boolean;
}) {
  const [hover, setHover] = useState<string | null>(null);
  const byId = useMemo(() => Object.fromEntries(markets.map((m) => [m.id, m])), [markets]);
  const regionBy = useMemo(() => Object.fromEntries(regions.map((r) => [r.id, r])), [regions]);
  const hoveredRegion = hover ? regionBy[hover] : undefined;
  const hoveredMarket = hover ? byId[hover] : undefined;

  return (
    <div className="map-wrap">
      <svg viewBox="0 0 960 460" role="img" aria-label="Global capital map">
        <rect width="960" height="460" fill="var(--map-ocean)" />
        <g opacity="0.25" stroke="var(--border-subtle)" fill="none">
          {Array.from({ length: 8 }, (_, i) => (
            <line key={i} x1="40" x2="920" y1={50 + i * 48} />
          ))}
        </g>
        {REGION_SHAPES.map((s) => {
          const r = regionBy[s.id];
          return (
            <path
              key={s.id}
              d={s.d}
              fill={salienceFill(r?.salience || 0)}
              stroke={qualityStroke(r?.quality || "")}
              strokeWidth={r && r.quality === "HEALTHY" ? 1.2 : 1.8}
              className="map-node"
              onMouseEnter={() => setHover(s.id)}
              onMouseLeave={() => setHover(null)}
              onClick={() => onRegion(s.id)}
            />
          );
        })}
        {flows.map((f, i) => {
          const a = NODES.find((n) => n.id === f.from) || NODES[0];
          const b = NODES.find((n) => n.id === f.to) || NODES[7];
          return (
            <line
              key={i}
              x1={a.x}
              y1={a.y}
              x2={b.x}
              y2={b.y}
              stroke="var(--attention)"
              strokeWidth={1 + f.strength / 50}
              strokeDasharray={f.class === "OBSERVED" ? undefined : "5 4"}
              opacity="0.7"
            >
              <title>{`${f.from} → ${f.to} · ${f.class_display} · ${f.evidence}`}</title>
            </line>
          );
        })}
        {NODES.map((n) => {
          const m = byId[n.id];
          const pulse = m && m.salience >= 70;
          return (
            <g key={n.id} className="map-node" onClick={() => onMarket(n.id)} onMouseEnter={() => setHover(n.id)} onMouseLeave={() => setHover(null)}>
              {pulse ? <circle cx={n.x} cy={n.y} r={n.r + 7} fill="none" stroke="var(--salience)" opacity="0.45" /> : null}
              <circle cx={n.x} cy={n.y} r={n.r} fill="var(--surface-elevated)" stroke="var(--map-active)" strokeWidth="1.5" />
              <text x={n.x + 10} y={n.y + 4} fontSize="10" fill="var(--text-primary)">
                {m?.label || n.id}
              </text>
            </g>
          );
        })}
        {(hoveredRegion || hoveredMarket) && (
          <foreignObject x="24" y="360" width="360" height="80">
            <div style={{ background: "var(--surface-elevated)", border: "1px solid var(--border-subtle)", padding: "8px 10px", fontSize: 12 }}>
              {hoveredMarket ? (
                <>
                  <strong>{hoveredMarket.label}</strong> · Attention {num(hoveredMarket.attention, 0)} · Salience {num(hoveredMarket.salience, 0)} · {hoveredMarket.trend_display}
                </>
              ) : hoveredRegion ? (
                <>
                  <strong>{hoveredRegion.label}</strong> · Salience {num(hoveredRegion.salience, 0)} · {hoveredRegion.narrative}
                </>
              ) : null}
            </div>
          </foreignObject>
        )}
      </svg>
      {!compact && (
        <div className="map-a11y">
          <table className="radar">
            <thead>
              <tr>
                <th>Region</th>
                <th>Markets</th>
                <th>Attention</th>
                <th>Salience</th>
                <th>Quality</th>
              </tr>
            </thead>
            <tbody>
              {regions.map((r) => (
                <tr key={r.id} onClick={() => onRegion(r.id)}>
                  <td>{r.label}</td>
                  <td>{r.markets.join(" · ")}</td>
                  <td className="num">{num(r.attention, 0)}</td>
                  <td className="num">{num(r.salience, 0)}</td>
                  <td>{r.quality}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
