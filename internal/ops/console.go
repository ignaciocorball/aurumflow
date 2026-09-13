package ops

const consoleHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1"/>
<title>AURUMFLOW · Global Intelligence</title>
<style>
:root{
  --bg:#090c11;
  --glass-bg:rgba(16,22,32,.62);
  --glass-border:rgba(220,230,245,.08);
  --glass-highlight:rgba(255,255,255,.07);
  --glass-shadow:0 10px 40px rgba(0,0,0,.35);
  --surface-solid:#10161f;
  --surface-elevated:#161d28;
  --text-primary:#e8edf5;
  --text-secondary:#9aa6b8;
  --text-tertiary:#6d7889;
  --ok:#5fbf96; --warn:#d4b15a; --bad:#c97a76; --long:#6aa8d4; --short:#c98a72;
  --ask:#7a4545; --bid:#2f5d4a; --guide:#2c3646;
  --up:#5fbf96; --down:#c98a72; --mix:#d4b15a; --unk:#6d7889;
  --s4:4px; --s8:8px; --s12:12px; --s16:16px; --s24:24px; --s32:32px;
}
*{box-sizing:border-box}
html,body{margin:0;height:100%;overflow:hidden;background:
  radial-gradient(1200px 600px at 12% -10%, rgba(40,70,110,.18), transparent 50%),
  radial-gradient(900px 500px at 90% 0%, rgba(40,90,80,.10), transparent 46%),
  var(--bg);
  color:var(--text-primary);
  font:13px/1.45 Inter,ui-sans-serif,-apple-system,BlinkMacSystemFont,"SF Pro Display","Segoe UI",sans-serif;
  font-variant-numeric:tabular-nums}
body{display:flex;flex-direction:column}
.dot{width:7px;height:7px;border-radius:50%;display:inline-block;margin-right:6px;background:var(--unk)}
.dot.ok{background:var(--ok)}.dot.warn{background:var(--warn)}.dot.bad{background:var(--bad)}
#cmd{display:grid;grid-template-columns:1fr auto 1fr;align-items:center;gap:16px;
  padding:8px 16px;position:sticky;top:0;z-index:8;
  background:var(--glass-bg);backdrop-filter:blur(18px) saturate(1.2);
  border-bottom:1px solid var(--glass-border);
  box-shadow:var(--glass-shadow), inset 0 1px 0 var(--glass-highlight)}
.brand{display:flex;align-items:baseline;gap:10px}
.brand strong{letter-spacing:.16em;font-size:13px}
.brand .sub{color:var(--text-secondary);font-size:11px;letter-spacing:.14em}
.sys{color:var(--text-tertiary);font-size:10px;letter-spacing:.08em}
#sessbar{display:flex;gap:16px;font-size:11px;letter-spacing:.1em;color:var(--text-secondary)}
#sessbar b{color:var(--text-primary);font-weight:600;margin-left:6px}
.saf{display:flex;justify-content:flex-end;align-items:center;gap:12px;font-size:11px;color:var(--text-secondary)}
.pill{padding:2px 8px;border-radius:999px;border:1px solid var(--glass-border);letter-spacing:.08em}
.pill.ok{color:var(--ok)}.pill.bad{color:var(--bad)}
#app{flex:1;min-height:0;overflow:hidden;display:grid;grid-template-columns:88px minmax(0,1fr) 220px;grid-template-rows:minmax(0,1fr)}
#rail{display:flex;flex-direction:column;gap:4px;padding:8px 8px 10px;
  min-height:0;overflow:auto;background:rgba(10,14,20,.45);border-right:1px solid var(--glass-border)}
#rail button{border:0;background:transparent;color:var(--text-tertiary);padding:8px 6px;
  font:10px inherit;letter-spacing:.1em;cursor:pointer;text-align:left;border-radius:8px}
#rail button.on{color:var(--text-primary);background:rgba(255,255,255,.04)}
#rail #dens{margin-top:auto;color:var(--text-tertiary)}
#stage{min-width:0;min-height:0;overflow:auto;padding:12px}
.view{display:none}
.view.on{display:block}
#view-overview.on{display:block}
.ov{display:grid;grid-template-columns:repeat(12,1fr);grid-template-rows:auto auto;gap:12px;align-content:start}
.hero{grid-column:1/9;grid-row:1}
.opp{grid-column:9/13;grid-row:1 / span 3}
.sessrot{grid-column:1/5}
.mstate{grid-column:5/9}
.dec{grid-column:9/13}
.glass{background:var(--glass-bg);backdrop-filter:blur(16px);
  border:1px solid var(--glass-border);border-radius:14px;
  box-shadow:inset 0 1px 0 var(--glass-highlight);padding:12px 14px;min-width:0}
.solid{background:var(--surface-solid);border:1px solid var(--glass-border);border-radius:12px;padding:12px}
h2{margin:0 0 8px;font-size:11px;letter-spacing:.12em;text-transform:uppercase;color:var(--text-secondary);font-weight:600}
.note{color:var(--text-tertiary);font-size:11px;margin-top:4px}
.pos{color:var(--up)}.neg{color:var(--down)}.mix{color:var(--mix)}.unk{color:var(--unk)}
.tiles{display:grid;grid-template-columns:repeat(6,minmax(0,1fr));gap:8px}
.tile{background:var(--surface-solid);border:1px solid var(--glass-border);border-radius:10px;padding:8px;cursor:pointer;min-width:0}
.tile b{display:block;font-size:11px;letter-spacing:.08em}
.tile .px{font-size:16px;font-weight:650;margin:2px 0}
.tile .meta{color:var(--text-tertiary);font-size:10px}
.tile.closed{opacity:.72}
.tile.live{box-shadow:inset 0 0 0 1px rgba(95,191,150,.25)}
.tile.openflash{animation:illum .24s ease}
@keyframes illum{from{box-shadow:0 0 0 0 rgba(95,191,150,.5)}to{box-shadow:none}}
.bar{height:4px;background:#1b2430;border-radius:99px;overflow:hidden;margin:3px 0}
.bar>i{display:block;height:100%;background:rgba(106,168,212,.75)}
.orow{display:grid;grid-template-columns:18px 1fr;gap:6px;padding:7px 0;border-bottom:1px solid rgba(255,255,255,.04);cursor:pointer}
.orow b{font-size:13px}
.strip{display:flex;gap:12px;flex-wrap:wrap}
.chipk{color:var(--text-tertiary);font-size:10px;letter-spacing:.08em}
.chipk b{display:block;color:var(--text-primary);font-size:13px;font-weight:650}
.sessline{position:relative;height:28px;margin:8px 0 4px;background:#10161f;border-radius:99px;overflow:hidden}
.sessline span{position:absolute;top:0;bottom:0;display:flex;align-items:center;justify-content:center;font-size:9px;letter-spacing:.08em;color:var(--text-tertiary)}
.sessline .now{position:absolute;top:0;bottom:0;width:2px;background:var(--text-primary);opacity:.7}
#events{display:flex;flex-direction:column;min-height:0;padding:8px 10px 8px;
  border-left:1px solid var(--glass-border);background:rgba(10,14,20,.4);overflow:hidden}
#events h2{margin:0 0 6px;flex-shrink:0}
#tl{flex:1;min-height:0;overflow:auto;margin:0}
#tl div{padding:5px 0;border-bottom:1px solid rgba(255,255,255,.04);font-size:11px;color:var(--text-secondary)}
#tl .sev-warn{color:var(--warn)}#tl .sev-crit{color:var(--bad)}#tl .sev-state{color:var(--long)}
#foot{flex-shrink:0;display:flex;flex-wrap:wrap;align-items:center;gap:12px;padding:6px 14px;
  border-top:1px solid var(--glass-border);background:rgba(10,14,20,.92);
  color:var(--text-tertiary);font-size:11px;z-index:9}
