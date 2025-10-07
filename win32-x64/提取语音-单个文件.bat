@echo off
chcp 65001 >nul
echo ========================================
echo CS:GO Demo 语音提取工具
echo 模式：单个完整文件（所有玩家合并）
echo ========================================
echo.

if "%~1"=="" (
    echo 使用方法：将demo文件拖拽到此批处理文件上
    echo 或者：提取语音-单个文件.bat demo文件名.dem
    echo.
    pause
    exit /b
)

echo 正在处理: %~nx1
echo.

csgove.exe -mode single-full "%~1"

echo.
if errorlevel 1 (
    echo.
    echo ========================================
    echo 出现错误！
    echo ========================================
    echo.
    echo 如果看到 "dlopen error #126"，请安装：
    echo Visual C++ Redistributable ^(32位^)
    echo 下载: https://aka.ms/vs/17/release/vc_redist.x86.exe
    echo.
    echo 详细说明请查看：使用说明.txt
    echo ========================================
) else (
    echo.
    echo ========================================
    echo 提取完成！
    echo 输出文件: %~n1.wav
    echo ========================================
)

echo.
pause 