export function money(n: number | undefined, digits = 2): string {
  if (n === undefined || Number.isNaN(n)) return "—";
  const sign = n < 0 ? "-" : "";
  return `${sign}$${Math.abs(n).toLocaleString("en-US", { minimumFractionDigits: digits, maximumFractionDigits: digits })}`;
}

export function pct(n: number | undefined, digits = 2): string {
  if (n === undefined || Number.isNaN(n)) return "—";
  const sign = n > 0 ? "+" : "";
  return `${sign}${n.toFixed(digits)}%`;
}

export function num(n: number | undefined, digits = 2): string {
  if (n === undefined || Number.isNaN(n)) return "—";
  return n.toLocaleString("en-US", { minimumFractionDigits: digits, maximumFractionDigits: digits });
}

export function rMult(n: number | undefined): string {
  if (n === undefined || Number.isNaN(n)) return "—";
  const sign = n > 0 ? "+" : "";
  return `${sign}${n.toFixed(2)}R`;
}

export function signedClass(n: number): string {
  if (n > 0) return "pos";
  if (n < 0) return "neg";
  return "";
}

export function healthClass(s: string | undefined): string {
  const u = (s || "").toUpperCase();
  if (u.includes("OFFLINE")) return "offline";
  if (u.includes("DEGRAD") || u.includes("STALE") || u.includes("RESYNC")) return "degraded";
  if (u.includes("HEALTH") || u.includes("CURRENT") || u.includes("READY")) return "healthy";
  return "";
}

export function utcClock(iso?: string): string {
  const d = iso ? new Date(iso) : new Date();
  if (Number.isNaN(d.getTime())) return "—";
  return d.toISOString().slice(11, 19) + " UTC";
}

export function localClock(): string {
  return new Date().toLocaleTimeString([], { hour12: false }) + " local";
}

export function kindClass(kind: string): string {
  const u = kind.toUpperCase();
  if (u.includes("SHADOW")) return "shadow";
  if (u.includes("RESEARCH")) return "research";
  return "demo";
}

export function statusTone(s: string): string {
  const u = s.toUpperCase();
  if (u.includes("REJECT")) return "research";
  if (u.includes("ELIGIBLE") || u.includes("VALIDATED") || u.includes("OPERATIONAL")) return "healthy";
  if (u.includes("BLOCK")) return "degraded";
  return "";
}
