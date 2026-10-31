# localsolver quick start (Windows PowerShell)
# First run: creates venv, installs deps, fetches camoufox browser, copies .env
$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

if (-not (Test-Path ".venv")) {
    Write-Host "Creating venv + installing deps (first run, a few minutes)..." -ForegroundColor Cyan
    python -m venv .venv
    .\.venv\Scripts\python.exe -m pip install -U pip -q
    .\.venv\Scripts\python.exe -m pip install -r requirements.txt
    Write-Host "Fetching camoufox browser..." -ForegroundColor Cyan
    .\.venv\Scripts\python.exe -m camoufox fetch
}
if (-not (Test-Path ".env")) { Copy-Item ".env.example" ".env"; Write-Host "Created .env from example" -ForegroundColor Yellow }

Write-Host "Starting localsolver..." -ForegroundColor Green
.\.venv\Scripts\python.exe main.py
