package ops

const consoleHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1"/>
<title>AURUMFLOW Observatory</title>
<style>
:root{
  --bg:#0c1016;--panel:#131922;--ink:#d8dee8;--dim:#8a93a3;--line:#243044;
  --ok:#3ecf8e;--warn:#e4c15a;--bad:#e36a6a;--long:#5aa7d4;--short:#c9846a;
  --ask:#8d4a4a;--bid:#2f6b52;--guide:#3a4456;
}
*{box-sizing:border-box}
html,body{margin:0;height:100%;background:var(--bg);color:var(--ink);
  font:12px/1.4 ui-sans-serif,system-ui,Segoe UI,sans-serif}
body{display:flex;flex-direction:column;min-height:100vh}
#top{display:flex;flex-wrap:wrap;align-items:center;gap:10px 14px;
  padding:7px 12px;border-bottom:1px solid var(--line);background:#0e141c;position:sticky;top:0;z-index:5}
#top strong{letter-spacing:.12em;font-size:12px}
.dot{width:7px;height:7px;border-radius:50%;display:inline-block;margin-right:5px;background:var(--dim)}
.dot.ok{background:var(--ok)}.dot.warn{background:var(--warn)}.dot.bad{background:var(--bad)}
.chip{color:var(--dim);white-space:nowrap}
.chip b{color:var(--ink);font-weight:600}
.chip.bad b{color:var(--bad)}.chip.ok b{color:var(--ok)}.chip.warn b{color:var(--warn)}
#work{flex:1;display:grid;grid-template-columns:minmax(0,1.7fr) minmax(280px,1fr);
  grid-template-rows:auto auto auto 1fr;gap:8px;padding:8px;min-height:0}
