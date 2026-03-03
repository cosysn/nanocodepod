@echo off
REM Build script for CodePod - Cross-platform build
REM Supports: Windows (amd64, arm64), Linux (amd64, arm64)
REM Requires Go (tested with 1.25.4)

setlocal EnableDelayedExpansion

set OUTPUT_DIR=dist

REM Parse arguments
set "PLATFORM=all"
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

REM Set platform list
set "PLATFORMS=windows/amd64,windows/arm64,linux/amd64,linux/arm64"
if /i "%PLATFORM%"=="windows" set "PLATFORMS=windows/amd64,windows/arm64"
if /i "%PLATFORM%"=="linux" set "PLATFORMS=linux/amd64,linux/arm64"

REM Build codepod CLI
for %%p in (%PLATFORMS%) do (
    for /f "tokens=1,2 delims=/" %%a in ("%%p") do (
        if "%%a"=="windows" set "EXT=.exe"
        if "%%a"=="linux" set "EXT="
        echo Building codepod%%EXT% (%%a/%%b)...
        go build -o "%OUTPUT_DIR%\codepod%%EXT%" -ldflags "-s -w" .
        if errorlevel 1 (
            echo Error: Failed to build codepod for %%a/%%b
            exit /b 1
        )
        echo   -^> %OUTPUT_DIR%\codepod%%EXT%
    )
)

REM Build codepod-agent
for %%p in (%PLATFORMS%) do (
    for /f "tokens=1,2 delims=/" %%a in ("%%p") do (
        if "%%a"=="windows" set "EXT=.exe"
        if "%%a"=="linux" set "EXT="
        echo Building codepod-agent%%EXT% (%%a/%%b)...
        go build -o "%OUTPUT_DIR%\codepod-agent%%EXT%" -ldflags "-s -w" ./cmd/agent
        if errorlevel 1 (
            echo Error: Failed to build codepod-agent for %%a/%%b
            exit /b 1
        )
        echo   -^> %OUTPUT_DIR%\codepod-agent%%EXT%
    )
)

echo.
echo ========================================
echo Build complete!
echo Output directory: %OUTPUT_DIR%
echo ========================================

dir /b "%OUTPUT_DIR%"

endlocal
