package ops

const consoleHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8"/>
<title>AurumFlow Decision Console</title>
<style>
:root { --bg:#0b0e14; --panel:#121722; --line:#243044; --txt:#d7deea; --dim:#8b97a8; --ok:#3dd68c; --warn:#f5c542; --bad:#ff5d5d; --accent:#6ea8fe; }
*{box-sizing:border-box}
body{margin:0;background:var(--bg);color:var(--txt);font:13px/1.4 ui-sans-serif,system-ui,Segoe UI,sans-serif}
header{display:flex;gap:16px;align-items:center;padding:10px 16px;border-bottom:1px solid var(--line);background:#0e1320;position:sticky;top:0}
.badge{padding:3px 8px;border:1px solid var(--line);border-radius:4px;letter-spacing:.04em}
.ok{color:var(--ok);border-color:#245c40}
.bad{color:var(--bad);border-color:#6b2a2a}
.warn{color:var(--warn);border-color:#6b5720}
main{display:grid;grid-template-columns:1fr 1fr;gap:10px;padding:10px}
section{background:var(--panel);border:1px solid var(--line);border-radius:6px;padding:10px 12px;min-height:140px}
h2{margin:0 0 8px;font-size:12px;color:var(--dim);text-transform:uppercase;letter-spacing:.08em}
.kv{display:grid;grid-template-columns:1fr auto;gap:2px 12px}
.kv span{color:var(--dim)}
.why{white-space:pre-wrap;color:var(--txt)}
</style>
</head>
<body>
<header id="safety">
  <strong>AURUMFLOW</strong>
  <span class="badge ok" id="env">CAPITAL DEMO</span>
  <span class="badge bad" id="live">LIVE IMPOSSIBLE / FAIL-CLOSED</span>
  <span class="badge" id="exec">EXECUTION —</span>
  <span class="badge" id="kill">KILL —</span>
  <span class="badge" id="pos">POSITIONS —</span>
  <span class="badge" id="bal">BALANCE —</span>
  <span class="badge warn">RADAR SHADOW</span>
  <span class="badge warn">EXHAUSTION SHADOW</span>
  <span class="badge warn">ABSORPTION SHADOW</span>
</header>
<main>
<section><h2>GOLD</h2><div class="kv" id="gold"></div></section>
<section><h2>BTC Intelligence</h2><div class="kv" id="btc"></div></section>
<section><h2>L2 Book</h2><div class="kv" id="l2"></div></section>
<section><h2>Absorption</h2><div class="kv" id="abs"></div><div class="why" id="why"></div></section>
<section><h2>Provenance</h2><div class="kv" id="prov"></div></section>
<section><h2>Prospective</h2><div class="kv" id="pros"></div></section>
<section><h2>Research V1 (event study, not account return)</h2>
<div class="kv">
<span>Spec</span><b>FLOW_EXHAUSTION_V1</b>
<span>Discovery n</span><b>56</b>
<span>Holdout n</span><b>50</b>
<span>Holdout 15m mean</span><b>+9.8 bp</b>
<span>Holdout hit</span><b>66.0%</b>
<span>MFE/MAE</span><b>1.99</b>
<span>Status</span><b>VALIDATED_EXTERNAL_HOLDOUT</b>
<span>Live execution</span><b>DISABLED</b>
</div></section>
<section><h2>Data health</h2><div class="kv" id="health"></div></section>
</main>
<script>
function row(el, rows){el.innerHTML=rows.map(function(kv){var v=kv[1]; if(v===undefined||v===null||v==='') v='—'; return '<span>'+kv[0]+'</span><b>'+v+'</b>';}).join('')}
function paint(s){
  document.getElementById('env').textContent='CAPITAL '+(s.capital_environment||s.api_environment||'DEMO');
  document.getElementById('exec').textContent='EXECUTION '+(s.execution_mode||'SHADOW');
  document.getElementById('kill').textContent='KILL '+(s.kill_switch?'ON':'OFF');
  document.getElementById('pos').textContent='POSITIONS '+(s.open_positions??0);
  document.getElementById('bal').textContent='BALANCE '+(s.demo_balance??0);
  row(document.getElementById('gold'),[
    ['Market',s.market_status],['Bid',s.gold_bid],['Ask',s.gold_ask],['Spread',s.gold_spread],
    ['Validation',s.gold_validation],['Session',s.gold_session],['Legacy',s.last_strategy],
    ['Open',s.open_positions],['Entry',s.gold_entry],['SL',s.gold_sl],['TP',s.gold_tp],
    ['uPnL',s.gold_upnl],['Daily PnL',s.daily_pnl],['Daily DD',s.daily_dd_pct],['Trades today',s.trades_today]
  ]);
  row(document.getElementById('btc'),[
    ['Price',s.btc_price],['Agg buy',s.aggressive_buy_flow],['Agg sell',s.aggressive_sell_flow],
    ['CVD',s.cvd],['Vel 1s',s.flow_velocity_1s],['Vel 5s',s.flow_velocity_5s],['Vel 30s',s.flow_velocity_30s],
    ['Pressure',s.pressure],['DirectionalP',s.directional_pressure],['FlowEff',s.flow_efficiency],
    ['ImpactFailure',s.impact_failure],['V1',s.last_v1_classification],['Legacy dir',s.last_legacy_direction]
  ]);
  row(document.getElementById('l2'),[
    ['Provider',s.l2_provider||s.primary_l2_sensor],['Instrument',s.l2_instrument],
    ['Proxy quality',s.l2_proxy_quality],['Synced',s.book_synced],['Age ms',s.book_age_ms],
    ['Spread',s.gold_spread],['Microprice',s.microprice],
    ['Imb1',s.imbalance_1],['Imb5',s.imbalance_5],['Imb10',s.imbalance_10],['Imb20',s.imbalance_20],
    ['Bid replenish',s.bid_replenishment],['Ask replenish',s.ask_replenishment],
    ['Bid deplete',s.bid_depletion],['Ask deplete',s.ask_depletion],
    ['Bid persist',s.bid_persistence],['Ask persist',s.ask_persistence]
  ]);
  row(document.getElementById('abs'),[
    ['Status',s.absorption_status],['Evidence',s.absorption_evidence]
  ]);
  document.getElementById('why').textContent=s.absorption_why||'';
  row(document.getElementById('prov'),[
    ['V1 source',s.v1_flow_provider],['L2 source',s.l2_provider],['Relation',s.l2_relation],
    ['Quality',s.l2_proxy_quality],['Event freshness',s.event_freshness],
    ['Book freshness',s.book_freshness],['Last update',s.last_update]
  ]);
  row(document.getElementById('pros'),[
    ['Legacy total',s.prospective_signals],['Exhaustion',s.prospective_exhaustion],
    ['Continuation',s.prospective_continuation],['Neutral',s.prospective_neutral],
    ['Mature 15m',s.mature_15m],['Mature 1h',s.mature_1h],
    ['Exh L2 unavailable',s.exh_l2_unavailable],['Exh supportive',s.exh_supportive]
  ]);
  row(document.getElementById('health'),[
    ['Events/s',s.event_rate],['Deltas',s.depth_deltas],['Snapshots',s.book_snapshots],
    ['Drops',s.dropped_events],['Reconnects',s.reconnects],['Gaps',s.book_gaps],
    ['Resyncs',s.resyncs],['p50',s.latency_p50_ms],['p95',s.latency_p95_ms],
    ['Disk MB',s.collector_disk_mb],['Error',s.last_error]
  ]);
}
const es=new EventSource('/api/stream');
es.onmessage=e=>{try{paint(JSON.parse(e.data))}catch(err){}};
fetch('/api/status').then(r=>r.json()).then(paint);
</script>
</body>
</html>`
