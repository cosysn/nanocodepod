@echo off
REM Build script for CodePod on Windows
REM Requires Go (tested with 1.25.4)

setlocal

REM Check Go version
go version >nul 2>&1
if errorlevel 1 (
    echo Error: Go is not installed
    exit /b 1
)

REM Get installed Go version
for /f "tokens=2" %%i in ('go version') do set INSTALLED_VERSION=%%i
echo Found Go %INSTALLED_VERSION%

REM Navigate to script directory (where build.bat is located)
cd /d "%~dp0"
if errorlevel 1 (
    echo Error: Cannot find script directory
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
