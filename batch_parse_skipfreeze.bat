@echo off
chcp 65001 >nul
REM 批量解析 CS:GO Demo 脚本 (跳过准备时间)
REM 用法: batch_parse_skipfreeze.bat [demo文件夹路径]

setlocal enabledelayedexpansion

if "%~1"=="" (
    set DEMO_DIR=.\demo
) else (
    set DEMO_DIR=%~1
)

echo ========================================
echo CS:GO Demo 批量解析工具 (跳过准备时间)
echo ========================================
echo Demo 目录: %DEMO_DIR%
echo 模式: 跳过 Freezetime (准备阶段)
echo.

if not exist "%DEMO_DIR%" (
    echo 错误: 目录不存在: %DEMO_DIR%
    echo 用法: batch_parse_skipfreeze.bat [demo文件夹路径]
    pause
    exit /b 1
)

REM 检查是否需要编译
if not exist "minidemo-encoder.exe" (
    echo 正在编译程序...
    go build -o minidemo-encoder.exe .\cmd\main.go
    if errorlevel 1 (
        echo 编译失败！
        pause
        exit /b 1
    )
    echo 编译完成！
    echo.
)

REM 执行批量解析 (跳过准备时间)
minidemo-encoder.exe -dir="%DEMO_DIR%" -skipfreeze

echo.
echo 解析完成！输出文件在 .\output 目录中
echo 注意: 生成的 rec 文件不包含准备时间内容
echo.
pause 