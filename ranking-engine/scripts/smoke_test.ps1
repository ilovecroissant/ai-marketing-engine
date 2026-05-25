# HTTP smoke tests against the running ranking-engine server.
# Start the server first: .\build\ranking-server.exe
# Usage: .\scripts\smoke_test.ps1

param([string]$Base = "http://localhost:8082")

$pass = 0
$fail = 0

function Assert-Response {
    param(
        [string]$Name,
        [string]$Method = "POST",
        [string]$Path,
        [hashtable]$Body = @{},
        [scriptblock]$Check
    )
    try {
        $params = @{ Uri = "$Base$Path"; Method = $Method; ContentType = "application/json" }
        if ($Method -eq "POST") { $params.Body = ($Body | ConvertTo-Json -Depth 5) }
        $resp = Invoke-RestMethod @params

        $ok = & $Check $resp
        if ($ok) {
            Write-Host "  PASS  $Name" -ForegroundColor Green
            $script:pass++
        } else {
            Write-Host "  FAIL  $Name  (check failed, got: $($resp | ConvertTo-Json -Compress))" -ForegroundColor Red
            $script:fail++
        }
    } catch {
        Write-Host "  FAIL  $Name  ($($_.Exception.Message))" -ForegroundColor Red
        $script:fail++
    }
}

Write-Host "`n=== Smoke tests → $Base ===`n" -ForegroundColor Cyan

# ── /health ──────────────────────────────────────────────────────────────────
Assert-Response "GET /health returns {status:ok}" -Method GET -Path "/health" -Check {
    param($r) $r.status -eq "ok"
}

# ── /keywords ─────────────────────────────────────────────────────────────────
Assert-Response "/keywords returns array" -Path "/keywords" -Body @{
    text = "digital marketing content strategy SEO growth analytics platform"
    n    = 5
} -Check { param($r) $r.keywords -is [array] -and $r.keywords.Count -gt 0 }

Assert-Response "/keywords terms are strings with scores" -Path "/keywords" -Body @{
    text = "machine learning artificial intelligence neural networks deep learning"
    n    = 3
} -Check { param($r)
    $r.keywords.Count -le 3 -and
    $r.keywords[0].term -is [string] -and
    $r.keywords[0].score -gt 0
}

Assert-Response "/keywords with empty text returns empty array" -Path "/keywords" -Body @{
    text = ""
    n    = 10
} -Check { param($r) $r.keywords -is [array] }

# ── /rank ─────────────────────────────────────────────────────────────────────
Assert-Response "/rank returns all four score fields" -Path "/rank" -Body @{
    text     = "Digital marketing drives SEO growth and content engagement. " * 10
    keywords = @("marketing", "seo", "content", "digital")
} -Check { param($r)
    $r.seo         -ge 0 -and $r.seo         -le 1 -and
    $r.engagement  -ge 0 -and $r.engagement  -le 1 -and
    $r.readability -ge 0 -and $r.readability -le 1 -and
    $r.total       -ge 0 -and $r.total       -le 1
}

Assert-Response "/rank total matches weighted formula" -Path "/rank" -Body @{
    text     = "content marketing strategy analytics growth digital platform"
    keywords = @("content", "marketing")
} -Check { param($r)
    $expected = 0.4 * $r.seo + 0.3 * $r.engagement + 0.3 * $r.readability
    [Math]::Abs($r.total - $expected) -lt 0.001
}

Assert-Response "/rank with no keywords returns neutral seo=0.5" -Path "/rank" -Body @{
    text     = "some article about things"
    keywords = @()
} -Check { param($r) [Math]::Abs($r.seo - 0.5) -lt 0.01 }

Assert-Response "/rank with empty text does not crash" -Path "/rank" -Body @{
    text     = ""
    keywords = @("marketing")
} -Check { param($r) $r.total -ge 0 }

# ── /similarity ───────────────────────────────────────────────────────────────
Assert-Response "/similarity identical texts → similarity=1.0" -Path "/similarity" -Body @{
    a = "machine learning algorithms neural networks"
    b = "machine learning algorithms neural networks"
} -Check { param($r) $r.similarity -ge 0.99 }

Assert-Response "/similarity disjoint texts → similarity≈0" -Path "/similarity" -Body @{
    a = "apple orange banana fruit salad recipe"
    b = "quantum physics electrons protons neutrons"
} -Check { param($r) $r.similarity -lt 0.1 }

Assert-Response "/similarity near-duplicate flagged" -Path "/similarity" -Body @{
    a = "digital marketing content creation strategy platform"
    b = "content creation digital marketing strategy platform"
} -Check { param($r) $r.similarity -gt 0.8 -and $r.is_duplicate -eq $true }

Assert-Response "/similarity different content not flagged" -Path "/similarity" -Body @{
    a = "cooking recipes for healthy dinner"
    b = "machine learning artificial intelligence"
} -Check { param($r) $r.is_duplicate -eq $false }

Assert-Response "/similarity empty input returns 0" -Path "/similarity" -Body @{
    a = ""
    b = "some text here"
} -Check { param($r) $r.similarity -eq 0.0 }

# ── summary ───────────────────────────────────────────────────────────────────
Write-Host ""
if ($fail -eq 0) {
    Write-Host "All $pass smoke tests PASSED" -ForegroundColor Green
} else {
    Write-Host "$fail FAILED, $pass passed" -ForegroundColor Red
    exit 1
}
