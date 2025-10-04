#!/bin/bash

# REC 文件平衡工具
# 确保每个地图的每个t/ct文件夹都有5个不同的rec文件

if [ -z "$1" ]; then
    BASE_PATH="./output"
else
    BASE_PATH="$1"
fi

echo "========================================"
echo "REC 文件平衡工具"
echo "========================================"
echo "目标路径: $BASE_PATH"
echo ""

if [ ! -d "$BASE_PATH" ]; then
    echo "错误: 路径不存在: $BASE_PATH"
    echo ""
    echo "用法: ./rec_balance.sh [基础路径]"
    echo "示例: ./rec_balance.sh /path/to/botmimic/demotest"
    exit 1
fi

# 切换到脚本所在目录
cd "$(dirname "$0")"

# 检查是否需要编译
if [ ! -f "./rec_balancer" ] && [ ! -f "./rec_balancer.exe" ]; then
    echo "正在编译程序..."
    go build -o rec_balancer ./cmd/rec_balancer/main.go
    if [ $? -ne 0 ]; then
        echo "编译失败！"
        exit 1
    fi
    echo "编译完成！"
    echo ""
fi

# 执行平衡操作
if [ -f "./rec_balancer" ]; then
    ./rec_balancer -path="$BASE_PATH"
elif [ -f "./rec_balancer.exe" ]; then
    ./rec_balancer.exe -path="$BASE_PATH"
else
    echo "错误: 找不到可执行文件"
    exit 1
fi

echo ""
echo "操作完成！" 