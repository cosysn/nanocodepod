@echo off
REM Build script for CodePod - Cross-platform build
REM Supports: Windows (amd64, arm64), Linux (amd64, arm64)
REM Requires Go (tested with 1.25.4)

setlocal EnableDelayedExpansion

set OUTPUT_DIR=dist
set PLATFORM=all

REM Parse arguments
if not "%1"=="" set "PLATFORM=%1"
if not "%2"=="" set "OUTPUT_DIR=%2"

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
echo Building for: windows/amd64, windows/arm64, linux/amd64, linux/arm64
echo Output directory: %OUTPUT_DIR%
echo.

REM Build for each platform
call :build_one windows amd64 .exe
call :build_one windows arm64 .exe
call :build_one linux amd64
call :build_one linux arm64

echo.
echo ========================================
echo Build complete!
echo Output directory: %OUTPUT_DIR%
echo ========================================

dir /b "%OUTPUT_DIR%"

endlocal
exit /b 0

:build_one
set "GOOS=%1"
set "GOARCH=%2"
set "EXT=%3"

if "%GOOS%"=="windows" (
    echo Building codepod%EXT% (%GOOS%/%GOARCH%)...
    go build -o "%OUTPUT_DIR%\codepod%EXT%" -ldflags "-s -w" .
    if errorlevel 1 goto :build_error

    echo Building codepod-agent%EXT% (%GOOS%/%GOARCH%)...
    go build -o "%OUTPUT_DIR%\codepod-agent%EXT%" -ldflags "-s -w" ./cmd/agent
    if errorlevel 1 goto :build_error
) else (
    echo Building codepod-%GOARCH% (%GOOS%/%GOARCH%)...
    go build -o "%OUTPUT_DIR%\codepod-%GOARCH%" -ldflags "-s -w" .
    if errorlevel 1 goto :build_error

    echo Building codepod-agent-%GOARCH% (%GOOS%/%GOARCH%)...
    go build -o "%OUTPUT_DIR%\codepod-agent-%GOARCH%" -ldflags "-s -w" ./cmd/agent
    if errorlevel 1 goto :build_error
)
exit /b 0

:build_error
echo Error: Failed to build for %GOOS%/%GOARCH%
exit /b 1
