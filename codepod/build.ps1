# Build script for CodePod - Cross-platform build
# Supports: Windows (amd64, arm64), Linux (amd64, arm64)
# Requires Go (tested with 1.25.4)

param(
    [string]$Platform = "all",
    [string]$OutputDir = "dist"
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

# Navigate to script directory
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path

# Create output directory
if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir | Out-Null
}

# Define build configurations
$configs = @(
    @{ GOOS = "windows"; GOARCH = "amd64"; Ext = ".exe" },
    @{ GOOS = "windows"; GOARCH = "arm64"; Ext = ".exe" },
    @{ GOOS = "linux"; GOARCH = "amd64"; Ext = "" },
    @{ GOOS = "linux"; GOARCH = "arm64"; Ext = "" }
)

# Filter by platform if specified
if ($Platform -ne "all") {
    $configs = $configs | Where-Object { $_.GOOS -eq $Platform }
}

Write-Host "`nBuilding for: $($configs | ForEach-Object { $_.GOOS + "/" + $_.GOARCH } -join ", ")" -ForegroundColor Cyan
Write-Host "Output directory: $OutputDir`n" -ForegroundColor Cyan

Push-Location $scriptDir

function Build-Binary {
    param(
        [string]$GOOS,
        [string]$GOARCH,
        [string]$Ext,
        [string]$Main,
        [string]$Name
    )

    $env:GOOS = $GOOS
    $env:GOARCH = $GOARCH

    $outputName = $Name + $Ext
    if ($GOOS -eq "linux") {
        $outputName = $Name + "-" + $GOARCH
    }

    Write-Host "Building $outputName ($GOOS/$GOARCH)..." -ForegroundColor Yellow

    go build -ldflags "-s -w" -o "$OutputDir/$outputName" $Main

    if ($LASTEXITCODE -ne 0) {
        throw "Failed to build $outputName"
    }

    Write-Host "  -> $OutputDir/$outputName" -ForegroundColor Green
}

try {
    foreach ($config in $configs) {
        Build-Binary -GOOS $config.GOOS -GOARCH $config.GOARCH -Ext $config.Ext -Main "." -Name "codepod"
        Build-Binary -GOOS $config.GOOS -GOARCH $config.GOARCH -Ext $config.Ext -Main "./cmd/agent" -Name "codepod-agent"
    }

    Write-Host "`n========================================" -ForegroundColor Green
    Write-Host "Build complete!" -ForegroundColor Green
    Write-Host "Output directory: $OutputDir" -ForegroundColor White
    Write-Host "========================================" -ForegroundColor Green

    Get-ChildItem $OutputDir | ForEach-Object { Write-Host "  - $($_.Name)" -ForegroundColor White }

} catch {
    Write-Host "`nError: $_" -ForegroundColor Red
    exit 1
} finally {
    Pop-Location
    $env:GOOS = $null
    $env:GOARCH = $null
}