.box{background:var(--panel);border:1px solid var(--line);padding:8px 10px;min-width:0}
.box h2{margin:0 0 6px;font-size:10px;letter-spacing:.1em;text-transform:uppercase;color:var(--dim);font-weight:600}
.tabs{display:flex;gap:6px;margin-bottom:6px}
.tab,.vp{border:1px solid var(--line);background:transparent;color:var(--dim);padding:2px 8px;font:11px inherit;cursor:pointer}
.tab.on,.vp.on{color:var(--ink);border-color:#4a5a72}
canvas{width:100%;display:block;background:#0d1219}
.row2{display:grid;grid-template-columns:1fr 1fr 1fr;gap:8px}
.kv{display:grid;grid-template-columns:1fr auto;gap:2px 10px;color:var(--dim)}
.kv b{color:var(--ink);font-weight:600}
.rail{display:flex;flex-direction:column;gap:8px;grid-row:1 / span 4}
.cls{font-size:20px;letter-spacing:.04em;margin:2px 0 8px;font-weight:650}
.cls.long{color:var(--long)}.cls.short{color:var(--short)}.cls.none{color:var(--dim)}
.why{white-space:pre-wrap;color:var(--ink);min-height:4.5em;border-top:1px solid var(--line);padding-top:6px;margin-top:6px}
#tl{height:92px;overflow:auto;font-family:ui-monospace,Consolas,monospace;font-size:11px;color:var(--dim)}
#tl div{padding:1px 0;border-bottom:1px solid #1b2430}
#foot{display:flex;flex-wrap:wrap;gap:12px;padding:5px 12px;border-top:1px solid var(--line);
  background:#0e141c;color:var(--dim);font-size:11px}
#foot b{color:var(--ink);font-weight:600}
.depth-wrap{display:grid;grid-template-columns:1.2fr .8fr;gap:8px}
.meter{height:8px;background:#1a2230;position:relative;margin:3px 0 8px}
.meter>i{position:absolute;top:0;bottom:0;background:#3d5a46}
.meter.mid>i{left:50%;width:2px;background:var(--guide)}
.sym{display:flex;height:10px;background:#1a2230;margin:3px 0 8px}
.sym .l{background:#2f6b52}.sym .r{background:#8d4a4a}
.note{color:var(--dim);font-size:11px;margin-top:4px}
.warnbox{border-color:#5a4a20;color:var(--warn)}
#gclosed{line-height:1.6}
#gpos{font-family:ui-monospace,Consolas,monospace;font-size:11px;white-space:pre}
@media (max-width:1100px){
  #work{grid-template-columns:1fr;grid-template-rows:auto}
  .rail{grid-row:auto}
  .row2{grid-template-columns:1fr}
}
</style>
</head>
<body>
<header id="top">
  <strong>AURUMFLOW</strong>
  <span class="chip ok"><i class="dot ok"></i><b>CAPITAL DEMO</b></span>
  <span class="chip bad"><i class="dot bad"></i><b>LIVE IMPOSSIBLE / FAIL-CLOSED</b></span>
  <span class="chip" id="goldm"><i class="dot warn"></i><b>GOLD WAITING</b></span>
  <span class="chip" id="btcfeed"><i class="dot"></i><b>BTC —</b></span>
  <span class="chip" id="booksync"><i class="dot"></i><b>BOOK —</b></span>
  <span class="chip" id="exec"><i class="dot"></i><b>EXEC SHADOW</b></span>
  <span class="chip" id="kill"><i class="dot"></i><b>KILL OFF</b></span>
  <span class="chip" id="pos"><i class="dot"></i><b>POS —</b></span>
  <span class="chip warn"><i class="dot warn"></i><b>SHADOW</b></span>
  <span class="chip" id="clk" style="margin-left:auto"></span>
</header>
<div id="work">
  <section class="box" id="chartbox">
    <h2>Market</h2>
    <div class="tabs">
      <button type="button" class="tab on" id="tab-btc">BTC INTELLIGENCE</button>
      <button type="button" class="tab" id="tab-gold">GOLD EXECUTION</button>
      <span style="flex:1"></span>
      <button type="button" class="vp on" data-m="15">15m</button>
      <button type="button" class="vp" data-m="30">30m</button>
      <button type="button" class="vp" data-m="60">1h</button>
    </div>
    <canvas id="mkt" height="220"></canvas>
    <div class="note" id="mktnote">price + microprice · markers are runtime-observed only</div>
  </section>
  <aside class="rail">
    <section class="box">
      <h2>What AurumFlow sees now</h2>
      <div class="kv"><span>Legacy</span><b id="leg">NONE</b></div>
      <div class="cls none" id="v1big">NEUTRAL</div>
      <div class="kv">
        <span>V1</span><b id="v1">NEUTRAL</b>
        <span>Absorption</span><b id="abs">UNAVAILABLE</b>
      </div>
      <div class="why" id="why">WAITING</div>
    </section>
    <section class="box" id="qbox">
      <h2>Data quality</h2>
      <div class="kv" id="qkv"></div>
    </section>
    <section class="box" id="gbox">
      <h2>GOLD execution · DEMO</h2>
      <div id="gclosed">
        <div>GOLD</div>
        <div><b>CLOSED</b></div>
        <div>Awaiting broker TRADEABLE status</div>
        <div>Monetary validation: <b>BROKER_METADATA_ONLY</b></div>
        <div>Execution: <b>NOT STARTED</b></div>
        <div>Next gate: <b>RUNTIME CALIBRATION</b></div>
      </div>
      <div id="gopen" hidden>
        <div class="kv" id="gkv"></div>
        <div id="gpos"></div>
      </div>
    </section>
    <section class="box">
      <h2>Prospective collection</h2>
      <div class="kv" id="pros"></div>
      <div class="note">Milestones 25 / 50 / 100 / 200 are collection counts, not validation.</div>
      <canvas id="mile" height="36"></canvas>
    </section>
    <section class="box">
      <h2>Research · event study, not account return</h2>
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
  </aside>
  <section class="box">
    <h2>Pressure · CVD · flow efficiency</h2>
    <div class="row2">
      <div>
        <canvas id="prs" height="110"></canvas>
        <div class="note">guides +15 / 0 / −15 = frozen V1 |DirectionalPressure| threshold, not overbought/oversold</div>
      </div>
      <div>
        <canvas id="cvd" height="110"></canvas>
        <div class="kv" id="cvdkv"></div>
      </div>
      <div>
        <canvas id="eff" height="110"></canvas>
        <div class="note">aggression vs deteriorating efficiency</div>
      </div>
    </div>
  </section>
  <section class="box">
    <h2>L2 book</h2>
    <div class="depth-wrap">
      <canvas id="book" height="200"></canvas>
      <div>
        <div class="kv" id="l2kv"></div>
        <div class="note">Imbalance 1 / 5 / 10 / 20</div>
        <div id="imbs"></div>
        <div class="note">Liquidity response · bid left / ask right</div>
        <div id="liq"></div>
      </div>
    </div>
  </section>
  <section class="box" style="grid-column:1">
    <h2>Event timeline</h2>
    <div id="tl"></div>
  </section>
</div>
<footer id="foot">
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
<script>
var MAX=3600, series=[], events=[], viewport=15, tab='btc', lastDraw=0, lastS=null, prevGold='';
function el(id){return document.getElementById(id)}
function miss(){return '—'}
function isNum(v){return typeof v==='number' && isFinite(v)}
function num(ok,v,n){if(!ok||!isNum(v)) return miss(); if(n==null) n=2; return v.toFixed(n)}
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
function pushPt(s){
  var t=Date.now();
  if(s.last_update){var p=Date.parse(s.last_update); if(isFinite(p)) t=p}
  series.push({
    t:t, price:s.btc_price||0, micro:s.microprice||0, gold:s.gold_quotes_ok?s.gold_bid:0,
    pr:s.pressure||0, dp:s.directional_pressure||0, cvd:s.cvd||0,
    fe:s.flow_efficiency||0, imp:s.impact_failure||0
  });
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
  var g=fit(c), d=g.d, w=g.w, h=g.h, pad=16;
  d.clearRect(0,0,w,h);
  if(!pts.length){d.fillStyle='#8a93a3'; d.fillText('WAITING',10,20); return}
  var i,j,mn=ymin, mx=ymax;
  if(mn==null||mx==null){
    mn=Infinity; mx=-Infinity;
    for(i=0;i<pts.length;i++) for(j=0;j<keys.length;j++){
      var v=pts[i][keys[j]]; if(isNum(v)){if(v<mn)mn=v; if(v>mx)mx=v}
    }
    if(!isFinite(mn)){mn=-1; mx=1}
    if(mn===mx){mn-=1; mx+=1}
    var sp=(mx-mn)*0.08; mn-=sp; mx+=sp;
  }
  var t0=pts[0].t, t1=pts[pts.length-1].t; if(t1<=t0) t1=t0+1;
  function X(t){return pad+(t-t0)/(t1-t0)*(w-pad*2)}
  function Y(v){return h-pad-(v-mn)/(mx-mn)*(h-pad*2)}
  d.strokeStyle='#243044'; d.beginPath(); d.moveTo(pad,Y(0)); d.lineTo(w-pad,Y(0)); d.stroke();
  if(guides){
    d.setLineDash([3,4]); d.strokeStyle='#3a4456';
    for(i=0;i<guides.length;i++){d.beginPath(); d.moveTo(pad,Y(guides[i])); d.lineTo(w-pad,Y(guides[i])); d.stroke()}
    d.setLineDash([]);
    d.fillStyle='#8a93a3';
    for(i=0;i<guides.length;i++) d.fillText(String(guides[i]), 4, Y(guides[i])+3);
  }
  for(j=0;j<keys.length;j++){
    d.strokeStyle=cols[j]||'#6ea8fe'; d.beginPath();
    var started=false;
    for(i=0;i<pts.length;i++){
      var v=pts[i][keys[j]]; if(!isNum(v)) continue;
      if(!started){d.moveTo(X(pts[i].t),Y(v)); started=true} else d.lineTo(X(pts[i].t),Y(v));
    }
    d.stroke();
  }
  if(id==='mkt'){
    for(i=0;i<events.length;i++){
      var ev=events[i];
      var x=X(ev.t);
      if(x<pad||x>w-pad) continue;
      d.fillStyle=markCol(ev.kind);
      d.fillRect(x, pad, 1, h-pad*2);
    }
  }
}
function markCol(k){
  if(k==='legacy') return '#5aa7d4';
  if(k==='v1') return '#e4c15a';
  if(k==='absorption') return '#3ecf8e';
  return '#8a93a3';
}
function drawBook(s){
  var c=el('book'), g=fit(c), d=g.d, w=g.w, h=g.h;
  d.clearRect(0,0,w,h);
  if(!s.book_synced){d.fillStyle='#e4c15a'; d.fillText('BOOK UNSYNCED',10,20); return}
  var bids=s.top_bids||[], asks=s.top_asks||[];
  var maxq=0,i;
  for(i=0;i<bids.length;i++) if(bids[i].qty>maxq) maxq=bids[i].qty;
  for(i=0;i<asks.length;i++) if(asks[i].qty>maxq) maxq=asks[i].qty;
  if(maxq<=0){d.fillStyle='#8a93a3'; d.fillText('WAITING',10,20); return}
  var rows=Math.max(asks.length,1)+1+Math.max(bids.length,1);
  var rh=Math.max(8,(h-8)/rows), y=4, mid=s.l2_mid;
  function bar(side, lv, y0){
    var bw=(lv.qty/maxq)*(w-90);
    d.fillStyle=side==='a'?'#5a3030':'#1f4a38';
    d.fillRect(80,y0,bw,rh-1);
    d.fillStyle='#d8dee8';
    d.fillText((lv.price||0).toFixed(2)+'  '+num(true,lv.qty,3), 4, y0+rh-2);
  }
  for(i=asks.length-1;i>=0;i--){bar('a',asks[i],y); y+=rh}
  d.fillStyle='#8a93a3'; d.fillText('mid '+num(l2ok(s),mid,2), 4, y+rh-2); y+=rh;
  for(i=0;i<bids.length;i++){bar('b',bids[i],y); y+=rh}
}
function meterHTML(v){
  var x=50+Math.max(-1,Math.min(1,v||0))*50;
  return '<div class="meter"><i style="left:'+x+'%;width:8px;background:#5aa7d4"></i></div>';
}
function liqHTML(s){
  function pair(a,b){
    var t=Math.abs(a)+Math.abs(b); if(t<=0) return '<div class="sym"><span class="l" style="width:50%"></span><span class="r" style="width:50%"></span></div>';
    var lp=100*Math.abs(a)/t;
    return '<div class="sym"><span class="l" style="width:'+lp+'%"></span><span class="r" style="width:'+(100-lp)+'%"></span></div>';
  }
  return 'Replenish'+pair(s.bid_replenishment,s.ask_replenishment)+
    'Deplete'+pair(s.bid_depletion,s.ask_depletion)+
    'Persist'+pair(s.bid_persistence,s.ask_persistence);
}
function kv(id, rows){
  var e=el(id); if(!e) return;
  e.innerHTML=rows.map(function(r){return '<span>'+r[0]+'</span><b>'+r[1]+'</b>'}).join('');
}
function clock(){
  var d=new Date();
  el('clk').textContent=d.toISOString().slice(11,19)+' UTC   '+d.toLocaleTimeString();
}
function setChip(id, cls, label){
  var e=el(id); e.className='chip '+cls;
  e.innerHTML='<i class="dot '+cls+'"></i><b>'+label+'</b>';
}
function paint(s){
  lastS=s;
  var ms=goldMS(s);
  if(ms==='TRADEABLE') setChip('goldm','ok','GOLD TRADEABLE');
  else if(ms==='CLOSED') setChip('goldm','warn','GOLD CLOSED');
  else setChip('goldm','warn','GOLD WAITING');
  prevGold=ms;
  setChip('btcfeed', s.btc_price>0?'ok':'warn', s.btc_price>0?('BTC '+num(true,s.btc_price,1)):'BTC WAITING');
  setChip('booksync', s.book_synced?'ok':'bad', s.book_synced?'BOOK SYNC':'BOOK UNSYNCED');
  setChip('exec','warn','EXEC '+(s.execution_mode||'SHADOW'));
  setChip('kill', s.kill_switch?'bad':'ok', s.kill_switch?'KILL ON':'KILL OFF');
  setChip('pos', s.positions_known?'ok':'warn', s.positions_known?('POS '+s.open_positions):'POS —');
  var lg=legacy(s); el('leg').textContent=lg;
  var v=v1rail(s.last_v1_classification);
  el('v1').textContent=v; el('v1big').textContent=v;
  el('v1big').className='cls '+(lg==='LONG'?'long':lg==='SHORT'?'short':'none');
  el('abs').textContent=txt(s.absorption_status)==='—'?'UNAVAILABLE':s.absorption_status;
  el('why').textContent=s.decision_why||'WAITING';
  var q=s.l2_proxy_quality||'WAITING';
  var qcls=q==='DEGRADED'||q==='UNUSABLE'||!s.book_synced;
  el('qbox').className=qcls?'box warnbox':'box';
  kv('qkv',[
    ['Binance trades', s.event_freshness?'HEALTHY':'WAITING'],
    ['Book', s.book_synced?'SYNCED':'UNSYNCED'],
    ['L2 quality', txt(s.l2_proxy_quality)],
    ['Book age', s.book_age_ms==null?miss():(s.book_age_ms+' ms')],
    ['Latency p95', num(isNum(s.latency_p95_ms)&&s.latency_p95_ms>0,s.latency_p95_ms,1)+' ms']
  ]);
  var closed=!goldOpen(s);
  el('gclosed').hidden=!closed;
  el('gopen').hidden=closed;
  if(!closed){
    kv('gkv',[
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
      ['uPnL', s.position_open?num(true,s.gold_upnl,2):miss()],
      ['Daily PnL', s.position_open||s.positions_known?num(isNum(s.daily_pnl),s.daily_pnl,2):miss()],
      ['Daily DD', s.positions_known?num(isNum(s.daily_dd_pct),s.daily_dd_pct,2):miss()],
      ['Trades today', s.positions_known?String(s.trades_today||0):miss()]
    ]);
    el('gpos').textContent=posVis(s);
  } else {
    var val=s.gold_validation||'BROKER_METADATA_ONLY';
    el('gclosed').innerHTML='<div>GOLD</div><div><b>'+(ms||'CLOSED')+'</b></div>'+
      '<div>Awaiting broker TRADEABLE status</div>'+
      '<div>Monetary validation: <b>'+val+'</b></div>'+
      '<div>Execution: <b>NOT STARTED</b></div>'+
      '<div>Next gate: <b>RUNTIME CALIBRATION</b></div>';
  }
  kv('cvdkv',[
    ['CVD', num(isNum(s.cvd),s.cvd,2)],
    ['Agg long', num(isNum(s.aggressive_buy_flow),s.aggressive_buy_flow,3)],
    ['Agg short', num(isNum(s.aggressive_sell_flow),s.aggressive_sell_flow,3)],
    ['Vel 1s', num(isNum(s.flow_velocity_1s),s.flow_velocity_1s,2)],
    ['Vel 5s', num(isNum(s.flow_velocity_5s),s.flow_velocity_5s,2)],
    ['Vel 30s', num(isNum(s.flow_velocity_30s),s.flow_velocity_30s,3)]
  ]);
  kv('l2kv',[
    ['Spread', num(l2ok(s), s.l2_spread, 3)],
    ['Microprice', num(s.l2_quotes_ok&&isNum(s.microprice), s.microprice, 2)],
    ['Provider', txt(s.l2_provider)],
    ['Age', s.book_age_ms==null?miss():(s.book_age_ms+' ms')]
  ]);
  el('imbs').innerHTML='1'+meterHTML(s.imbalance_1)+'5'+meterHTML(s.imbalance_5)+'10'+meterHTML(s.imbalance_10)+'20'+meterHTML(s.imbalance_20);
  el('liq').innerHTML=liqHTML(s);
  var exh=s.prospective_exhaustion||0;
  kv('pros',[
    ['Legacy', String(s.prospective_signals||0)],
    ['Exhaustion', String(exh)],
    ['Continuation', String(s.prospective_continuation||0)],
    ['Neutral', String(s.prospective_neutral||0)],
    ['Mature 15m', String(s.mature_15m||0)],
    ['Mature 1h', String(s.mature_1h||0)],
    ['Exhaustion with valid L2', String((s.exh_supportive||0)+(s.exh_strongly_supportive||0))]
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
}
function posVis(s){
  if(!s.position_open) return '';
  var cur=s.gold_quotes_ok?s.gold_bid:s.gold_entry;
  function dist(px){
    if(!isNum(px)||!isNum(cur)||cur===0) return miss();
    var dpx=px-cur, pct=100*dpx/cur, r=miss();
    if(isNum(s.gold_entry)&&isNum(s.gold_sl)&&s.gold_entry!==s.gold_sl){
      r=((px-s.gold_entry)/Math.abs(s.gold_entry-s.gold_sl)).toFixed(2);
    }
    return num(true,px,2)+'   '+num(true,pct,2)+'%   R '+r;
  }
  return 'ENTRY    '+dist(s.gold_entry)+'\nCURRENT  '+dist(cur)+'\nTP       '+dist(s.gold_tp)+'\nSL       '+dist(s.gold_sl);
}
function drawMile(n){
  var c=el('mile'), g=fit(c), d=g.d, w=g.w, h=g.h, ms=[25,50,100,200], i;
  d.clearRect(0,0,w,h);
  d.strokeStyle='#243044'; d.beginPath(); d.moveTo(8,h/2); d.lineTo(w-8,h/2); d.stroke();
  for(i=0;i<ms.length;i++){
    var x=8+(ms[i]/200)*(w-16);
    d.fillStyle=n>=ms[i]?'#3ecf8e':'#8a93a3';
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
    plot('mkt', pts, ['gold'], ['#e4c15a'], null, null, null);
    el('mktnote').textContent=goldOpen(lastS||{})?'GOLD DEMO · price only after TRADEABLE':'GOLD CLOSED · no quote series';
  } else {
    plot('mkt', pts, ['price','micro'], ['#d8dee8','#5aa7d4'], null, null, null);
    el('mktnote').textContent='BTC price + microprice · runtime markers only';
  }
  plot('prs', pts, ['pr','dp'], ['#d8dee8','#e4c15a'], [15,0,-15], -40, 40);
  plot('cvd', pts, ['cvd'], ['#5aa7d4'], null, null, null);
  plot('eff', pts, ['fe','imp'], ['#3ecf8e','#c9846a'], null, null, null);
  if(lastS) drawBook(lastS);
}
function onEvt(list){
  events=list||[];
  if(events.length>200) events=events.slice(events.length-200);
  var box=el('tl'), i, html='';
  for(i=events.length-1;i>=0;i--){
    var e=events[i], t=new Date(e.t).toISOString().slice(11,19);
    html+='<div>'+t+'  '+e.kind+'  '+e.text+'</div>';
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
var es=new EventSource('/api/stream');
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
</script>
</body>
</html>`
