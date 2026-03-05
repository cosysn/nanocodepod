# Build script for CodePod with embedded agents (Windows PowerShell)
# Builds agent binaries for all platforms and embeds them into the codepod binary

param(
    [string]$Platform = "all",
    [string]$OutputDir = "dist"
)

$ErrorActionPreference = "Stop"

# Colors
function Write-Green { param($msg) Write-Host $msg -ForegroundColor Green }
function Write-Yellow { param($msg) Write-Host $msg -ForegroundColor Yellow }
function Write-Red { param($msg) Write-Host $msg -ForegroundColor Red }

Write-Green "=== CodePod Embedded Agent Build ==="

# Check Go version
try {
    $goVersion = go version
    Write-Green "Found: $goVersion"
} catch {
    Write-Red "Error: Go is not installed"
    exit 1
}

# Navigate to script directory
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$embedDir = Join-Path $scriptDir "pkg\embed"

# Create temp directory
$tempDir = Join-Path $env:TEMP "codepod-embed-build"
if (Test-Path $tempDir) {
    Remove-Item -Recurse -Force $tempDir
}
New-Item -ItemType Directory -Path $tempDir | Out-Null

# Create output directory
if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir | Out-Null
}

# Define platforms to build
$platforms = @(
    @{ GOOS = "linux"; GOARCH = "amd64"; OutputName = "agent-linux-amd64" },
    @{ GOOS = "linux"; GOARCH = "arm64"; OutputName = "agent-linux-arm64" },
    @{ GOOS = "darwin"; GOARCH = "amd64"; OutputName = "agent-darwin-amd64" },
    @{ GOOS = "darwin"; GOARCH = "arm64"; OutputName = "agent-darwin-arm64" },
    @{ GOOS = "windows"; GOARCH = "amd64"; OutputName = "agent-windows-amd64.exe" }
)

Write-Yellow "`nBuilding agent binaries..."

Push-Location $scriptDir

try {
    foreach ($p in $platforms) {
        $env:GOOS = $p.GOOS
        $env:GOARCH = $p.GOARCH

        Write-Host "  Building $($p.OutputName) ($($p.GOOS)/$($p.GOARCH))..." -NoNewline

        go build -ldflags "-s -w" -o "$tempDir\$($p.OutputName)" .\cmd\agent

        if ($LASTEXITCODE -ne 0) {
            throw "Failed to build $($p.OutputName)"
        }

        Write-Green " OK"
    }

    # Reset Go environment
    $env:GOOS = $null
    $env:GOARCH = $null

    # Create agents.zip
    Write-Yellow "`nCreating agents.zip..."

    Set-Location $tempDir
    $zipPath = Join-Path $embedDir "agents.zip"
    if (Test-Path $zipPath) {
        Remove-Item $zipPath -Force
    }

    # Use Compress-Archive (PowerShell 5.1+)
    Get-ChildItem -Filter "agent-*" | ForEach-Object {
        # PowerShell's Compress-Archive doesn't support stored (uncompressed) files
        # So we use .NET directly for zip creation
    }

    # Use .NET to create uncompressed zip
    Add-Type -AssemblyName System.IO.Compression.FileSystem

    $zipStream = [System.IO.File]::Create($zipPath)
    $zipArchive = New-Object System.IO.Compression.ZipArchive($zipStream, [System.IO.Compression.CompressionMode]::Create)

    foreach ($file in Get-ChildItem -Filter "agent-*") {
        $entry = $zipArchive.CreateEntry($file.Name, [System.IO.Compression.CompressionLevel]::NoCompression)
        $entryStream = $entry.Open()
        $fileStream = [System.IO.File]::OpenRead($file.FullName)
        $fileStream.CopyTo($entryStream)
        $fileStream.Close()
        $entryStream.Close()
    }

    $zipArchive.Dispose()
    $zipStream.Dispose()

    Write-Green "  Created: $zipPath"

    # Show zip contents
    Write-Host "  Contents:"
    $zip = [System.IO.Compression.ZipFile]::OpenRead($zipPath)
    foreach ($entry in $zip.Entries) {
        Write-Host "    $($entry.Name) ($($entry.Length) bytes)"
    }
    $zip.Dispose()

    # Clean up temp files
    Set-Location $scriptDir
    Remove-Item -Recurse -Force $tempDir

    Write-Green "`nEmbedded agents created successfully!"

    # Build final codepod binary with embedded agents
    Write-Yellow "`nBuilding codepod binary with embedded agents..."

    go build -ldflags "-s -w" -o "$OutputDir\codepod" .

    Write-Green "OK"

    Write-Green "`n=== Build Complete ==="
    Write-Host "Output: $OutputDir\codepod" -ForegroundColor White

} catch {
    Write-Red "`nError: $_"
    exit 1
} finally {
    Pop-Location
    $env:GOOS = $null
    $env:GOARCH = $null
}
