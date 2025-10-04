#!/bin/bash

# 批量解析 CS:GO Demo 脚本
# 用法: ./batch_parse.sh [demo文件夹路径]

DEMO_DIR="${1:-./demo}"

echo "========================================"
echo "CS:GO Demo 批量解析工具"
echo "========================================"
echo "Demo 目录: $DEMO_DIR"
echo ""

if [ ! -d "$DEMO_DIR" ]; then
    echo "错误: 目录不存在: $DEMO_DIR"
    echo "用法: ./batch_parse.sh [demo文件夹路径]"
    exit 1
fi

# 编译程序（如果需要）
if [ ! -f "./minidemo-encoder" ] && [ ! -f "./minidemo-encoder.exe" ]; then
    echo "正在编译程序..."
    go build -o minidemo-encoder ./cmd/main.go
    if [ $? -ne 0 ]; then
        echo "编译失败！"
        exit 1
    fi
    echo "编译完成！"
    echo ""
fi

# 执行批量解析
if [ -f "./minidemo-encoder" ]; then
    ./minidemo-encoder -dir="$DEMO_DIR"
elif [ -f "./minidemo-encoder.exe" ]; then
    ./minidemo-encoder.exe -dir="$DEMO_DIR"
else
    echo "错误: 找不到可执行文件"
    exit 1
fi

echo ""
echo "解析完成！输出文件在 ./output 目录中" 