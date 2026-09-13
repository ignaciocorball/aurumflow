# Sunday GO: explicit GOLD DEMO start. Never starts automatically. Never touches LIVE.
# Does not restart BTC SHADOW on 8766.
$ErrorActionPreference = "Stop"
Set-Location (Split-Path -Parent $PSScriptRoot)
if (-not $PSScriptRoot) { Set-Location $PSScriptRoot }

function Fail([string]$msg) {
    Write-Host "STOP: $msg"
    exit 1
}

if (-not $env:AURUMFLOW_CONFIG) {
    $env:AURUMFLOW_CONFIG = "config/demo_config.json"
}
$cfgLeaf = Split-Path $env:AURUMFLOW_CONFIG -Leaf
if ($cfgLeaf -ne "demo_config.json") {
    Fail "AURUMFLOW_CONFIG must be config/demo_config.json (got $cfgLeaf). LIVE config.json is forbidden."
}
if (-not $env:AURUMFLOW_DEMO_API_KEY -or -not $env:AURUMFLOW_DEMO_IDENTIFIER -or -not $env:AURUMFLOW_DEMO_PASSWORD) {
    Fail "missing AURUMFLOW_DEMO_* (will not load LIVE secrets)"
}

function Invoke-Bot([string[]]$botArgs) {
    $prev = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    & go run ./cmd/bot @botArgs
    $code = $LASTEXITCODE
    $ErrorActionPreference = $prev
    if ($code -ne 0) { Fail "command failed: go run ./cmd/bot $($botArgs -join ' ')" }
}

function Get-BotText([string[]]$botArgs) {
    $prev = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    $raw = & go run ./cmd/bot @botArgs 2>&1 | ForEach-Object { "$_" } | Out-String
    $script:LastBotExit = $LASTEXITCODE
    $ErrorActionPreference = $prev
    return $raw
}

function Get-Preflight {
    $raw = Get-BotText @("--ops-preflight")
    Write-Host $raw
    if ($script:LastBotExit -ne 0) { Fail "ops-preflight exit $script:LastBotExit" }
    $clean = (($raw -split "`n") | Where-Object { $_ -notmatch '^\d{4}/\d{2}/\d{2}' }) -join "`n"
    $jsonStart = $clean.IndexOf("{")
    if ($jsonStart -lt 0) { Fail "ops-preflight produced no JSON" }
    $json = $clean.Substring($jsonStart)
    $end = $json.LastIndexOf("}")
    $obj = $json.Substring(0, $end + 1) | ConvertFrom-Json
    return $obj
}

function Require-Check($pf, [string]$name, [string[]]$ok) {
    $c = $pf.checks | Where-Object { $_.name -eq $name } | Select-Object -First 1
    if (-not $c) { Fail "missing preflight check $name" }
    if ($ok -notcontains $c.status) { Fail "$name is $($c.status): $($c.detail)" }
    return $c
}

Write-Host "=== 1 ops-preflight ==="
$pf = Get-Preflight
if ($pf.result -eq "BLOCKED") { Fail "ops-preflight BLOCKED" }
Require-Check $pf "demo_host" @("PASS")
Require-Check $pf "live" @("PASS")
$pos = Require-Check $pf "positions" @("PASS")
if ($pos.detail -ne "0") { Fail "open positions must be 0 (got $($pos.detail))" }

Write-Host "=== 2 prepare-instrument GOLD ==="
$prep = Get-BotText @("--prepare-instrument", "GOLD")
Write-Host $prep
if ($script:LastBotExit -ne 0) { Fail "prepare-instrument GOLD" }

Write-Host "=== 3 verify GOLD TRADEABLE ==="
if ($prep -notmatch "status=TRADEABLE" -and $prep -notmatch "market GOLD is TRADEABLE") {
    Fail "GOLD is not TRADEABLE"
}

Write-Host "=== 4 verify positions == 0 ==="
$pf = Get-Preflight
$pos = Require-Check $pf "positions" @("PASS")
if ($pos.detail -ne "0") { Fail "positions != 0 after prepare" }

Write-Host "=== 5 calibrate-instrument GOLD ==="
Invoke-Bot @("--calibrate-instrument", "GOLD")

Write-Host "=== 6-7 verify calibration closed and positions == 0 ==="
$pf = Get-Preflight
$pos = Require-Check $pf "positions" @("PASS")
if ($pos.detail -ne "0") { Fail "positions != 0 after calibrate" }

Write-Host "=== 8 verify GOLD monetary RUNTIME_VALIDATED ==="
$specPath = Join-Path (Get-Location) "journals\instruments\GOLD.json"
if (-not (Test-Path $specPath)) { Fail "missing $specPath" }
$spec = Get-Content $specPath -Raw | ConvertFrom-Json
if ($spec.ValidationStatus -ne "RUNTIME_VALIDATED") {
    Fail "GOLD monetary validation=$($spec.ValidationStatus) (require RUNTIME_VALIDATED)"
}

Write-Host "=== 9 ops-preflight again ==="
$pf = Get-Preflight
if ($pf.result -eq "BLOCKED") { Fail "ops-preflight BLOCKED after calibrate" }
Require-Check $pf "gold" @("PASS")
$gold = $pf.checks | Where-Object { $_.name -eq "gold" } | Select-Object -First 1
if ($gold.detail -notmatch "TRADEABLE") { Fail "GOLD no longer TRADEABLE" }

Write-Host "=== 10 final GO state ==="
Require-Check $pf "demo_host" @("PASS")
Require-Check $pf "live" @("PASS")
Require-Check $pf "journal" @("PASS")
$pos = Require-Check $pf "positions" @("PASS")
if ($pos.detail -ne "0") { Fail "positions != 0 at GO" }

Write-Host "=== 11 start demo-week :8765 (SHADOW on 8766 is left running) ==="
$env:AURUMFLOW_CONFIG = "config/demo_config.json"
Start-Process -FilePath "go" -ArgumentList @(
    "run", "./cmd/bot", "--demo-week", "--epic", "GOLD", "--status-addr", "127.0.0.1:8765"
) -WorkingDirectory (Get-Location) -WindowStyle Minimized
Write-Host "DEMO-WEEK launched on 127.0.0.1:8765 - Legacy only. Radar/Exhaustion/Absorption remain SHADOW."
exit 0
