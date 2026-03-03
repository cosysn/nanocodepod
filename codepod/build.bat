@echo off
REM Build script for CodePod on Windows
REM Requires Go 1.25.4

setlocal

set GO_VERSION=1.25.4

REM Check Go version
go version >nul 2>&1
if errorlevel 1 (
    echo Error: Go is not installed
    exit /b 1
)

REM Get installed Go version
for /f "tokens=2" %%i in ('go version') do set INSTALLED_VERSION=%%i
echo Found Go %INSTALLED_VERSION%

REM Navigate to codepod directory
cd /d "%~dp0codepod"
if errorlevel 1 (
    echo Error: Cannot find codepod directory
    exit /b 1
)

echo.
echo Building codepod (CLI)...
go build -o codepod.exe .
if errorlevel 1 (
    echo Error: Failed to build codepod
    exit /b 1
)
echo Build complete: codepod.exe

echo.
echo Building codepod-agent...
go build -o codepod-agent.exe ./cmd/agent
if errorlevel 1 (
    echo Error: Failed to build codepod-agent
    exit /b 1
)
echo Build complete: codepod-agent.exe

echo.
echo Build successful!
echo   - codepod.exe       (CLI)
echo   - codepod-agent.exe (Agent)

endlocal
