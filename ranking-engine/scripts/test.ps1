# Run C++ unit tests
# Usage: .\scripts\test.ps1

$ErrorActionPreference = "Stop"
$root = Split-Path $PSScriptRoot -Parent

Write-Host "`n=== Building ranking-engine ===" -ForegroundColor Cyan
cmake --build "$root\build" --config Release
if ($LASTEXITCODE -ne 0) { Write-Host "Build FAILED" -ForegroundColor Red; exit 1 }

Write-Host "`n=== Running C++ unit tests ===" -ForegroundColor Cyan
Push-Location "$root\build"
ctest --output-on-failure
$result = $LASTEXITCODE
Pop-Location

if ($result -eq 0) {
    Write-Host "`nAll C++ tests PASSED" -ForegroundColor Green
} else {
    Write-Host "`nC++ tests FAILED" -ForegroundColor Red
    exit 1
}
