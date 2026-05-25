# End-to-end test: fires a real campaign through the full stack and
# checks that the campaign-worker logs show ranking output.
#
# Prerequisites: docker compose up -d is already running.
# Usage: .\scripts\e2e_test.ps1

param(
    [string]$GatewayBase = "http://localhost:8081",
    [string]$RankerBase  = "http://localhost:8082"
)

$pass = 0
$fail = 0

function Write-Pass($msg) { Write-Host "  PASS  $msg" -ForegroundColor Green;  $script:pass++ }
function Write-Fail($msg) { Write-Host "  FAIL  $msg" -ForegroundColor Red;    $script:fail++ }

Write-Host "`n=== Phase 6 End-to-End Test ===`n" -ForegroundColor Cyan

# ── 1. Ranking engine health ──────────────────────────────────────────────────
Write-Host "-- Step 1: ranking engine health" -ForegroundColor Yellow
try {
    $h = Invoke-RestMethod "$RankerBase/health"
    if ($h.status -eq "ok") { Write-Pass "ranking-engine is up" }
    else                     { Write-Fail "ranking-engine /health bad response" }
} catch {
    Write-Fail "ranking-engine is NOT reachable: $($_.Exception.Message)"
    Write-Host "  Hint: run 'docker compose up -d' or '.\build\ranking-server.exe'" -ForegroundColor DarkYellow
}

# ── 2. API gateway health ─────────────────────────────────────────────────────
Write-Host "`n-- Step 2: API gateway reachable" -ForegroundColor Yellow
try {
    Invoke-RestMethod "$GatewayBase/auth/register" -Method POST `
        -ContentType "application/json" `
        -Body '{"email":"e2e@test.com","password":"test1234"}' `
        -ErrorAction SilentlyContinue | Out-Null
    Write-Pass "API gateway is reachable"
} catch {
    if ($_.Exception.Response.StatusCode -in 400,409) {
        Write-Pass "API gateway is reachable (user may already exist)"
    } else {
        Write-Fail "API gateway not reachable: $($_.Exception.Message)"
    }
}

# ── 3. Register + login ───────────────────────────────────────────────────────
Write-Host "`n-- Step 3: auth flow" -ForegroundColor Yellow
$token = $null
try {
    Invoke-RestMethod "$GatewayBase/auth/register" -Method POST `
        -ContentType "application/json" `
        -Body '{"email":"e2e_phase6@test.com","password":"test1234"}' `
        -ErrorAction SilentlyContinue | Out-Null
} catch {}

try {
    $login = Invoke-RestMethod "$GatewayBase/auth/login" -Method POST `
        -ContentType "application/json" `
        -Body '{"email":"e2e_phase6@test.com","password":"test1234"}'
    $token = $login.token
    if ($token) { Write-Pass "login returned JWT token" }
    else        { Write-Fail "login response missing token field" }
} catch {
    Write-Fail "login failed: $($_.Exception.Message)"
}

# ── 4. Create a campaign (fires Kafka job) ────────────────────────────────────
Write-Host "`n-- Step 4: create campaign (triggers full pipeline)" -ForegroundColor Yellow
$campaignId = $null
if ($token) {
    try {
        $campaign = Invoke-RestMethod "$GatewayBase/api/campaigns" -Method POST `
            -ContentType "application/json" `
            -Headers @{ Authorization = "Bearer $token" } `
            -Body '{"industry":"tech","product":"E2E Test Product","website":""}'
        $campaignId = $campaign.campaign_id ?? $campaign.id
        Write-Pass "campaign created (id=$campaignId)"
    } catch {
        Write-Fail "campaign creation failed: $($_.Exception.Message)"
    }
} else {
    Write-Host "  SKIP  campaign creation (no token)" -ForegroundColor DarkYellow
}

# ── 5. Wait and check campaign-worker logs for ranking output ─────────────────
Write-Host "`n-- Step 5: verify campaign-worker ran the ranker" -ForegroundColor Yellow
Write-Host "  Waiting 15s for crawler + ranker to finish..." -ForegroundColor DarkGray
Start-Sleep -Seconds 15

try {
    $logs = docker logs campaign-worker 2>&1
    if ($logs -match "Ranked http") {
        Write-Pass "campaign-worker logged ranking output"
    } elseif ($logs -match "Processing campaign") {
        Write-Host "  WARN  campaign processed but no ranking log found yet" -ForegroundColor DarkYellow
        Write-Host "  Tip: run 'docker logs campaign-worker' manually to inspect" -ForegroundColor DarkGray
        $script:fail++
    } else {
        Write-Fail "no campaign processing found in logs — is campaign-worker running?"
    }
} catch {
    Write-Fail "could not read docker logs: $($_.Exception.Message)"
}

# ── 6. Verify ranking engine still healthy after load ─────────────────────────
Write-Host "`n-- Step 6: ranking-engine healthy after campaign run" -ForegroundColor Yellow
try {
    $h = Invoke-RestMethod "$RankerBase/health"
    if ($h.status -eq "ok") { Write-Pass "ranking-engine still healthy" }
    else                     { Write-Fail "ranking-engine health check failed after load" }
} catch {
    Write-Fail "ranking-engine went down during test"
}

# ── 7. Non-regression: existing endpoints still work ─────────────────────────
Write-Host "`n-- Step 7: non-regression — existing gateway routes intact" -ForegroundColor Yellow
try {
    $r = Invoke-RestMethod "$GatewayBase/auth/login" -Method POST `
        -ContentType "application/json" `
        -Body '{"email":"wrong@test.com","password":"wrongpass"}' `
        -ErrorAction SilentlyContinue
} catch {
    # We expect a 401; if we get a connection error that's a regression
    if ($_.Exception.Response.StatusCode -eq 401) {
        Write-Pass "auth rejects bad credentials (401) — gateway routes intact"
    } elseif ($_.Exception.Response) {
        Write-Pass "gateway responded (status=$($_.Exception.Response.StatusCode))"
    } else {
        Write-Fail "gateway did not respond: $($_.Exception.Message)"
    }
}

# ── Summary ───────────────────────────────────────────────────────────────────
Write-Host ""
Write-Host "─────────────────────────────────────" -ForegroundColor DarkGray
if ($fail -eq 0) {
    Write-Host "All $pass checks PASSED — Phase 6 is working end-to-end" -ForegroundColor Green
} else {
    Write-Host "$fail FAILED, $pass passed" -ForegroundColor Red
    Write-Host "`nDebug tips:" -ForegroundColor DarkYellow
    Write-Host "  docker logs ranking-engine"
    Write-Host "  docker logs campaign-worker"
    Write-Host "  docker compose ps"
    exit 1
}
