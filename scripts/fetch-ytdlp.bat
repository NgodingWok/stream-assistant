@echo off
:: fetch-ytdlp.bat — download yt-dlp release binaries into third_party\bin\
::
:: Usage:
::   scripts\fetch-ytdlp.bat [options]
::
:: Options:
::   --version <ver>   yt-dlp release tag  (default: latest via GitHub API)
::   --dir <path>      Output directory    (default: third_party\bin)
::   --platform <p>    linux,macos,windows,all  (default: all; comma-separated)
::   --help            Show this message

setlocal EnableDelayedExpansion

cd /d "%~dp0\.."

set "YTDLP_VERSION="
set "OUTPUT_DIR=third_party\bin"
set "PLATFORM=all"

:parse
if "%~1"=="" goto :run
if /i "%~1"=="--version"  ( set "YTDLP_VERSION=%~2" & shift & shift & goto :parse )
if /i "%~1"=="--dir"      ( set "OUTPUT_DIR=%~2"    & shift & shift & goto :parse )
if /i "%~1"=="--platform" ( set "PLATFORM=%~2"      & shift & shift & goto :parse )
if /i "%~1"=="--help"     ( goto :help )
echo Unknown option: %~1 1>&2
exit /b 1

:run
:: ── curl availability check ─────────────────────────────────────────────────
where curl >nul 2>&1
if errorlevel 1 (
  echo curl not found. Install Git for Windows or Windows 10+ curl. 1>&2
  exit /b 1
)

:: ── resolve version ──────────────────────────────────────────────────────────
if "%YTDLP_VERSION%"=="" (
  echo Fetching latest yt-dlp version...
  for /f "delims=" %%v in ('curl -fsSL "https://api.github.com/repos/yt-dlp/yt-dlp/releases/latest" ^| findstr "tag_name" ^| for /f "tokens=2 delims=:," %a in ^('^) do @echo %~a') do (
    set "RAW=%%v"
  )
  :: Simpler: use PowerShell for JSON parsing if available
  for /f "delims=" %%v in ('powershell -NoProfile -Command "(Invoke-RestMethod https://api.github.com/repos/yt-dlp/yt-dlp/releases/latest).tag_name"') do (
    set "YTDLP_VERSION=%%v"
  )
  echo Latest version: !YTDLP_VERSION!
)

set "BASE_URL=https://github.com/yt-dlp/yt-dlp/releases/download/!YTDLP_VERSION!"

if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

echo Downloading yt-dlp !YTDLP_VERSION! to %OUTPUT_DIR%\

:: ── download helper ──────────────────────────────────────────────────────────
if "%PLATFORM%"=="all" (
  call :download "!BASE_URL!/yt-dlp_linux"           "%OUTPUT_DIR%\yt-dlp_linux"
  call :download "!BASE_URL!/yt-dlp_linux_aarch64"   "%OUTPUT_DIR%\yt-dlp_linux_aarch64"
  call :download "!BASE_URL!/yt-dlp_linux_musl"      "%OUTPUT_DIR%\yt-dlp_musllinux"
  call :download "!BASE_URL!/yt-dlp_linux_musl_aarch64" "%OUTPUT_DIR%\yt-dlp_musllinux_aarch64"
  call :download "!BASE_URL!/yt-dlp_macos"           "%OUTPUT_DIR%\yt-dlp_macos"
  call :download "!BASE_URL!/yt-dlp.exe"             "%OUTPUT_DIR%\yt-dlp.exe"
  call :download "!BASE_URL!/yt-dlp_x86.exe"         "%OUTPUT_DIR%\yt-dlp_x86.exe"
  goto :done
)

:: ── selective by platform ────────────────────────────────────────────────────
for %%p in (%PLATFORM%) do (
  if /i "%%p"=="linux"          call :download "!BASE_URL!/yt-dlp_linux"              "%OUTPUT_DIR%\yt-dlp_linux"
  if /i "%%p"=="linux-arm64"    call :download "!BASE_URL!/yt-dlp_linux_aarch64"      "%OUTPUT_DIR%\yt-dlp_linux_aarch64"
  if /i "%%p"=="linux-musl"     call :download "!BASE_URL!/yt-dlp_linux_musl"         "%OUTPUT_DIR%\yt-dlp_musllinux"
  if /i "%%p"=="linux-musl-arm64" call :download "!BASE_URL!/yt-dlp_linux_musl_aarch64" "%OUTPUT_DIR%\yt-dlp_musllinux_aarch64"
  if /i "%%p"=="macos"          call :download "!BASE_URL!/yt-dlp_macos"              "%OUTPUT_DIR%\yt-dlp_macos"
  if /i "%%p"=="windows"        call :download "!BASE_URL!/yt-dlp.exe"                "%OUTPUT_DIR%\yt-dlp.exe"
  if /i "%%p"=="windows-x86"    call :download "!BASE_URL!/yt-dlp_x86.exe"            "%OUTPUT_DIR%\yt-dlp_x86.exe"
)

:done
echo Done. Binaries written to %OUTPUT_DIR%\
exit /b 0

:download
echo   Downloading %~nx2...
curl -fsSL "%~1" -o "%~2"
if errorlevel 1 ( echo Failed to download %~nx2 1>&2 & exit /b 1 )
exit /b 0

:help
echo.
echo Usage: scripts\fetch-ytdlp.bat [options]
echo.
echo Options:
echo   --version ^<ver^>    yt-dlp release tag (default: latest)
echo   --dir ^<path^>       Output directory   (default: third_party\bin)
echo   --platform ^<p^>     Comma-separated: linux,linux-arm64,linux-musl,
echo                      linux-musl-arm64,macos,windows,windows-x86,all
echo   --help             Show this message
echo.
exit /b 0
