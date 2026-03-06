@echo off
:: fetch-ffmpeg.bat — download the FFmpeg Windows binary into third_party\bin\
::
:: Usage:
::   scripts\fetch-ffmpeg.bat [options]
::
:: Options:
::   --dir <path>   Output directory  (default: third_party\bin)
::   --help         Show this message
::
:: Source: BtbN FFmpeg Builds (https://github.com/BtbN/FFmpeg-Builds)
:: Downloads the GPL essentials build (statically linked - no extra DLLs needed).

setlocal EnableDelayedExpansion

cd /d "%~dp0\.."

set "OUTPUT_DIR=third_party\bin"

:parse
if "%~1"=="" goto :run
if /i "%~1"=="--dir"  ( set "OUTPUT_DIR=%~2" & shift & shift & goto :parse )
if /i "%~1"=="--help" ( goto :help )
echo Unknown option: %~1 1>&2
exit /b 1

:run
where curl >nul 2>&1
if errorlevel 1 (
  echo curl not found. Install Git for Windows or Windows 10+ curl. 1>&2
  exit /b 1
)

if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

set "DEST=%OUTPUT_DIR%\ffmpeg.exe"
set "ZIP_URL=https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl-essentials.zip"
set "TMP_ZIP=%TEMP%\ffmpeg_download_%RANDOM%.zip"

echo Downloading FFmpeg Windows binary (gpl-essentials)...
curl -fsSL "%ZIP_URL%" -o "%TMP_ZIP%"
if errorlevel 1 ( echo curl failed 1>&2 & exit /b 1 )

echo Extracting ffmpeg.exe...
powershell -NoProfile -Command ^
  "Add-Type -Assembly System.IO.Compression.FileSystem; ^
   $zip = [IO.Compression.ZipFile]::OpenRead('!TMP_ZIP!'); ^
   $entry = $zip.Entries ^| Where-Object { $_.Name -eq 'ffmpeg.exe' } ^| Select-Object -First 1; ^
   [IO.Compression.ZipFileExtensions]::ExtractToFile($entry, '!DEST!', $true); ^
   $zip.Dispose()"
if errorlevel 1 ( echo Extraction failed 1>&2 & del "!TMP_ZIP!" 2>nul & exit /b 1 )

del "!TMP_ZIP!" 2>nul
echo Done -> %DEST%
exit /b 0

:help
echo.
echo Usage: scripts\fetch-ffmpeg.bat [options]
echo.
echo Options:
echo   --dir ^<path^>   Output directory  (default: third_party\bin)
echo   --help          Show this message
echo.
exit /b 0
