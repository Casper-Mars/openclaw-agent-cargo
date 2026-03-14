#!/bin/bash

# OpenClaw Agent Cargo 使用示例

set -e

BINARY="./openclaw-agent-cargo"

echo "=========================================="
echo "OpenClaw Agent Cargo 使用示例"
echo "=========================================="
echo ""

# 检查 binary 是否存在
if [ ! -f "$BINARY" ]; then
    echo "❌ 未找到 $BINARY，请先运行 'make build'"
    exit 1
fi

echo "✅ 找到 $BINARY"
echo ""

# 示例 1: 导出 Agent
echo "------------------------------------------"
echo "示例 1: 导出 Agent"
echo "------------------------------------------"
echo "命令：$BINARY export --agent main --output ./main-backup.tar.gz --verbose"
echo ""
# $BINARY export --agent main --output ./main-backup.tar.gz --verbose
echo "（跳过实际执行，避免修改文件系统）"
echo ""

# 示例 2: 导入 Agent
echo "------------------------------------------"
echo "示例 2: 导入 Agent"
echo "------------------------------------------"
echo "命令：$BINARY import --file ./main-backup.tar.gz --force"
echo ""
# $BINARY import --file ./main-backup.tar.gz --force
echo "（跳过实际执行，需要真实的备份文件）"
echo ""

# 示例 3: 导入并重命名
echo "------------------------------------------"
echo "示例 3: 导入并重命名 Agent"
echo "------------------------------------------"
echo "命令：$BINARY import --file ./main-backup.tar.gz --rename main-restored"
echo ""
# $BINARY import --file ./main-backup.tar.gz --rename main-restored
echo "（跳过实际执行，需要真实的备份文件）"
echo ""

echo "=========================================="
echo "完整帮助信息"
echo "=========================================="
$BINARY --help

echo ""
echo "=========================================="
echo "完成!"
echo "=========================================="
