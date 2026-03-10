@echo off
:: build.bat — compile stream-assistant on Windows
::
:: Usage:
::   scripts\build.bat [options]
::
:: Options:
::   --gui              Build the GUI binary (root package) instead of the CLI
::   --embedded         Embed yt-dlp binaries into the output  (requires third_party\bin\)
::   --embed-ffmpeg     Embed FFmpeg for Windows (requires third_party\bin\ffmpeg.exe)
::   --output <path>    Output binary path         (default: bin\stream-assistant[-gui].exe)
::   --os <goos>        Target GOOS                (default: windows)
::   --arch <goarch>    Target GOARCH              (default: current arch)
::   --version <ver>    Embed version string in binary via ldflags
::   --race             Enable race detector
::   --help             Show this message

setlocal EnableDelayedExpansion

cd /d "%~dp0\.."

set OUTPUT=
set GUI=0
set EMBED=0
set EMBED_FFMPEG=0
set GOOS=
set GOARCH=
set VERSION=
set RACE=0

:parse
if "%~1"=="" goto :build
if /i "%~1"=="--gui"          ( set "GUI=1" & shift & goto :parse )
if /i "%~1"=="--embedded"     ( set "EMBED=1" & shift & goto :parse )
if /i "%~1"=="--embed-ffmpeg" ( set "EMBED_FFMPEG=1" & shift & goto :parse )
if /i "%~1"=="--output"    ( set "OUTPUT=%~2" & shift & shift & goto :parse )
if /i "%~1"=="--os"        ( set "GOOS=%~2" & shift & shift & goto :parse )
if /i "%~1"=="--arch"      ( set "GOARCH=%~2" & shift & shift & goto :parse )
if /i "%~1"=="--version"   ( set "VERSION=%~2" & shift & shift & goto :parse )
if /i "%~1"=="--race"      ( set "RACE=1" & shift & goto :parse )
if /i "%~1"=="--help"      ( goto :usage )
echo Unknown option: %~1 >&2
exit /b 1

:usage
echo Usage:  scripts\build.bat [options]
echo.
echo Options:
echo   --gui              Build the GUI binary instead of the CLI
echo   --embedded         Embed yt-dlp binaries into the output
echo   --embed-ffmpeg     Embed FFmpeg for Windows (requires third_party\bin\ffmpeg.exe)
echo   --output ^<path^>   Output binary path      (default: bin\stream-assistant[-gui].exe)
echo   --os ^<goos^>       Target GOOS             (default: windows)
echo   --arch ^<goarch^>   Target GOARCH           (default: current arch)
echo   --version ^<ver^>   Embed version string in binary via ldflags
echo   --race            Enable race detector
echo   --help            Show this message
exit /b 0

:build
set "PKG=.\cmd\stream-assistant\"
set "LABEL=stream-assistant"
if "%GUI%"=="1" ( set "PKG=." & set "LABEL=stream-assistant-gui" )
if "%OUTPUT%"=="" set "OUTPUT=bin\%LABEL%.exe"

set "TAG_LIST="
if "%EMBED%"=="1" (
  if not exist "third_party\bin" (
    echo error: third_party\bin\ not found - run scripts\fetch-ytdlp.bat first >&2
    exit /b 1
  )
  set "TAG_LIST=embed_ytdlp"
)
if "%EMBED_FFMPEG%"=="1" (
  if not exist "third_party\bin\ffmpeg.exe" (
    echo error: third_party\bin\ffmpeg.exe not found - run scripts\fetch-ffmpeg.bat first >&2
    exit /b 1
  )
  if "!TAG_LIST!"=="" ( set "TAG_LIST=embed_ffmpeg" ) else ( set "TAG_LIST=!TAG_LIST!,embed_ffmpeg" )
)
set "TAGS="
if not "!TAG_LIST!"=="" set "TAGS=-tags !TAG_LIST!"

set "LDFLAGS_VAL="
if not "%VERSION%"=="" (
  set "LDFLAGS_VAL=-X main.version=%VERSION%"
)

set RACE_FLAG=
if "%RACE%"=="1" set "RACE_FLAG=-race"

echo Building %LABEL%...
echo   gui          : %GUI%
echo   embedded     : %EMBED%
echo   embed-ffmpeg : %EMBED_FFMPEG%
echo   output   : %OUTPUT%
if "%GOOS%"==""   ( for /f "delims=" %%g in ('go env GOOS')   do echo   GOOS     : %%g )
if not "%GOOS%"=="" echo   GOOS     : %GOOS%
if "%GOARCH%"=="" ( for /f "delims=" %%g in ('go env GOARCH') do echo   GOARCH   : %%g )
if not "%GOARCH%"=="" echo   GOARCH   : %GOARCH%
if not "%VERSION%"=="" echo   version  : %VERSION%

for %%d in ("%OUTPUT%") do if not exist "%%~dpd" mkdir "%%~dpd"

if "!LDFLAGS_VAL!"=="" (
  go build %RACE_FLAG% %TAGS% -o "%OUTPUT%" %PKG%
) else (
  go build %RACE_FLAG% %TAGS% -ldflags="!LDFLAGS_VAL!" -o "%OUTPUT%" %PKG%
)

if %ERRORLEVEL% neq 0 ( echo Build failed >&2 & exit /b %ERRORLEVEL% )
echo Done ^> %OUTPUT%
endlocal
