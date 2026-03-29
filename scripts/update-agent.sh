#!/bin/bash
# JumpFrp Agent 更新脚本
# 用法: sudo ./update-agent.sh

set -e

AGENT_DIR="/opt/jumpfrp"
AGENT_BIN="$AGENT_DIR/jumpfrp-agent"
REPO_URL="https://github.com/AKEXZ/JumpFrp.git"
TEMP_DIR=$(mktemp -d)

echo "🔄 JumpFrp Agent 更新"
echo "================================"

# 检查权限
if [ "$EUID" -ne 0 ]; then 
  echo "❌ 需要 root 权限，请使用 sudo 运行"
  exit 1
fi

# 检查依赖
echo "📦 检查依赖..."
if ! command -v git &> /dev/null; then
  echo "❌ 未找到 git，请先安装"
  exit 1
fi

if ! command -v go &> /dev/null; then
  echo "❌ 未找到 Go，请先安装 Go 1.24+"
  exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "✅ Go 版本: $GO_VERSION"

# 克隆/更新仓库
echo "📥 克隆最新代码..."
if [ -d "$TEMP_DIR/JumpFrp" ]; then
  cd "$TEMP_DIR/JumpFrp"
  git pull origin main
else
  cd "$TEMP_DIR"
  git clone --depth 1 "$REPO_URL"
  cd JumpFrp
fi

# 编译 Agent
echo "🔨 编译 Agent..."
cd agent
go build -o jumpfrp-agent ./cmd/main.go

if [ ! -f jumpfrp-agent ]; then
  echo "❌ 编译失败"
  exit 1
fi

echo "✅ 编译成功"

# 停止现有 Agent
echo "⏹️  停止现有 Agent..."
if systemctl is-active --quiet jumpfrp-agent; then
  systemctl stop jumpfrp-agent
  sleep 2
fi

# 备份旧二进制
if [ -f "$AGENT_BIN" ]; then
  cp "$AGENT_BIN" "$AGENT_BIN.bak"
  echo "💾 已备份旧版本到 $AGENT_BIN.bak"
fi

# 复制新二进制
cp jumpfrp-agent "$AGENT_BIN"
chmod +x "$AGENT_BIN"
echo "✅ 已安装新版本"

# 启动 Agent
echo "🚀 启动 Agent..."
systemctl start jumpfrp-agent
sleep 2

if systemctl is-active --quiet jumpfrp-agent; then
  echo "✅ Agent 已启动"
  systemctl status jumpfrp-agent --no-pager
else
  echo "❌ Agent 启动失败，恢复旧版本..."
  cp "$AGENT_BIN.bak" "$AGENT_BIN"
  systemctl start jumpfrp-agent
  exit 1
fi

# 清理
rm -rf "$TEMP_DIR"

echo ""
echo "================================"
echo "✅ Agent 更新完成！"
echo ""
echo "📝 日志位置: /var/log/jumpfrp-agent.log"
echo "⚙️  配置位置: /etc/jumpfrp/agent.env"
echo ""
echo "💡 提示: Agent 会在 2 秒内从主控获取最新 frps 配置"
