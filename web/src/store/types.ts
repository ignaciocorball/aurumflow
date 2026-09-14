export type ViewId =
  | "overview"
  | "world"
  | "markets"
  | "intelligence"
  | "execution"
  | "micro"
  | "research";

export type ThemePref = "system" | "light" | "dark";

export interface Snapshot {
  generated_at: string;
  fixture_mode: boolean;
  environment: Environment;
  world: WorldView;
  markets: Market[];
  portfolio: Portfolio;
  execution: ExecutionView;
  research: ResearchView;
  news: NewsCluster[];
  events: EconEvent[];
  health: HealthItem[];
  narratives: Narratives;
  activity: ActivityItem[];
  micro: MicroView;
  regions: RegionView[];
  flows: FlowView[];
}

export interface Environment {
  name: string;
  demo: boolean;
  live_possible: string;
  live_possible_display: string;
  account_masked: string;
  broker_status: string;
  world_valid: string;
  world_valid_display: string;
  data_health: string;
  utc: string;
}

export interface WorldView {
  as_of: string;
  valid: string;
  valid_display: string;
  hash: string;
  session: string;
  regime: string;
  regime_display: string;
  dislocation: string;
  dislocation_display: string;
  confidence: number;
  liquidity: string;
  usd: string;
}

export interface Market {
  id: string;
  label: string;
  region: string;
  region_label: string;
  asset_class: string;
  bid: number;
  ask: number;
  mid: number;
  change_pct: number;
  spark: number[];
  attention: number;
  salience: number;
  coverage: number;
  trend: string;
  trend_display: string;
  regime: string;
  setup: string;
  setup_display: string;
  session: string;
  session_display: string;
  research: string;
  research_display: string;
  demo_status: string;
  demo_display: string;
  quality: string;
  quality_display: string;
  market_status: string;
  market_status_display: string;
  eligibility: string;
  positioning: string;
  flow_context: string;
  salience_reason: string;
  evidence: string[] | null;
  pilot: boolean;
}

export interface RegionView {
  id: string;
  label: string;
  salience: number;
  attention: number;
  coverage: number;
  quality: string;
  markets: string[];
  narrative: string;
  next_event: string;
}

export interface FlowView {
  from: string;
  to: string;
  class: string;
  class_display: string;
  evidence: string;
  confidence: string;
  strength: number;
}

export interface Portfolio {
  equity: number;
  currency: string;
  today_pnl: number;
  today_r: number;
  unrealized: number;
  realized: number;
  net: number;
  return_pct: number;
  drawdown_pct: number;
  open_risk: number;
  open_risk_cap: number;
  pilot_cap: number;
  aggregate_cap: number;
  precious_risk: number;
  precious_cap: number;
  us_equity_risk: number;
  us_equity_cap: number;
  scope: string;
  pilots_armed: number;
  calendar: DayPnL[] | null;
  weekly_total: number;
  monthly_total: number;
  all_time: number;
  all_time_r: number;
  trade_count: number;
  equity_curve: { t: string; v: number }[] | null;
  has_broker_trades: boolean;
}

export interface DayPnL {
  date: string;
  realized: number;
  unrealized: number;
  r: number;
  trades: number;
  wins: number;
  losses: number;
  best: number;
  worst: number;
}

export interface ExecutionView {
  banner: string;
  positions: Position[];
  risk_groups: RiskBar[];
  halt: boolean;
  trust: TrustItem[];
}

export interface Position {
  market: string;
  origin: string;
  origin_display: string;
  kind: string;
  strategy: string;
  direction: string;
  entry: number;
  current: number;
  sl: number;
  tp: number;
  upl: number;
  return_r: number;
  planned_risk: number;
  actual_risk: number;
  account: string;
  open: boolean;
  lifecycle: LifeStage[];
  lifecycle_label: string;
  broker_reconcile: string;
  session: string;
  waiting: string;
  last_execution: string;
  deal_ref: string;
  size: number;
  eligible: string;
  armed: string;
}

export interface LifeStage {
  id: string;
  label: string;
  state: string;
}

export interface RiskBar {
  id: string;
  label: string;
  used: number;
  cap: number;
}

export interface TrustItem {
  code: string;
  display: string;
  hint: string;
  market: string;
}

export interface ResearchView {
  rows: ResearchRow[];
}

export interface ResearchRow {
  market: string;
  status: string;
  status_display: string;
  holdout_n: number;
  discovery_n: number;
  expectancy: number;
  profit_factor: number;
  hit: number;
  max_dd: number;
  sample: number;
  spec: string;
  spec_hash: string;
  note: string;
  dataset: string;
}

export interface NewsCluster {
  id: string;
  title: string;
  summary: string;
  source: string;
  sources: string[];
  url: string;
  published: string;
  age: string;
  markets: string[];
  themes: string[];
  relevance: number;
  importance: string;
  count: number;
  region: string;
}

export interface EconEvent {
  id: string;
  name: string;
  region: string;
  at: string;
  importance: string;
  source: string;
  source_url: string;
  asset_classes: string[];
  markets: string[];
  status: string;
  window: string;
  countdown: string;
  actual: string;
  forecast: string;
  previous: string;
}

export interface HealthItem {
  id: string;
  label: string;
  status: string;
  display: string;
  last_update: string;
  detail: string;
}

export interface Narratives {
  world: string;
  matters: string[];
  execution: string;
  health: string;
}

export interface ActivityItem {
  at: string;
  kind: string;
  market: string;
  text: string;
  process: string;
}

export interface MicroView {
  price: number;
  microprice: number;
  spread: number;
  cvd: number;
  pressure: number;
  imbalance: number;
  imb1: number;
  imb5: number;
  imb10: number;
  book_synced: boolean;
  book_age_ms: number;
  gaps: number;
  resyncs: number;
  drops: number;
  latency_p50: number;
  latency_p95: number;
  latency_p99: number;
  provider: string;
  capability: string;
  absorption: string;
  exhaustion: string;
  integrity: string;
  event_rate: number;
  series: { t: number; price: number; microprice: number; pressure: number; cvd: number }[] | null;
  bids: { price: number; qty: number }[] | null;
  asks: { price: number; qty: number }[] | null;
  resync_reason: string;
}

export interface StreamEvent {
  type: string;
  at: string;
  data?: unknown;
}

export interface UIState {
  snapshot: Snapshot | null;
  connected: boolean;
  stale: boolean;
  lastOk: string;
  view: ViewId;
  theme: ThemePref;
  drawer: DrawerState | null;
  period: "day" | "week" | "month" | "all";
  sortKey: string;
  sortDir: "asc" | "desc";
  error: string;
}

export type DrawerState =
  | { kind: "news"; id: string }
  | { kind: "event"; id: string }
  | { kind: "region"; id: string }
  | { kind: "market"; id: string }
  | { kind: "position"; id: string };

export type Action =
  | { type: "HYDRATE"; snapshot: Snapshot }
  | { type: "EVENT"; event: StreamEvent }
  | { type: "STALE" }
  | { type: "CONNECTED"; value: boolean }
  | { type: "VIEW"; view: ViewId }
  | { type: "THEME"; theme: ThemePref }
  | { type: "DRAWER"; drawer: DrawerState | null }
  | { type: "PERIOD"; period: UIState["period"] }
  | { type: "SORT"; key: string }
  | { type: "ERROR"; error: string };

export const THEME_KEY = "aurumflow.theme";
