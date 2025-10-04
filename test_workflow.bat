@echo off
chcp 65001 >nul
REM 完整工作流程测试脚本
REM 此脚本用于演示从批量解析到平衡的完整流程

setlocal enabledelayedexpansion

echo ========================================
echo  MiniDemo Encoder 完整工作流程测试
echo ========================================
echo.
echo 此脚本将演示以下流程：
echo   1. 检查环境
echo   2. 编译程序
echo   3. 批量解析 demo（如果有）
echo   4. 平衡 rec 文件（如果指定）
echo.
echo ========================================
echo.

REM 检查 Go 环境
echo [1/4] 检查 Go 环境...
go version >nul 2>&1
if errorlevel 1 (
    echo ❌ 错误: 未找到 Go 环境
    echo 请先安装 Go: https://golang.org/dl/
    pause
    exit /b 1
)
echo ✓ Go 环境已安装
echo.

REM 切换到脚本目录
cd /d "%~dp0"

REM 编译主程序
echo [2/4] 编译 Demo 解析器...
if not exist "minidemo-encoder.exe" (
    go build -o minidemo-encoder.exe .\cmd\main.go
    if errorlevel 1 (
        echo ❌ 编译失败！
        pause
        exit /b 1
    )
    echo ✓ 解析器编译完成
) else (
    echo ✓ 解析器已存在
)
echo.

REM 编译平衡工具
echo 编译 REC 平衡工具...
if not exist "rec_balancer.exe" (
    go build -o rec_balancer.exe .\cmd\rec_balancer\main.go
    if errorlevel 1 (
        echo ❌ 编译失败！
        pause
        exit /b 1
    )
    echo ✓ 平衡工具编译完成
) else (
    echo ✓ 平衡工具已存在
)
echo.

REM 检查 demo 文件夹
echo [3/4] 检查 demo 文件...
if exist "demo" (
    dir /b /a-d demo\*.dem >nul 2>&1
    if not errorlevel 1 (
        echo ✓ 找到 demo 文件
        echo.
        echo 是否解析 demo 文件夹？ (Y/N)
        set /p PARSE_DEMO=^> 
        if /i "!PARSE_DEMO!"=="Y" (
            echo.
            echo 开始批量解析...
            echo ----------------------------------------
            minidemo-encoder.exe -dir="demo"
            echo ----------------------------------------
            echo ✓ 解析完成
        ) else (
            echo 跳过解析步骤
        )
    ) else (
        echo ⚠ demo 文件夹存在但未找到 .dem 文件
        echo 请将 demo 文件放入 demo 文件夹后再试
    )
) else (
    echo ⚠ 未找到 demo 文件夹
    echo 提示: 创建 demo 文件夹并放入 .dem 文件
)
echo.

REM 询问是否平衡 rec 文件
echo [4/4] REC 文件平衡...
echo.
echo 是否需要平衡 REC 文件？ (Y/N)
set /p DO_BALANCE=^> 
if /i "!DO_BALANCE!"=="Y" (
    echo.
    echo 请输入需要平衡的路径：
    echo （直接回车使用默认路径：./output）
    set /p BALANCE_PATH=^> 
    
    if "!BALANCE_PATH!"=="" (
        set BALANCE_PATH=.\output
    )
    
    if exist "!BALANCE_PATH!" (
        echo.
        echo 开始平衡 REC 文件...
        echo ----------------------------------------
        rec_balancer.exe -path="!BALANCE_PATH!"
        echo ----------------------------------------
        echo ✓ 平衡完成
    ) else (
        echo ❌ 路径不存在: !BALANCE_PATH!
    )
) else (
    echo 跳过平衡步骤
)
echo.

echo ========================================
echo  测试完成！
echo ========================================
echo.
echo 如果一切正常，您现在可以：
echo   • 查看 output 文件夹中的输出文件
echo   • 使用 batch_parse.bat 批量解析 demo
echo   • 使用 rec_balance.bat 平衡 rec 文件
echo.
echo 详细文档：
echo   • 批量解析说明.md
echo   • REC平衡工具说明.md
echo   • 快速开始.md
echo.
pause 