#foot b{color:var(--text-primary);font-weight:600}
.hq{display:flex;gap:10px;margin-right:auto}
.hq span{letter-spacing:.06em}
.kv{display:grid;grid-template-columns:1fr auto;gap:3px 12px;color:var(--text-secondary);font-size:12px}
.kv b{color:var(--text-primary);font-weight:600}
.hero-px{font-size:32px;font-weight:650;letter-spacing:-.02em;margin:2px 0}
.hero-nm{font-size:22px;letter-spacing:.08em}
.cls{font-size:20px;font-weight:650;margin:4px 0 8px}
.cls.long{color:var(--long)}.cls.short{color:var(--short)}.cls.none{color:var(--text-tertiary)}
.why{white-space:pre-wrap;color:var(--text-primary);min-height:3.2em;margin-top:6px;font-size:12px}
canvas{width:100%;display:block;background:var(--surface-solid);border-radius:8px}
.tabs{display:flex;gap:6px;margin-bottom:8px}
.tab,.vp{border:1px solid var(--glass-border);background:transparent;color:var(--text-secondary);
  padding:3px 10px;font:11px inherit;cursor:pointer;border-radius:99px}
.tab.on,.vp.on{color:var(--text-primary);background:rgba(255,255,255,.05)}
table.heat{width:100%;border-collapse:collapse;font-size:12px}
.heat td,.heat th{padding:5px 8px;text-align:left;border-bottom:1px solid rgba(255,255,255,.05);color:var(--text-secondary)}
.heat th{font-size:10px;letter-spacing:.08em;text-transform:uppercase}
.grid2{display:grid;grid-template-columns:1fr 1fr;gap:12px}
.grid3{display:grid;grid-template-columns:1fr 1fr 1fr;gap:12px}
.axis{display:grid;grid-template-columns:72px 1fr 1fr;gap:6px;align-items:center;font-size:11px;color:var(--text-secondary);margin:4px 0}
.axis .pair{display:flex;height:8px;background:#1a2230;border-radius:99px;overflow:hidden}
.axis .l{background:var(--bid)}.axis .r{background:var(--ask)}
.meter{height:8px;background:#1a2230;position:relative;margin:3px 0 8px;border-radius:99px}
.meter>i{position:absolute;top:0;bottom:0;background:var(--long);width:8px;border-radius:99px}
.gates span{display:flex;justify-content:space-between;padding:3px 0;color:var(--text-secondary);font-size:12px}
.gates b{color:var(--text-primary)}
#pal{display:none;position:fixed;inset:18% 30%;z-index:20;background:var(--glass-bg);
  backdrop-filter:blur(20px);border:1px solid var(--glass-border);border-radius:14px;padding:12px;box-shadow:var(--glass-shadow)}
#pal.on{display:block}
#pal input{width:100%;background:#0d1219;border:1px solid var(--glass-border);color:var(--text-primary);padding:8px;border-radius:8px}
#pal button{display:block;width:100%;text-align:left;margin-top:6px;background:transparent;border:0;color:var(--text-secondary);padding:6px;cursor:pointer}
body.dense .glass,body.dense .solid,body.dense .tile{padding:8px}
body.dense .ov{gap:8px}
@media (max-width:1366px){
  #app{grid-template-columns:72px 1fr 180px}
  .tiles{grid-template-columns:repeat(4,minmax(0,1fr))}
  .hero{grid-column:1/13;grid-row:auto}
  .opp{grid-column:1/13;grid-row:auto}
}
@media (max-width:1100px){
  #app{grid-template-columns:1fr}
  #rail{flex-direction:row;overflow:auto}
  #events{display:none}
}
</style>
</head>
<body>
<header id="cmd">
  <div class="brand">
    <strong>AURUMFLOW</strong>
    <span class="sub">GLOBAL INTELLIGENCE</span>
    <span class="sys"><i class="dot ok"></i>SYSTEM ONLINE</span>
  </div>
  <div id="sessbar">
    <span>ASIA<b id="cmd-asia">—</b></span>
    <span>EUROPE<b id="cmd-eu">—</b></span>
    <span>US<b id="cmd-us">—</b></span>
  </div>
  <div class="saf">
    <span class="pill ok">DEMO</span>
    <span class="pill bad" title="LIVE IMPOSSIBLE / FAIL-CLOSED">LIVE LOCKED</span>
    <span id="booksync"><i class="dot"></i>BOOK —</span>
    <span id="kill"><i class="dot"></i>KILL OFF</span>
    <span id="clk"></span>
  </div>
</header>
<div id="app">
<nav id="rail" aria-label="Primary">
  <button type="button" class="on" data-view="overview">OVERVIEW</button>
  <button type="button" data-view="world">WORLD</button>
  <button type="button" data-view="markets">MARKETS</button>
  <button type="button" data-view="micro">MICRO</button>
  <button type="button" data-view="execution">EXECUTION</button>
  <button type="button" data-view="research">RESEARCH</button>
  <button type="button" id="dens">COMFORTABLE</button>
</nav>
<main id="stage">
<section id="view-overview" class="view on">
  <div class="ov">
    <section class="glass hero">
      <h2>What is the world doing?</h2>
      <div class="note">SESSION STRIP</div>
      <div class="sessline" id="sessline" aria-label="World session timeline"></div>
      <div class="note" id="sslead">TOKYO · HONG KONG · LONDON · NEW YORK</div>
      <div class="strip" id="wstrip" style="margin:10px 0"></div>
      <div class="tiles" id="constel"></div>
      <div class="grid2" style="margin-top:12px">
        <div class="solid">
          <h2>Capital flow evidence</h2>
          <div class="note">Official slow context. Not price. CAPITAL FLOW EVIDENCE.</div>
          <div class="kv" id="flowev"></div>
        </div>
        <div class="solid">
          <h2>PRICE LEADERSHIP</h2>
          <div class="note">Relative strength · CORRELATION ≠ CAUSATION</div>
          <div class="kv" id="rotkv"></div>
        </div>
      </div>
    </section>
    <aside class="glass opp">
      <h2>Where should I look?</h2>
      <div class="note">ATTENTION — not a recommendation</div>
      <div id="oppbox"></div>
      <div class="why" id="odetail">Select a market for why.</div>
    </aside>
    <section class="glass sessrot">
      <h2>Session / rotation</h2>
      <div class="kv" id="sesskv"></div>
    </section>
    <section class="glass mstate">
      <h2>Market state</h2>
      <div class="hero-nm" id="mhero-nm">GOLD</div>
      <div class="hero-px" id="mhero-px">—</div>
      <div class="note" id="mhero-st">MARKET CLOSED</div>
      <div class="kv" id="mherokv"></div>
    </section>
    <section class="glass dec">
      <h2>What AurumFlow sees · is it allowed to act?</h2>
      <div class="cls none" id="v1big">NO SETUP</div>
      <div class="kv">
        <span>Legacy</span><b id="leg">—</b>
        <span>Flow</span><b id="v1">NEUTRAL</b>
        <span>Absorption</span><b id="abs">UNAVAILABLE</b>
        <span>Data quality</span><b id="dqnow">—</b>
      </div>
      <div class="why" id="why">WAITING</div>
    </section>
  </div>
</section>
<section id="view-world" class="view">
  <div class="grid2">
    <section class="glass"><h2>World state</h2><div class="kv" id="gkv"></div><div class="note" id="truth">Official origin: WAITING</div></section>
    <section class="glass"><h2>Region cards</h2><div class="kv" id="rkv"></div></section>
  </div>
  <section class="glass" style="margin-top:12px">
    <h2>Asset-class heatmap</h2>
    <table class="heat" id="heat"><thead><tr><th></th><th>Momentum</th><th>Capital flow</th><th>Positioning</th><th>Macro</th><th>Risk</th><th>Opportunity</th></tr></thead><tbody></tbody></table>
  </section>
  <section class="glass" style="margin-top:12px">
    <h2>Official sources</h2>
    <div class="note">Source health · Official sources · CONTEXT CALENDAR</div>
    <table class="heat" id="srctab"><thead><tr><th>Source</th><th>Latest published</th><th>Latest observation</th><th>Retrieved</th><th>Freshness</th><th>Origin</th><th>Status</th></tr></thead><tbody></tbody></table>
    <div class="kv" id="calkv"></div>
    <div class="kv" id="instkv"></div>
  </section>
</section>
<section id="view-markets" class="view">
  <section class="glass">
    <h2>WORLD MARKET TAPE</h2>
    <table class="heat" id="tape"><thead><tr><th>Market</th><th>Tape</th><th>Session</th><th>Price</th><th>15m</th><th>1h</th><th>4h</th><th>Vol</th><th>Attention</th><th>Coverage</th><th>Eligibility</th></tr></thead><tbody></tbody></table>
  </section>
  <section class="glass" style="margin-top:12px">
    <h2>Comparison</h2>
    <div class="tabs">
      <button type="button" class="tab on" data-cmp="GOLD,SILVER">GOLD vs SILVER</button>
      <button type="button" class="tab" data-cmp="US100,US500">US100 vs US500</button>
      <button type="button" class="tab" data-cmp="DE40,UK100">DE40 vs UK100</button>
      <button type="button" class="tab" data-cmp="J225,CN50">J225 vs CN50</button>
    </div>
    <div class="kv" id="cmpkv"></div>
    <div class="note">Normalized paths. CORRELATION ≠ CAUSATION.</div>
  </section>
</section>
<section id="view-micro" class="view">
  <section class="glass" id="chartbox">
    <h2>BTC microstructure</h2>
    <div class="tabs">
      <button type="button" class="tab on" id="tab-btc">BTC</button>
      <button type="button" class="tab" id="tab-gold">GOLD DEMO</button>
      <span style="flex:1"></span>
      <button type="button" class="vp on" data-m="15">15m</button>
      <button type="button" class="vp" data-m="30">30m</button>
      <button type="button" class="vp" data-m="60">1h</button>
    </div>
    <canvas id="mkt" height="220" aria-label="Price chart"></canvas>
    <div class="note" id="mktnote">price + microprice · markers are runtime-observed only</div>
    <div class="kv" id="mpkv"></div>
  </section>
  <div class="grid3" style="margin-top:12px">
    <section class="solid"><h2>Pressure</h2><canvas id="prs" height="110"></canvas><div class="note">guides +15 / 0 / −15 frozen V1</div></section>
    <section class="solid"><h2>CVD</h2><canvas id="cvd" height="110"></canvas><div class="kv" id="cvdkv"></div></section>
    <section class="solid"><h2>Flow magnitude vs price efficiency</h2><canvas id="eff" height="110"></canvas><div class="note">IMPACT FAILURE as derived state</div></section>
  </div>
  <div class="grid2" style="margin-top:12px">
    <section class="glass"><h2>L2 book</h2><canvas id="book" height="240"></canvas><div class="kv" id="l2kv"></div></section>
    <section class="glass">
      <h2>Liquidity response</h2>
      <div id="liq"></div>
      <div class="note">Imbalance 1 / 5 / 10 / 20</div>
      <div id="imbs"></div>
      <h2 style="margin-top:12px">FLOW EXHAUSTION V1</h2>
      <div class="kv" id="exhkv"></div>
      <h2 style="margin-top:12px">Absorption</h2>
      <div class="kv" id="abskv"></div>
    </section>
  </div>
</section>
<section id="view-execution" class="view">
  <div class="grid2">
    <section class="glass">
      <h2>GOLD execution · DEMO · SHADOW</h2>
      <div id="gclosed">
        <div class="hero-nm">GOLD</div>
        <div class="hero-px">CLOSED</div>
        <div>Awaiting broker TRADEABLE status</div>
        <div class="note">NEXT · Runtime monetary calibration</div>
      </div>
      <div id="gopen" hidden>
        <div class="kv" id="gexec"></div>
        <div id="gpos" class="note"></div>
      </div>
    </section>
    <section class="glass">
      <h2>Eligibility / operational trust</h2>
      <div class="gates" id="gates"></div>
      <div class="note">Intelligence plane is separate from execution safety plane.</div>
    </section>
  </div>
  <section class="glass" style="margin-top:12px">
    <h2>Portfolio risk</h2>
    <div class="note">account_currency_risk · Aggregate risk cap: $300 DEMO</div>
    <div class="kv" id="riskkv"></div>
    <div id="riskbars"></div>
  </section>
</section>
<section id="view-research" class="view">
  <div class="grid2">
    <section class="glass">
      <h2>FLOW EXHAUSTION V1</h2>
      <div class="kv">
        <span>Spec</span><b>FLOW_EXHAUSTION_V1</b>
        <span>Discovery</span><b>n=56</b>
        <span>Holdout</span><b>n=50</b>
        <span>15m mean</span><b>+9.8 bp</b>
        <span>Hit</span><b>66.0%</b>
        <span>MFE/MAE</span><b>1.99</b>
        <span>Status</span><b>VALIDATED_EXTERNAL_HOLDOUT</b>
        <span>Execution</span><b>DISABLED</b>
      </div>
    </section>
    <section class="glass">
      <h2>Prospective collection</h2>
      <div class="note">NOT VALIDATION THRESHOLDS</div>
      <div class="kv" id="pros"></div>
      <canvas id="mile" height="36"></canvas>
    </section>
  </div>
  <section class="glass" style="margin-top:12px">
    <h2>Opportunity research · ATTENTION_SCORE_V1</h2>
    <div class="note">Prospective 15m · 1h · 4h · 1d. Do not tune.</div>
    <div class="kv" id="oppresearch"></div>
    <div class="note">L2 quality binds l2_proxy_quality. DEGRADED is amber, not failure.</div>
    <div class="kv" id="qkv"></div>
  </section>
</section>
</main>
<aside id="events">
  <h2>Activity</h2>
  <div id="tl"></div>
</aside>
</div>
<footer id="foot">
  <div class="hq">
    <span>TRADES <b id="h-tr">—</b></span>
    <span>BOOK <b id="h-bk">—</b></span>
    <span>L2 <b id="h-l2">—</b></span>
    <span>WORLD <b id="h-wd">—</b></span>
    <span>CAPITAL <b id="h-cp">PREOPEN</b></span>
  </div>
  <span>events/s <b id="f-rate">—</b></span>
  <span>trades <b id="f-tr">—</b></span>
  <span>depth deltas <b id="f-dlt">—</b></span>
  <span>drops <b id="f-drop">—</b></span>
  <span>gaps <b id="f-gap">—</b></span>
  <span>resyncs <b id="f-rs">—</b></span>
  <span>p50 <b id="f-p50">—</b></span>
  <span>p95 <b id="f-p95">—</b></span>
  <span>mem <b id="f-mem">—</b></span>
  <span>disk <b id="f-disk">—</b></span>
  <span>last event <b id="f-ev">—</b></span>
</footer>
<div id="pal">
  <input id="palq" placeholder="Open GOLD · Open BTC Micro · Show sources"/>
  <button type="button" data-go="overview">Overview</button>
  <button type="button" data-go="world">Show sources</button>
  <button type="button" data-go="markets" data-mkt="GOLD">Open GOLD</button>
  <button type="button" data-go="markets" data-mkt="US100">Open US100</button>
  <button type="button" data-go="micro">Open BTC Micro</button>
</div>
<script>
var MAX=3600, series=[], events=[], viewport=15, tab='btc', lastDraw=0, lastS=null, prevGold='', lastWorld=null, selMkt='GOLD', cmpPair=['GOLD','SILVER'];
function el(id){return document.getElementById(id)}
function miss(){return '—'}
function isNum(v){return typeof v==='number' && isFinite(v)}
function num(ok,v,n){if(!ok||!isNum(v)||(v===0&&n!=null&&idZero(v))) return miss(); if(!ok||!isNum(v)) return miss(); if(n==null) n=2; return group(v.toFixed(n))}
function idZero(v){return v===0}
function group(s){var n=String(s),p=n.split('.'),h=p[0],o='',i; for(i=0;i<h.length;i++){if(i&&(h.length-i)%3===0)o+=',';o+=h[i]} return p[1]!=null?o+'.'+p[1]:o}
function txt(v){if(v===undefined||v===null||v==='') return miss(); return String(v)}
function goldMS(s){return String((s&&s.market_status)||'').toUpperCase()}
function goldOpen(s){return goldMS(s)==='TRADEABLE'}
function l2ok(s){return !!(s&&s.l2_quotes_ok&&isNum(s.l2_spread))}
function legacy(s){
  if(s.last_legacy_direction>0) return 'LONG';
  if(s.last_legacy_direction<0) return 'SHORT';
  if(s.last_strategy==='LONG'||s.last_strategy==='SHORT') return s.last_strategy;
  return 'NONE';
}
function v1rail(c){
  c=String(c||'').toUpperCase();
  if(c.indexOf('EXHAUSTION')>=0) return 'EXHAUSTION';
  if(c.indexOf('CONTINUATION')>=0) return 'CONTINUATION';
  return 'NEUTRAL';
}
function tone(v){
  v=String(v||'').toUpperCase();
  if(v==='HEALTHY'||v==='SYNCED'||v==='OK'||v==='LIVE') return 'ok';
  if(v==='DEGRADED'||v==='PARTIAL'||v==='STALE'||v==='PREOPEN') return 'warn';
  if(v==='FAILED'||v==='UNSYNCED'||v==='UNUSABLE') return 'bad';
  return '';
}
function pushPt(s){
  var t=Date.now();
  if(s.last_update){var p=Date.parse(s.last_update); if(isFinite(p)) t=p}
  series.push({t:t, price:s.btc_price||0, micro:s.microprice||0, gold:s.gold_quotes_ok?s.gold_bid:0,
    pr:s.pressure||0, dp:s.directional_pressure||0, cvd:s.cvd||0, fe:s.flow_efficiency||0, imp:s.impact_failure||0});
  if(series.length>MAX) series=series.slice(series.length-MAX);
}
function fit(c){
  var d=c.getContext('2d'), r=window.devicePixelRatio||1, w=c.clientWidth, h=c.height;
  if(c.width!==w*r){c.width=w*r; c.style.height=h+'px'}
  d.setTransform(r,0,0,r,0,0);
  return {d:d,w:w,h:h};
}
function winSeries(){
  var cut=Date.now()-viewport*60000, out=[], i;
  for(i=0;i<series.length;i++) if(series[i].t>=cut) out.push(series[i]);
  return out;
}
function plot(id, pts, keys, cols, guides, ymin, ymax){
  var c=el(id); if(!c) return;
  var g=fit(c), d=g.d, w=g.w, h=g.h, pad=18;
  d.clearRect(0,0,w,h);
  if(!pts.length){d.fillStyle='#6d7889'; d.fillText('WAITING',10,20); return}
  var i,j,mn=ymin, mx=ymax;
  if(mn==null||mx==null){
    mn=Infinity; mx=-Infinity;
    for(i=0;i<pts.length;i++) for(j=0;j<keys.length;j++){
      var v=pts[i][keys[j]]; if(isNum(v)&&v!==0){if(v<mn)mn=v; if(v>mx)mx=v}
    }
    if(!isFinite(mn)){mn=-1; mx=1}
    if(mn===mx){mn-=1; mx+=1}
    var sp=(mx-mn)*0.08; mn-=sp; mx+=sp;
  }
  var t0=pts[0].t, t1=pts[pts.length-1].t; if(t1<=t0) t1=t0+1;
  function X(t){return pad+(t-t0)/(t1-t0)*(w-pad*2)}
  function Y(v){return h-pad-(v-mn)/(mx-mn)*(h-pad*2)}
  if(guides){
    d.fillStyle='rgba(212,177,90,.06)';
    d.fillRect(pad, Y(Math.max(mx,40)), w-pad*2, Y(15)-Y(Math.max(mx,40)));
    d.fillRect(pad, Y(-15), w-pad*2, Y(Math.min(mn,-40))-Y(-15));
    d.setLineDash([3,4]); d.strokeStyle='#2c3646';
    for(i=0;i<guides.length;i++){d.beginPath(); d.moveTo(pad,Y(guides[i])); d.lineTo(w-pad,Y(guides[i])); d.stroke()}
    d.setLineDash([]);
    d.fillStyle='#6d7889';
    for(i=0;i<guides.length;i++) d.fillText(String(guides[i]), 4, Y(guides[i])+3);
  } else {
    d.strokeStyle='#1b2430'; d.beginPath(); d.moveTo(pad,h-pad); d.lineTo(w-pad,h-pad); d.stroke();
  }
  for(j=0;j<keys.length;j++){
    d.strokeStyle=cols[j]||'#6ea8fe'; d.lineWidth=1.4; d.beginPath();
    var started=false;
    for(i=0;i<pts.length;i++){
      var v=pts[i][keys[j]]; if(!isNum(v)||v===0&&keys[j]==='gold') continue;
      if(!started){d.moveTo(X(pts[i].t),Y(v)); started=true} else d.lineTo(X(pts[i].t),Y(v));
    }
    d.stroke(); d.lineWidth=1;
  }
  if(id==='mkt'){
    var last=pts[pts.length-1], pk=keys[0], lastv=last[pk];
    if(isNum(lastv)&&lastv!==0){
      d.fillStyle='#e8edf5'; d.fillText(group(lastv.toFixed(1)), w-64, Y(lastv)-4);
    }
    for(i=0;i<events.length;i++){
      var ev=events[i], x=X(ev.t);
      if(x<pad||x>w-pad) continue;
      d.fillStyle=markCol(ev.kind);
      d.beginPath(); d.moveTo(x,8); d.lineTo(x-3,14); d.lineTo(x+3,14); d.fill();
    }
  }
}
function markCol(k){
  if(k==='legacy') return '#6aa8d4';
  if(k==='v1') return '#d4b15a';
  if(k==='absorption') return '#5fbf96';
  return '#6d7889';
}
function drawBook(s){
  var c=el('book'), g=fit(c), d=g.d, w=g.w, h=g.h;
  d.clearRect(0,0,w,h);
  if(!s.book_synced){d.fillStyle='#d4b15a'; d.fillText('BOOK UNSYNCED',10,20); return}
  var bids=s.top_bids||[], asks=s.top_asks||[];
  var maxq=0,i;
  for(i=0;i<bids.length;i++) if(bids[i].qty>maxq) maxq=bids[i].qty;
  for(i=0;i<asks.length;i++) if(asks[i].qty>maxq) maxq=asks[i].qty;
  if(maxq<=0){d.fillStyle='#6d7889'; d.fillText('WAITING',10,20); return}
  var rows=Math.max(asks.length,1)+1+Math.max(bids.length,1);
  var rh=Math.max(10,(h-10)/rows), y=6, mid=s.l2_mid;
  function bar(side, lv, y0){
    var bw=(lv.qty/maxq)*(w-100);
    d.fillStyle=side==='a'?'#5a3030':'#1f4a38';
    if(side==='a') d.fillRect(w-bw-8,y0,bw,rh-2); else d.fillRect(8,y0,bw,rh-2);
    d.fillStyle='#e8edf5';
    d.fillText((lv.price||0).toFixed(2), side==='a'?8:w-86, y0+rh-3);
  }
  for(i=asks.length-1;i>=0;i--){bar('a',asks[i],y); y+=rh}
  d.fillStyle='#9aa6b8'; d.fillText('MID '+num(l2ok(s),mid,2), 8, y+rh-3); y+=rh;
  for(i=0;i<bids.length;i++){bar('b',bids[i],y); y+=rh}
}
function meterHTML(v){
  var x=50+Math.max(-1,Math.min(1,v||0))*50;
  return '<div class="meter"><i style="left:'+x+'%"></i></div>';
}
function liqHTML(s){
  function pair(a,b){
    var t=Math.abs(a)+Math.abs(b); if(t<=0) return '<div class="pair"><span class="l" style="width:50%"></span><span class="r" style="width:50%"></span></div>';
    var lp=100*Math.abs(a)/t;
    return '<div class="pair"><span class="l" style="width:'+lp+'%"></span><span class="r" style="width:'+(100-lp)+'%"></span></div>';
  }
  return '<div class="axis"><span>REFILL</span>'+pair(s.bid_replenishment,s.ask_replenishment)+'</div>'+
    '<div class="axis"><span>DEPLETE</span>'+pair(s.bid_depletion,s.ask_depletion)+'</div>'+
    '<div class="axis"><span>PERSIST</span>'+pair(s.bid_persistence,s.ask_persistence)+'</div>';
}
function kv(id, rows){
  var e=el(id); if(!e) return;
  e.innerHTML=rows.map(function(r){return '<span>'+r[0]+'</span><b>'+r[1]+'</b>'}).join('');
}
function clock(){
  var d=new Date();
  el('clk').textContent=d.toISOString().slice(11,19)+' UTC';
}
function paint(s){
  lastS=s;
  var ms=goldMS(s);
  if(ms && ms!==prevGold && prevGold && (ms==='TRADEABLE'||prevGold==='TRADEABLE')){
    var t=document.querySelector('[data-sym="GOLD"]');
    if(t){ t.classList.add('openflash'); setTimeout(function(){t.classList.remove('openflash')},240); }
    events.push({t:Date.now(), kind:'state', text:'GOLD MARKET OPEN', sev:'state'});
  }
  prevGold=ms;
  el('booksync').innerHTML='<i class="dot '+(s.book_synced?'ok':'bad')+'"></i>'+(s.book_synced?'BOOK SYNC':'BOOK UNSYNCED');
  el('kill').innerHTML='<i class="dot '+(s.kill_switch?'bad':'ok')+'"></i>'+(s.kill_switch?'KILL ON':'KILL OFF');
  var lg=legacy(s); el('leg').textContent=lg==='NONE'?'—':lg;
  var v=v1rail(s.last_v1_classification);
  el('v1').textContent=v;
  el('v1big').textContent=lg==='NONE'&&v==='NEUTRAL'?'NO SETUP':v;
  el('v1big').className='cls '+(lg==='LONG'?'long':lg==='SHORT'?'short':'none');
  el('abs').textContent=txt(s.absorption_status)==='—'?'UNAVAILABLE':s.absorption_status;
  var why=(s.decision_why||'WAITING').split('\n').filter(Boolean).slice(0,5).join('\n');
  el('why').textContent=why;
  el('dqnow').textContent=s.l2_proxy_quality||'WAITING';
  var closed=!goldOpen(s);
  el('gclosed').hidden=!closed;
  el('gopen').hidden=closed;
  if(!closed){
    kv('gexec',[
      ['Market', 'TRADEABLE'],
      ['Bid', num(s.gold_quotes_ok,s.gold_bid,2)],
      ['Ask', num(s.gold_quotes_ok,s.gold_ask,2)],
      ['Spread', num(s.gold_quotes_ok,s.gold_spread,3)],
      ['Validation', txt(s.gold_validation)],
      ['Legacy', lg],
      ['Open', s.positions_known?String(s.open_positions):miss()],
      ['Entry', s.position_open?num(true,s.gold_entry,2):miss()],
      ['SL', s.position_open?num(true,s.gold_sl,2):miss()],
      ['TP', s.position_open?num(true,s.gold_tp,2):miss()],
      ['uPnL', s.position_open?num(true,s.gold_upnl,2):miss()]
    ]);
    el('gpos').textContent=posVis(s);
  } else {
    el('gclosed').innerHTML='<div class="hero-nm">GOLD</div><div class="hero-px">'+(ms||'CLOSED')+'</div>'+
      '<div>Awaiting broker TRADEABLE status</div><div class="note">NEXT · Runtime monetary calibration</div>';
  }
  kv('cvdkv',[
    ['CVD', num(isNum(s.cvd)&&s.cvd!==0,s.cvd,2)],
    ['Vel 1s', num(isNum(s.flow_velocity_1s)&&s.flow_velocity_1s!==0,s.flow_velocity_1s,2)]
  ]);
  kv('l2kv',[
    ['Spread', num(l2ok(s), s.l2_spread, 3)],
    ['Microprice', num(s.l2_quotes_ok&&isNum(s.microprice), s.microprice, 2)],
    ['Provider', txt(s.l2_provider)],
    ['Age', s.book_age_ms==null?miss():(s.book_age_ms+' ms')]
  ]);
  kv('mpkv',[
    ['MID', num(l2ok(s), s.l2_mid, 2)],
    ['MICRO', num(s.l2_quotes_ok&&isNum(s.microprice), s.microprice, 2)],
    ['Δ', (s.l2_quotes_ok&&isNum(s.microprice)&&isNum(s.l2_mid))?((s.microprice-s.l2_mid)>=0?'+':'')+ (s.microprice-s.l2_mid).toFixed(2):miss()]
  ]);
  el('imbs').innerHTML='1'+meterHTML(s.imbalance_1)+'5'+meterHTML(s.imbalance_5)+'10'+meterHTML(s.imbalance_10)+'20'+meterHTML(s.imbalance_20);
  el('liq').innerHTML=liqHTML(s);
  kv('exhkv',[
    ['Status', v],
    ['Directional pressure', num(isNum(s.directional_pressure),s.directional_pressure,1)],
    ['Note', 'No execution implication']
  ]);
  kv('abskv',[
    ['State', txt(s.absorption_status)],
    ['Supporting fill', num(isNum(s.supporting_replenishment),s.supporting_replenishment,2)],
    ['Opposing delete', num(isNum(s.opposing_depletion),s.opposing_depletion,2)],
    ['Persistence', num(isNum(s.supporting_persistence),s.supporting_persistence,2)]
  ]);
  var exh=s.prospective_exhaustion||0;
  kv('pros',[
    ['EXHAUSTION', exh+' / 25 · '+exh+' / 50 · '+exh+' / 100 · '+exh+' / 200'],
    ['Continuation', String(s.prospective_continuation||0)],
    ['Mature 15m', String(s.mature_15m||0)],
    ['Mature 1h', String(s.mature_1h||0)]
  ]);
  drawMile(exh);
  el('f-rate').textContent=num(isNum(s.event_rate)&&s.event_rate>0,s.event_rate,2);
  el('f-tr').textContent=s.trades>0?String(s.trades):miss();
  el('f-dlt').textContent=s.depth_deltas>0?String(s.depth_deltas):miss();
  el('f-drop').textContent=String(s.dropped_events||0);
  el('f-gap').textContent=String(s.book_gaps||0);
  el('f-rs').textContent=String(s.resyncs||0);
  el('f-p50').textContent=num(isNum(s.latency_p50_ms)&&s.latency_p50_ms>0,s.latency_p50_ms,1);
  el('f-p95').textContent=num(isNum(s.latency_p95_ms)&&s.latency_p95_ms>0,s.latency_p95_ms,1);
  el('f-mem').textContent=num(isNum(s.peak_mem_mb)&&s.peak_mem_mb>0,s.peak_mem_mb,1);
  el('f-disk').textContent=num(isNum(s.collector_disk_mb)&&s.collector_disk_mb>0,s.collector_disk_mb,1);
  el('f-ev').textContent=txt(s.last_event);
  el('h-tr').textContent=s.event_rate>0?'HEALTHY':'WAITING';
  el('h-bk').textContent=s.book_synced?'SYNCED':'UNSYNCED';
  el('h-l2').textContent=s.l2_proxy_quality||'WAITING';
  el('h-tr').className=tone(el('h-tr').textContent);
  el('h-bk').className=tone(el('h-bk').textContent);
  el('h-l2').className=tone(s.l2_proxy_quality);
  paintGates(s);
}
function paintGates(s){
  var rows=[
    ['Identity', true],
    ['Market data', !!s.gold_quotes_ok||goldMS(s)!==''],
    ['History', false],
    ['Strategy', false],
    ['Monetary', false],
    ['Lifecycle', false],
    ['Risk', true]
  ];
  el('gates').innerHTML=rows.map(function(r){return '<span>'+r[0]+'<b>'+(r[1]?'✓':'○')+'</b></span>'}).join('');
}
function posVis(s){
  if(!s.position_open) return '';
  var cur=s.gold_quotes_ok?s.gold_bid:s.gold_entry;
  return 'ENTRY '+num(true,s.gold_entry,2)+'   CURRENT '+num(isNum(cur),cur,2)+'   SL '+num(true,s.gold_sl,2)+'   TP '+num(true,s.gold_tp,2);
}
function drawMile(n){
  var c=el('mile'); if(!c) return;
  var g=fit(c), d=g.d, w=g.w, h=g.h, ms=[25,50,100,200], i;
  d.clearRect(0,0,w,h);
  d.strokeStyle='#243044'; d.beginPath(); d.moveTo(8,h/2); d.lineTo(w-8,h/2); d.stroke();
  for(i=0;i<ms.length;i++){
    var x=8+(ms[i]/200)*(w-16);
    d.fillStyle=n>=ms[i]?'#5fbf96':'#6d7889';
    d.fillRect(x-1,8,2,h-16);
    d.fillText(String(ms[i]), x-8, h-2);
  }
}
function drawAll(){
  var now=Date.now();
  if(now-lastDraw<500) return;
  lastDraw=now;
  var pts=winSeries();
  if(tab==='gold'){
    plot('mkt', pts, ['gold'], ['#d4b15a'], null, null, null);
    el('mktnote').textContent=goldOpen(lastS||{})?'GOLD DEMO · price · entry / SL / TP when open':'GOLD CLOSED · no quote series';
  } else {
    plot('mkt', pts, ['price','micro'], ['#e8edf5','#6aa8d4'], null, null, null);
    el('mktnote').textContent='BTC price + microprice · glyphs only';
  }
  plot('prs', pts, ['pr','dp'], ['#e8edf5','#d4b15a'], [15,0,-15], -40, 40);
  plot('cvd', pts, ['cvd'], ['#6aa8d4'], null, null, null);
  plot('eff', pts, ['fe','imp'], ['#5fbf96','#c98a72'], null, null, null);
  if(lastS) drawBook(lastS);
}
function sevClass(e){
  var t=String((e&& (e.sev||e.kind||e.text))||'').toUpperCase();
  if(t.indexOf('CRIT')>=0||t.indexOf('UNSYNC')>=0) return 'sev-crit';
  if(t.indexOf('WARN')>=0||t.indexOf('DEGRAD')>=0) return 'sev-warn';
  if(t.indexOf('STATE')>=0||t.indexOf('OPEN')>=0) return 'sev-state';
  return '';
}
function onEvt(list){
  events=list||[];
  if(events.length>200) events=events.slice(events.length-200);
  var box=el('tl'), i, html='';
  for(i=events.length-1;i>=0;i--){
    var e=events[i], t=new Date(e.t).toISOString().slice(11,19);
    html+='<div class="'+sevClass(e)+'">'+t+'  '+(e.text||e.kind)+'</div>';
  }
  box.innerHTML=html||'<div>WAITING</div>';
}
el('tab-btc').onclick=function(){tab='btc'; el('tab-btc').className='tab on'; el('tab-gold').className='tab'; lastDraw=0; drawAll()};
el('tab-gold').onclick=function(){tab='gold'; el('tab-gold').className='tab on'; el('tab-btc').className='tab'; lastDraw=0; drawAll()};
var vps=document.querySelectorAll('.vp');
for(var i=0;i<vps.length;i++) (function(b){
  b.onclick=function(){viewport=+b.getAttribute('data-m'); for(var j=0;j<vps.length;j++) vps[j].className='vp'; b.className='vp on'; lastDraw=0; drawAll()}
})(vps[i]);
clock(); setInterval(clock,1000); setInterval(drawAll,750);
fetch('/api/timeseries').then(function(r){return r.json()}).then(function(j){
  var ps=j.points||[];
  for(var i=0;i<ps.length && series.length<MAX;i++){
    series.push({t:ps[i].t, price:ps[i].price, micro:ps[i].microprice, gold:0,
      pr:ps[i].pressure, dp:ps[i].directional_pressure, cvd:ps[i].cvd,
      fe:ps[i].flow_efficiency, imp:ps[i].impact_failure});
  }
}).catch(function(){});
fetch('/api/events').then(function(r){return r.json()}).then(function(j){onEvt(j.events||[])}).catch(function(){});
fetch('/api/status').then(function(r){return r.json()}).then(function(s){pushPt(s); paint(s); drawAll()});
var es=null;
function connectSSE(){
  if(es) try{es.close()}catch(e){}
  es=new EventSource('/api/stream');
  es.onmessage=function(ev){
    try{
      var s=JSON.parse(ev.data);
      pushPt(s); paint(s);
      if(s.last_event){
        var last=events.length?events[events.length-1].text:'';
        if(s.last_event!==last){
          events.push({t:Date.now(), kind:'live', text:s.last_event});
          if(events.length>200) events=events.slice(events.length-200);
          onEvt(events);
        }
      }
    }catch(e){}
  };
  es.onerror=function(){ setTimeout(connectSSE, 2000); };
}
connectSSE();
function showView(name){
  var vs=document.querySelectorAll('.view'), bs=document.querySelectorAll('#rail [data-view]'), i;
  for(i=0;i<vs.length;i++) vs[i].className='view'+(vs[i].id==='view-'+name?' on':'');
  for(i=0;i<bs.length;i++) bs[i].className=bs[i].getAttribute('data-view')===name?'on':'';
  lastDraw=0; drawAll();
}
var navs=document.querySelectorAll('#rail [data-view]');
for(var ni=0;ni<navs.length;ni++) (function(b){ b.onclick=function(){ showView(b.getAttribute('data-view')); }; })(navs[ni]);
el('dens').onclick=function(){
  var d=document.body.classList.toggle('dense');
  el('dens').textContent=d?'DENSE':'COMFORTABLE';
};
document.addEventListener('keydown', function(e){
  if(e.target && e.target.id==='palq') return;
  var map={'1':'overview','2':'world','3':'markets','4':'micro','5':'execution','6':'research'};
  if(map[e.key]) showView(map[e.key]);
  if((e.ctrlKey||e.metaKey)&&e.key==='k'){ e.preventDefault(); el('pal').classList.toggle('on'); el('palq').focus(); }
});
var pbs=document.querySelectorAll('#pal [data-go]');
for(var pi=0;pi<pbs.length;pi++) (function(b){
  b.onclick=function(){
    showView(b.getAttribute('data-go'));
    if(b.getAttribute('data-mkt')) selectMarket(b.getAttribute('data-mkt'));
    el('pal').classList.remove('on');
  };
})(pbs[pi]);
function cell(v){
  v=String(v||'unknown').toLowerCase();
  var cls='unk';
  if(v==='inflow'||v==='up'||v==='expanding'||v==='risk_on'||v.indexOf('up')>=0) cls='pos';
  else if(v==='outflow'||v==='down'||v==='contracting'||v==='risk_off'||v.indexOf('down')>=0) cls='neg';
  else if(v==='mixed'||v==='transition') cls='mix';
  return '<span class="'+cls+'">'+v+'</span>';
}
function mget(w,id){return (w.Markets&&w.Markets[id])||{}}
function feat(w,id){return (w.Live&&w.Live.Features&&w.Live.Features[id])||{}}
function pct(v){if(!isNum(v)) return '—'; return (v*100).toFixed(2)+'%'}
function tapeOf(m){
  if(m.QuoteHealth==='STALE') return 'STALE';
  if(String(m.MarketStatus||'').toUpperCase()==='TRADEABLE') return 'LIVE';
  if((m.HistoryStatus||'')==='WARMING') return 'WARMING';
  return 'CLOSED';
}
function attnBar(v){ var n=isNum(v)?Math.max(0,Math.min(100,v)):0; return '<div class="bar"><i style="width:'+n+'%"></i></div>'; }
function paintWorld(w){
  lastWorld=w;
  if(!w||w.status==='WAITING'){ el('gkv').innerHTML='<span>WorldState</span><b>WAITING</b>'; return; }
  kv('gkv',[
    ['Liquidity', w.Liquidity?w.Liquidity.Class:'—'],
    ['Risk', w.Risk||'—'],
    ['USD', w.USD&&w.USD.USD?w.USD.USD:'—'],
    ['Session', w.Session||'—'],
    ['World hash', w.Hash||'—'],
    ['World valid', w.Valid||'—']
  ]);
  el('truth').textContent='Official origin: '+(w.Origins&&w.Origins.length?w.Origins.join(', '):'UNKNOWN')+' · '+(w.Valid||'WAITING');
  el('h-wd').textContent=w.Valid==='CURRENT_WORLD_STATE_VALID'?'PARTIAL':'WAITING';
  var R=w.Regions||{};
  el('cmd-asia').textContent=(R.JAPAN&&R.JAPAN.Equity)||'WAITING';
  el('cmd-eu').textContent=(R.EUROPE&&R.EUROPE.Equity)||'WAITING';
  el('cmd-us').textContent=(R.UNITED_STATES&&R.UNITED_STATES.Equity)||'WAITING';
  var rows=[];
  ['UNITED_STATES','EUROPE','JAPAN','CHINA_HONG_KONG'].forEach(function(k){
    var x=R[k]||{}; rows.push([k, (x.Equity||'—')+' · '+(x.Health||'UNKNOWN')]);
  });
  kv('rkv', rows);
  var A=w.AssetClasses||{};
  var tb=document.querySelector('#heat tbody'); tb.innerHTML='';
  ['EQUITIES','FIXED_INCOME','FX','PRECIOUS_METALS','ENERGY','CRYPTO'].forEach(function(k){
    var a=A[k]||{};
    var tr=document.createElement('tr');
    tr.innerHTML='<td>'+k+'</td><td>'+cell(a.Momentum)+'</td><td>'+cell(a.Flow)+'</td><td>'+cell(a.Positioning)+'</td><td>'+cell(a.Macro)+'</td><td>'+cell(a.Risk)+'</td><td>'+cell(a.Opportunity)+'</td>';
    tb.appendChild(tr);
  });
  paintOpp(w);
  paintTape(w);
  paintConstel(w);
  paintRotation(w);
  paintSessionStrip(w);
  paintCompare(w);
  paintHero(selMkt,w);
  var fl=w.Flows||[];
  kv('flowev', fl.length?fl.map(function(f){return [f.Region+' '+f.AssetClass, (f.Direction||'UNKNOWN')]}):[['Official flow','UNKNOWN']]);
  kv('sesskv',[
    ['Session', w.Session||'WAITING'],
    ['Liquidity', w.Liquidity?w.Liquidity.Class:'UNKNOWN'],
    ['Risk', w.Risk||'UNKNOWN']
  ]);
  kv('riskkv',[
    ['Unit', 'account_currency_risk'],
    ['Aggregate risk cap', '$300 DEMO'],
    ['Used', 'UNKNOWN until monetary calibration'],
    ['Remaining', '$300 DEMO']
  ]);
  el('riskbars').innerHTML='<div class="note">TOTAL — / $300 DEMO</div><div class="bar"><i style="width:0"></i></div>'+
    '<div class="note">PRECIOUS — / $150 · US EQUITY — / $150</div>';
  if(el('wstrip')){
    el('wstrip').innerHTML=
      chip('LIQUIDITY', w.Liquidity?w.Liquidity.Class:'UNKNOWN')+
      chip('RISK', w.Risk||'UNKNOWN')+
      chip('USD', w.USD&&w.USD.USD?w.USD.USD:'UNKNOWN')+
      chip('RATES', 'US —');
  }
}
function chip(k,v){return '<div class="chipk">'+k+'<b>'+v+'</b></div>'}
function paintOpp(w){
  var op=w.Opportunity||[], box=el('oppbox'); if(!box) return; box.innerHTML='';
  op.slice(0,8).forEach(function(m,i){
    var d=document.createElement('div'); d.className='orow';
    d.innerHTML='<div>'+(i+1)+'</div><div><b>'+m.Market+'</b>'+attnBar(m.Attention)+
      '<div class="meta">ATTENTION '+num(isNum(m.Attention),m.Attention,0)+' · Evidence '+num(isNum(m.Coverage),m.Coverage,0)+'% · '+(m.Eligibility||'ANALYSIS_ONLY')+'</div></div>';
    d.onclick=function(){ selectMarket(m.Market); showWhy(w,m); };
    box.appendChild(d);
  });
}
function showWhy(w,m){
  el('odetail').textContent='WHY ATTENTION '+m.Market+
    '\nMomentum '+(m.PriceTrend||'—')+' · vol '+(m.Volatility||'—')+
    '\nLegacy '+(m.Setup||'INSUFFICIENT_DATA')+
    '\nEligibility '+(m.Eligibility||'ANALYSIS_ONLY');
}
function paintConstel(w){
  var groups=[['US',['US100','US500','US30']],['EUROPE',['DE40','UK100']],['ASIA',['J225','CN50']],['METALS',['GOLD','SILVER']],['ENERGY',['OIL_CRUDE']],['CRYPTO',['BTC']]];
  var host=el('constel'); if(!host) return; host.innerHTML='';
  var opp={}; (w.Opportunity||[]).forEach(function(m){opp[m.Market]=m});
  groups.forEach(function(g){
    g[1].forEach(function(id){
      var m=Object.assign({},mget(w,id),opp[id]||{});
      var f=feat(w,id);
      var st=tapeOf(m);
      var d=document.createElement('div');
      d.className='tile'+(st==='LIVE'?' live':' closed');
      d.setAttribute('data-sym',id);
      var px=isNum(m.Mid)&&m.Mid>0?group(m.Mid.toFixed(id==='BTC'?1:2)):'—';
      var dir=String(m.PriceTrend||'');
      var arrow=dir.indexOf('UP')>=0?'▲':dir.indexOf('DOWN')>=0?'▼':'·';
      d.innerHTML='<b>'+id+'</b><div class="px">'+px+'</div><div class="meta">'+arrow+' '+st+' · 15m '+pct(f.Ret15m)+' · 1h '+pct(f.Ret1h)+'</div>'+attnBar(m.Attention);
      d.onclick=function(){ selectMarket(id); };
      host.appendChild(d);
    });
  });
}
function selectMarket(id){
  selMkt=id;
  if(lastWorld) paintHero(id,lastWorld);
}
function paintHero(id,w){
  var m=Object.assign({},mget(w,id));
  (w.Opportunity||[]).forEach(function(x){ if(x.Market===id) m=Object.assign(m,x); });
  el('mhero-nm').textContent=id;
  el('mhero-px').textContent=isNum(m.Mid)&&m.Mid>0?group(m.Mid.toFixed(2)):EmptyClosed(m);
  el('mhero-st').textContent=tapeOf(m)==='CLOSED'?'MARKET CLOSED':tapeOf(m);
  kv('mherokv',[
    ['Session', m.SessionLocal||m.MarketStatus||'UNKNOWN'],
    ['ATTENTION', num(isNum(m.Attention),m.Attention,0)],
    ['Coverage', num(isNum(m.Coverage),m.Coverage,1)+'%'],
    ['Eligibility', m.Eligibility||'DEMO_DISCOVERED']
  ]);
}
function EmptyClosed(m){ return tapeOf(m)==='CLOSED'?'—':'—'; }
function paintTape(w){
  var ids=['US100','US500','US30','GOLD','SILVER','OIL_CRUDE','DE40','UK100','J225','CN50','BTC'];
  var tb=document.querySelector('#tape tbody'); if(!tb) return; tb.innerHTML='';
  var opp={}; (w.Opportunity||[]).forEach(function(m){opp[m.Market]=m});
  ids.forEach(function(id){
    var m=Object.assign({},mget(w,id),opp[id]||{});
    var f=feat(w,id);
    var tr=document.createElement('tr');
    tr.innerHTML='<td>'+id+'</td><td>'+tapeOf(m)+'</td><td>'+(m.SessionLocal||m.MarketStatus||'UNKNOWN')+'</td><td>'+num(isNum(m.Mid)&&m.Mid>0,m.Mid,2)+'</td><td>'+pct(f.Ret15m)+'</td><td>'+pct(f.Ret1h)+'</td><td>'+pct(f.Ret4h)+'</td><td>'+(m.Volatility||f.VolState||'UNKNOWN')+'</td><td>'+num(isNum(m.Attention),m.Attention,1)+'</td><td>'+num(isNum(m.Coverage),m.Coverage,0)+'%</td><td>'+(m.Eligibility||'ANALYSIS_ONLY')+'</td>';
    tr.onclick=function(){ selectMarket(id); };
    tb.appendChild(tr);
  });
}
function paintRotation(w){
  var L=w.Live||{};
  kv('rotkv',[
    ['US equity', (L.Breadth&&L.Breadth.State)||'UNKNOWN'],
    ['Europe equity', (L.Europe&&L.Europe.Leadership)||'UNKNOWN'],
    ['Asia equity', (L.Asia&&L.Asia.Leadership)||'UNKNOWN'],
    ['Precious metals', (L.Precious&&L.Precious.GoldRS)||'UNKNOWN'],
    ['Energy', (L.Energy&&L.Energy.OilMomentum)||'UNKNOWN'],
    ['Crypto', (mget(w,'BTC').PriceTrend)||'UNKNOWN'],
    ['Label', 'PRICE LEADERSHIP']
  ]);
}
function paintSessionStrip(w){
  var t=new Date(w.AsOf||Date.now()), h=t.getUTCHours()+t.getUTCMinutes()/60;
  function st(a,b,pre){ if(h>=a&&h<b) return 'OPEN'; if(pre!=null&&((pre>a&&(h>=pre||h<a))||(pre<a&&h>=pre&&h<a))) return 'PREOPEN'; return 'CLOSED'; }
  var segs=[['TOKYO',0,8,23],['HONG KONG',1,8,0],['LONDON',7,16,6],['NEW YORK',13,21,12]];
  var host=el('sessline'); if(!host) return; host.innerHTML='';
  segs.forEach(function(s){
    var sp=document.createElement('span');
    var left=(s[1]/24)*100, width=((s[2]-s[1]+24)%24)/24*100;
    sp.style.left=left+'%'; sp.style.width=width+'%';
    var state=st(s[1],s[2],s[3]);
    sp.textContent=s[0]+' '+state;
    sp.style.background=state==='OPEN'?'rgba(95,191,150,.12)':state==='PREOPEN'?'rgba(212,177,90,.10)':'transparent';
    host.appendChild(sp);
  });
  var now=document.createElement('i'); now.className='now'; now.style.left=((h/24)*100)+'%'; host.appendChild(now);
  if(el('sslead')) el('sslead').textContent='TOKYO · HONG KONG · LONDON · NEW YORK · clock-descriptive · broker marketStatus authoritative';
}
document.querySelectorAll('[data-cmp]').forEach(function(b){
  b.onclick=function(){
    document.querySelectorAll('[data-cmp]').forEach(function(x){x.className='tab'});
    b.className='tab on';
    cmpPair=b.getAttribute('data-cmp').split(',');
    if(lastWorld) paintCompare(lastWorld);
  };
});
function paintCompare(w){
  if(!w) return;
  var a=cmpPair[0], b=cmpPair[1], fa=feat(w,a), fb=feat(w,b);
  kv('cmpkv',[[a+' 1h', pct(fa.Ret1h)],[b+' 1h', pct(fb.Ret1h)],['Label', 'normalized path · CORRELATION ≠ CAUSATION']]);
}
function paintSources(j){
  if(!j) return;
  var cat=j.catalog||[], tb=document.querySelector('#srctab tbody');
  if(tb){
    tb.innerHTML='';
    cat.forEach(function(s){
      var tr=document.createElement('tr');
      tr.innerHTML='<td>'+(s.provider||'—')+'</td><td>'+(s.latest_publication||'—')+'</td><td>'+(s.latest_observation||'—')+'</td><td>'+(s.retrieved_at||'—')+'</td><td>'+(s.freshness||'—')+'</td><td>'+(s.origin||'UNKNOWN')+'</td><td>'+(s.quality||s.last_error||'—')+'</td>';
      tb.appendChild(tr);
    });
  }
  var cal=j.calendar||[];
  kv('calkv', cal.map(function(c){return [c.Source||c.source, 'last '+(c.LastActual||c.last_actual||'—')+' · next '+(c.NextExpected||c.next_expected||'—')]}));
}
function loadWorld(){
  fetch('/api/world').then(function(r){return r.json()}).then(paintWorld).catch(function(){});
  fetch('/api/context/sources').then(function(r){return r.json()}).then(paintSources).catch(function(){});
  fetch('/api/institutional').then(function(r){return r.json()}).then(function(j){
    kv('instkv',[['Label', j.label||'DELAYED'],['Note', j.note||'']]);
  }).catch(function(){});
}
loadWorld(); setInterval(loadWorld, 15000);
</script>
</body>
</html>`
