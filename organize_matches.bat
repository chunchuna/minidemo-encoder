@echo off
chcp 65001 >nul
echo ========================================
echo CS:GO Match Organizer
echo ========================================
echo.

REM Check if output directory exists
if not exist "output" (
    echo Error: output directory not found
    echo Please make sure you have parsed demos first
    pause
    exit /b 1
)

REM Run the organizer
match_organizer.exe -output=./output

echo.
echo Press any key to exit...
pause >nul 