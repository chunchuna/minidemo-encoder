@echo off
chcp 65001 >nul
REM REC 文件平衡工具
REM 确保每个地图的每个t/ct文件夹都有5个不同的rec文件

setlocal enabledelayedexpansion

if "%~1"=="" (
    set BASE_PATH=D:\SteamLibrary\steamapps\common\Counter-Strike Global Offensive\csgo\addons\sourcemod\data\botmimic\demotest
) else (
    set BASE_PATH=%~1
)

echo ========================================
echo REC 文件平衡工具
echo ========================================
echo 目标路径: %BASE_PATH%
echo.

if not exist "%BASE_PATH%" (
    echo 错误: 路径不存在: %BASE_PATH%
    echo.
    echo 用法: rec_balance.bat [基础路径]
    echo 示例: rec_balance.bat "D:\SteamLibrary\steamapps\common\Counter-Strike Global Offensive\csgo\addons\sourcemod\data\botmimic\demotest"
    echo.
    pause
    exit /b 1
)

REM 检查是否需要编译
cd /d "%~dp0"
if not exist "rec_balancer.exe" (
    echo 正在编译程序...
    go build -o rec_balancer.exe .\cmd\rec_balancer\main.go
    if errorlevel 1 (
        echo 编译失败！
        pause
        exit /b 1
    )
    echo 编译完成！
    echo.
)

REM 执行平衡操作
rec_balancer.exe -path="%BASE_PATH%"

echo.
pause 