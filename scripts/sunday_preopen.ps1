# Read-only Sunday pre-open supervisor. Never mutates the broker.
$ErrorActionPreference = "Continue"
$Root = Split-Path -Parent $PSScriptRoot
if (-not $PSScriptRoot) { $Root = Get-Location }
Set-Location $Root

$env:AURUMFLOW_CONFIG = "config/demo_config.json"
$Log = Join-Path $Root "journals\sunday-preopen.log"
New-Item -ItemType Directory -Force -Path (Join-Path $Root "journals") | Out-Null

# Session-only: keep this machine awake while the supervisor runs.
Add-Type -Namespace Native -Name Power -MemberDefinition @"
[System.Runtime.InteropServices.DllImport("kernel32.dll")]
public static extern uint SetThreadExecutionState(uint esFlags);
"@
[void][Native.Power]::SetThreadExecutionState(0x80000000 -bor 0x00000001 -bor 0x00000040)

function Write-Log([string]$msg) {
    $line = "{0:u} {1}" -f (Get-Date).ToUniversalTime(), $msg
    Add-Content -Path $Log -Value $line
    Write-Host $line
}

function Get-Json($url) {
    try {
        return Invoke-RestMethod -Uri $url -TimeoutSec 5
    } catch {
        return $null
    }
}

$prev = @{
    health = ""
    synced = ""
    gold = ""
    shadow = ""
    gold8765 = ""
    quality = ""
}
$tradeableAnnounced = $false

Write-Log "supervisor start poll=60s read-only"

while ($true) {
    $health = Get-Json "http://127.0.0.1:8766/healthz"
    $st = Get-Json "http://127.0.0.1:8766/status"
    $h = if ($health -and $health.ok) { "ok" } else { "down" }
    if ($h -ne $prev.health) { Write-Log "8766 health $h"; $prev.health = $h }

    $shadowAlive = $false
    Get-CimInstance Win32_Process -Filter "Name='go.exe' OR Name='bot.exe' OR Name='aurumflow.exe'" -ErrorAction SilentlyContinue | ForEach-Object {
        if ($_.CommandLine -match "shadow-runtime") { $shadowAlive = $true }
    }
    $ss = if ($shadowAlive) { "up" } else { "missing" }
    if ($ss -ne $prev.shadow) { Write-Log "SHADOW process $ss"; $prev.shadow = $ss }

    if ($st) {
        $syn = [string]$st.book_synced
        if ($syn -ne $prev.synced) { Write-Log "book_synced=$syn gaps=$($st.book_gaps) drops=$($st.dropped_events)"; $prev.synced = $syn }
        $q = [string]$st.l2_proxy_quality
        if ($q -and $q -ne $prev.quality) { Write-Log "L2 quality $q"; $prev.quality = $q }
        if (-not $st.book_synced) { Write-Log "WARN book unsynced Absorption=UNAVAILABLE" }
        if ($st.dropped_events -gt 0) { Write-Log "WARN drops=$($st.dropped_events)" }
        if ($st.book_gaps -gt 5) { Write-Log "WARN gaps=$($st.book_gaps)" }
    } elseif ($h -eq "down") {
        Write-Log "ERROR console/status unreachable on 8766"
    }

    $pros = Join-Path $Root "research\prospective\FLOW_EXHAUSTION_V1"
    if (-not (Test-Path $pros)) { Write-Log "WARN prospective dir missing" }

    $drive = (Get-Item $Root).PSDrive.Name
    $free = (Get-PSDrive $drive).Free
    if ($free -lt 256MB) { Write-Log "WARN low disk free=$([int]($free/1MB))MB" }

    $g5 = Get-Json "http://127.0.0.1:8765/healthz"
    $g5s = if ($g5 -and $g5.ok) { "ok" } else { "down" }
    if ($g5s -ne $prev.gold8765) { Write-Log "8765 health $g5s"; $prev.gold8765 = $g5s }

    # Periodic read-only preflight for GOLD market-status only (every 5th tick).
    if (-not (Get-Variable tick -ErrorAction SilentlyContinue)) { $tick = 0 }
    $tick++
    if ($tick % 5 -eq 1) {
        $out = & go run ./cmd/bot --ops-preflight 2>&1 | Out-String
        $gold = ""
        if ($out -match "GOLD market_status=(\w+)") { $gold = $Matches[1] }
        elseif ($out -match '"detail": "CLOSED"') { $gold = "CLOSED" }
        elseif ($out -match '"detail": "TRADEABLE"') { $gold = "TRADEABLE" }
        if ($gold -and $gold -ne $prev.gold) {
            Write-Log "GOLD market-status $gold"
            $prev.gold = $gold
        }
        if ($gold -eq "TRADEABLE" -and -not $tradeableAnnounced) {
            $tradeableAnnounced = $true
            Write-Host "================================="
            Write-Host "GOLD IS TRADEABLE"
            Write-Host "READY FOR DEMO CALIBRATION"
            Write-Host "================================="
            Write-Log "GOLD IS TRADEABLE READY FOR DEMO CALIBRATION (no auto start)"
        }
        if ($out -match "ops-preflight BLOCKED") {
            Write-Log "ERROR ops-preflight BLOCKED"
        }
    }

    Start-Sleep -Seconds 60
}
