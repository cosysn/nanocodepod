@echo off
REM Build script for CodePod - Cross-platform build
REM Supports: Windows (amd64, arm64), Linux (amd64, arm64)
REM Requires Go (tested with 1.25.4)

setlocal

set OUTPUT_DIR=dist

REM Check Go version
go version >nul 2>&1
if errorlevel 1 (
    echo Error: Go is not installed
    exit /b 1
)

for /f "tokens=2" %%i in ('go version') do set "GO_VERSION=%%i"
echo Found Go %GO_VERSION%

REM Navigate to script directory
cd /d "%~dp0"
if errorlevel 1 (
    echo Error: Cannot find script directory
    exit /b 1
)

REM Create output directory
if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

echo.
echo Building for: windows amd64, windows arm64, linux amd64, linux arm64
echo Output directory: %OUTPUT_DIR%
echo.

REM Build windows amd64
set GOOS=windows
set GOARCH=amd64
echo Building codepod.exe (windows/amd64)...
go build -o "%OUTPUT_DIR%\codepod.exe" -ldflags "-s -w" .
echo Building codepod-agent.exe (windows/amd64)...
go build -o "%OUTPUT_DIR%\codepod-agent.exe" -ldflags "-s -w" ./cmd/agent

REM Build windows arm64
set GOOS=windows
set GOARCH=arm64
echo Building codepod-arm64.exe (windows/arm64)...
go build -o "%OUTPUT_DIR%\codepod-arm64.exe" -ldflags "-s -w" .
echo Building codepod-agent-arm64.exe (windows/arm64)...
go build -o "%OUTPUT_DIR%\codepod-agent-arm64.exe" -ldflags "-s -w" ./cmd/agent

REM Build linux amd64
set GOOS=linux
set GOARCH=amd64
echo Building codepod-amd64 (linux/amd64)...
go build -o "%OUTPUT_DIR%\codepod-amd64" -ldflags "-s -w" .
echo Building codepod-agent-amd64 (linux/amd64)...
go build -o "%OUTPUT_DIR%\codepod-agent-amd64" -ldflags "-s -w" ./cmd/agent

REM Build linux arm64
set GOOS=linux
set GOARCH=arm64
echo Building codepod-arm64 (linux/arm64)...
go build -o "%OUTPUT_DIR%\codepod-arm64" -ldflags "-s -w" .
echo Building codepod-agent-arm64 (linux/arm64)...
go build -o "%OUTPUT_DIR%\codepod-agent-arm64" -ldflags "-s -w" ./cmd/agent

echo.
echo ========================================
echo Build complete!
echo Output directory: %OUTPUT_DIR%
echo ========================================

dir /b "%OUTPUT_DIR%"

endlocal
