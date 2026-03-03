# Build script for CodePod - Cross-platform build
# Supports: Windows (amd64, arm64), Linux (amd64, arm64)
# Requires Go (tested with 1.25.4)

param(
    [string]$Platform = "all",  # windows, linux, all
    [string]$OutputDir = "dist"
)

$ErrorActionPreference = "Stop"

# Platform configurations
$platforms = @{
    "windows" = @(
        @{ GOOS = "windows"; GOARCH = "amd64"; Ext = ".exe" },
        @{ GOOS = "windows"; GOARCH = "arm64"; Ext = ".exe" }
    )
    "linux" = @(
        @{ GOOS = "linux"; GOARCH = "amd64"; Ext = "" },
        @{ GOOS = "linux"; GOARCH = "arm64"; Ext = "" }
    )
}

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
$codepodDir = $scriptDir

# Create output directory
if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir | Out-Null
}

# Determine which platforms to build
if ($Platform -eq "all") {
    $toBuild = @()
    foreach ($p in $platforms.Keys) {
        $toBuild += $platforms[$p]
    }
} elseif ($platforms.ContainsKey($Platform)) {
    $toBuild = $platforms[$Platform]
} else {
    Write-Host "Error: Invalid platform '$Platform'. Use: windows, linux, or all" -ForegroundColor Red
    exit 1
}

Write-Host "`nBuilding for: $($toBuild | ForEach-Object { $_.GOOS + "/" + $_.GOARCH })" -ForegroundColor Cyan
Write-Host "Output directory: $OutputDir`n" -ForegroundColor Cyan

Push-Location $codepodDir

try {
    $binaries = @("codepod", "codepod-agent")
    $mainFiles = @(".", "./cmd/agent")

    foreach ($config in $toBuild) {
        $os = $config.GOOS
        $arch = $config.GOARCH
        $ext = $config.Ext

        foreach ($i in 0..($binaries.Length - 1)) {
            $binaryName = $binaries[$i] + $ext
            $mainFile = $mainFiles[$i]

            Write-Host "Building $binaryName ($os/$arch)..." -ForegroundColor Yellow

            $env:GOOS = $os
            $env:GOARCH = $arch

            go build -ldflags "-s -w" -o "$OutputDir/$binaryName" $mainFile

            if ($LASTEXITCODE -ne 0) {
                throw "Failed to build $binaryName for $os/$arch"
            }

            Write-Host "  -> $OutputDir/$binaryName" -ForegroundColor Green
        }
    }

    Write-Host "`n========================================" -ForegroundColor Green
    Write-Host "Build complete!" -ForegroundColor Green
    Write-Host "Output directory: $OutputDir" -ForegroundColor White
    Write-Host "========================================" -ForegroundColor Green

    # List output files
    Get-ChildItem $OutputDir | ForEach-Object { Write-Host "  - $($_.Name)" -ForegroundColor White }

} catch {
    Write-Host "`nError: $_" -ForegroundColor Red
    exit 1
} finally {
    Pop-Location
    Remove-Item Env:GOOS -ErrorAction SilentlyContinue
    Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
}
