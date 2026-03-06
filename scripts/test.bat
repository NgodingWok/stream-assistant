@echo off
:: test.bat — run tests for stream-assistant
::
:: Usage:
::   scripts\test.bat [options]
::
:: Options:
::   --unit            Run unit tests only           (default: all non-integration)
::   --integration     Run integration tests only    (requires network access)
::   --all             Run both unit and integration tests
::   --coverage        Generate coverage report      (outputs coverage\coverage.html)
::   --race            Enable race detector
::   --run <pattern>   Filter tests by name pattern
::   --timeout <dur>   Test timeout                  (default: 2m; integration: 5m)
::   --verbose         Enable verbose output (-v)
::   --help            Show this message

setlocal EnableDelayedExpansion

cd /d "%~dp0\.."

set "RUN_UNIT=1"
set "RUN_INTEGRATION=0"
set "COVERAGE=0"
set "RACE_FLAG="
set "VERBOSE_FLAG="
set "FILTER_FLAG="
set "UNIT_TIMEOUT=2m"
set "INT_TIMEOUT=5m"

:parse
if "%~1"=="" goto :run
if /i "%~1"=="--unit"        ( set "RUN_UNIT=1" & set "RUN_INTEGRATION=0" & shift & goto :parse )
if /i "%~1"=="--integration" ( set "RUN_UNIT=0" & set "RUN_INTEGRATION=1" & shift & goto :parse )
if /i "%~1"=="--all"         ( set "RUN_UNIT=1" & set "RUN_INTEGRATION=1" & shift & goto :parse )
if /i "%~1"=="--coverage"    ( set "COVERAGE=1" & shift & goto :parse )
if /i "%~1"=="--race"        ( set "RACE_FLAG=-race" & shift & goto :parse )
if /i "%~1"=="--verbose"     ( set "VERBOSE_FLAG=-v" & shift & goto :parse )
if /i "%~1"=="--run"         ( set "FILTER_FLAG=-run %~2" & shift & shift & goto :parse )
if /i "%~1"=="--timeout"     ( set "UNIT_TIMEOUT=%~2" & set "INT_TIMEOUT=%~2" & shift & shift & goto :parse )
if /i "%~1"=="--help"        ( goto :help )
echo Unknown option: %~1 1>&2
exit /b 1

:run
set "COVER_UNIT="
set "COVER_INT="

if "%RUN_UNIT%"=="1" (
  echo Running unit tests (timeout: !UNIT_TIMEOUT!)...
  if "%COVERAGE%"=="1" (
    if not exist coverage mkdir coverage
    set "COVER_UNIT=-coverprofile=coverage\unit.out -covermode=atomic"
  )
  for /f "delims=" %%p in ('go list ./...') do (
    echo %%p | findstr /v "test/integration" >nul 2>&1
    if not errorlevel 1 (
      set "UNIT_PKGS=!UNIT_PKGS! %%p"
    )
  )
  go test !RACE_FLAG! !VERBOSE_FLAG! -timeout !UNIT_TIMEOUT! !COVER_UNIT! !FILTER_FLAG! !UNIT_PKGS!
  if errorlevel 1 ( echo Unit tests FAILED. & exit /b 1 )
  echo Unit tests passed.
)

if "%RUN_INTEGRATION%"=="1" (
  echo Running integration tests (timeout: !INT_TIMEOUT!)...
  echo Warning: integration tests require network access.
  if "%COVERAGE%"=="1" (
    if not exist coverage mkdir coverage
    set "COVER_INT=-coverprofile=coverage\integration.out -covermode=atomic"
  )
  go test -tags integration !RACE_FLAG! !VERBOSE_FLAG! -timeout !INT_TIMEOUT! !COVER_INT! !FILTER_FLAG! .\test\integration\
  if errorlevel 1 ( echo Integration tests FAILED. & exit /b 1 )
  echo Integration tests passed.
)

if "%COVERAGE%"=="1" (
  if exist coverage\unit.out (
    go tool cover -html=coverage\unit.out -o coverage\coverage.html
    echo Coverage report: coverage\coverage.html
    go tool cover -func=coverage\unit.out | findstr "total:"
  )
)

exit /b 0

:help
echo.
echo Usage: scripts\test.bat [options]
echo.
echo Options:
echo   --unit            Run unit tests only           (default)
echo   --integration     Run integration tests only
echo   --all             Run unit and integration tests
echo   --coverage        Generate coverage report
echo   --race            Enable race detector
echo   --run ^<pattern^>   Filter tests by name pattern
echo   --timeout ^<dur^>   Test timeout (default: 2m / 5m for integration)
echo   --verbose         Enable verbose output
echo   --help            Show this message
echo.
exit /b 0
