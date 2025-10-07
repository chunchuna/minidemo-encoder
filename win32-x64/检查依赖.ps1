#!/usr/bin/env pwsh
# CS:GO Voice Extractor - 依赖检查脚本

Write-Host "========================================"  -ForegroundColor Cyan
Write-Host "CS:GO Demo 语音提取工具 - 依赖检查" -ForegroundColor Cyan
Write-Host "========================================"  -ForegroundColor Cyan
Write-Host ""

# 检查DLL文件
Write-Host "检查必需的DLL文件..." -ForegroundColor Yellow
$requiredDlls = @("csgove.exe", "vaudio_celt.dll", "tier0.dll", "opus.dll")
$allDllsPresent = $true

foreach ($dll in $requiredDlls) {
    if (Test-Path $dll) {
        Write-Host "  ✓ $dll" -ForegroundColor Green
    } else {
        Write-Host "  ✗ $dll - 缺失！" -ForegroundColor Red
        $allDllsPresent = $false
    }
}

Write-Host ""

# 检查Visual C++ Redistributable
Write-Host "检查 Visual C++ Redistributable (32位)..." -ForegroundColor Yellow

$vcRedistKeys = @(
    "HKLM:\SOFTWARE\WOW6432Node\Microsoft\VisualStudio\14.0\VC\Runtimes\x86",
    "HKLM:\SOFTWARE\Microsoft\VisualStudio\14.0\VC\Runtimes\x86",
    "HKLM:\SOFTWARE\WOW6432Node\Microsoft\VisualStudio\12.0\VC\Runtimes\x86",
    "HKLM:\SOFTWARE\Classes\Installer\Dependencies\VC,redist.x86,x86,14.*",
    "HKLM:\SOFTWARE\Classes\Installer\Dependencies\Microsoft.VS.VC_RuntimeMinimumVSU_x86,v14"
)

$vcInstalled = $false
foreach ($key in $vcRedistKeys) {
    if (Test-Path $key) {
        $vcInstalled = $true
        Write-Host "  ✓ 已安装" -ForegroundColor Green
        break
    }
}

if (-not $vcInstalled) {
    Write-Host "  ✗ 未找到（可能未安装）" -ForegroundColor Red
    Write-Host ""
    Write-Host "  需要安装 Visual C++ Redistributable (x86)" -ForegroundColor Yellow
    Write-Host "  下载地址: https://aka.ms/vs/17/release/vc_redist.x86.exe" -ForegroundColor Cyan
    Write-Host ""
    
    $download = Read-Host "  是否打开下载页面？(Y/N)"
    if ($download -eq "Y" -or $download -eq "y") {
        Start-Process "https://aka.ms/vs/17/release/vc_redist.x86.exe"
    }
}

Write-Host ""
Write-Host "========================================"  -ForegroundColor Cyan

if ($allDllsPresent -and $vcInstalled) {
    Write-Host "✓ 所有依赖检查通过！" -ForegroundColor Green
    Write-Host ""
    Write-Host "您现在可以运行提取命令：" -ForegroundColor Green
    Write-Host "  .\csgove.exe -mode single-full demo文件名.dem" -ForegroundColor White
} else {
    Write-Host "✗ 缺少必要的依赖" -ForegroundColor Red
    Write-Host ""
    Write-Host "请根据上述提示安装缺失的组件" -ForegroundColor Yellow
}

Write-Host "========================================"  -ForegroundColor Cyan
Write-Host ""
Read-Host "按回车键退出" 