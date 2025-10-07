@echo off
chcp 65001 >nul
REM Batch parse CS:GO Demo script (skip freezetime)
REM Usage: batch_parse_skipfreeze.bat [demo folder path]

setlocal enabledelayedexpansion

if "%~1"=="" (
    set DEMO_DIR=.\demo
) else (
    set DEMO_DIR=%~1
)

echo ========================================
echo CS:GO Demo Batch Parser (Skip Freezetime)
echo ========================================
echo Demo Directory: %DEMO_DIR%
echo Mode: Skip Freezetime (Preparation Phase)
echo.

if not exist "%DEMO_DIR%" (
    echo Error: Directory does not exist: %DEMO_DIR%
    echo Usage: batch_parse_skipfreeze.bat [demo folder path]
    pause
    exit /b 1
)

REM Check if compilation is needed
if not exist "minidemo-encoder.exe" (
    echo Compiling program...
    go build -o minidemo-encoder.exe .\cmd\main.go
    if errorlevel 1 (
        echo Compilation failed!
        pause
        exit /b 1
    )
    echo Compilation completed!
    echo.
)

REM Execute batch parsing (skip freezetime)
minidemo-encoder.exe -dir="%DEMO_DIR%" -skipfreeze

echo.
echo Parsing completed! Output files are in .\output directory
echo Note: 
echo   - Generated rec files do not include freezetime content
echo   - Output folder name format for each demo: {tickrate}{demo-filename}
echo   - Example: 128.015625your-demo-name
echo   - Folder contains tickrate information text file for that demo
echo.
pause 