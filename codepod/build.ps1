# Build script for CodePod on Windows (PowerShell)
# Requires Go (tested with 1.25.4)

param(
    [string]$OutputDir = "."
)

$ErrorActionPreference = "Stop"

# Check Go version
try {
    $goVersion = go version
    Write-Host "Found: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "Error: Go is not installed" -ForegroundColor Red
    exit 1
}

# Navigate to script directory (where build.ps1 is located)
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$codepodDir = $scriptDir

Push-Location $codepodDir

try {
    Write-Host "`nBuilding codepod (CLI)..." -ForegroundColor Cyan
    go build -o codepod.exe .
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to build codepod"
    }
    Write-Host "Build complete: codepod.exe" -ForegroundColor Green

    Write-Host "`nBuilding codepod-agent..." -ForegroundColor Cyan
    go build -o codepod-agent.exe ./cmd/agent
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to build codepod-agent"
    }
    Write-Host "Build complete: codepod-agent.exe" -ForegroundColor Green

    # Copy to output directory if specified
    if ($OutputDir -ne "." -and (Test-Path $OutputDir)) {
        Copy-Item codepod.exe, codepod-agent.exe -Destination $OutputDir -Force
        Write-Host "`nCopied executables to: $OutputDir" -ForegroundColor Green
    }

    Write-Host "`n========================================" -ForegroundColor Green
    Write-Host "Build successful!" -ForegroundColor Green
    Write-Host "  - codepod.exe       (CLI)" -ForegroundColor White
    Write-Host "  - codepod-agent.exe (Agent)" -ForegroundColor White
    Write-Host "========================================" -ForegroundColor Green

} catch {
    Write-Host "`nError: $_" -ForegroundColor Red
    exit 1
} finally {
    Pop-Location
}
