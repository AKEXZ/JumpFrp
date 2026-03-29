#!/bin/bash
# JumpFrp 节点一键安装脚本（自动注册版本）
# 用法: bash <(wget -qO- https://raw.githubusercontent.com/AKEXZ/JumpFrp/main/scripts/install.sh)

set -e

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 默认参数
MASTER_URL="https://api.jumpfrp.top"
FRPS_PORT=7000
AGENT_PORT=7500
USE_PROXY=false
NODE_ID=""
TOKEN=""

echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   JumpFrp 节点一键安装脚本 v1.1.0     ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"
echo ""

# 检查 root
if [[ $EUID -ne 0 ]]; then
  echo -e "${RED}❌ 错误: 请使用 root 用户运行此脚本${NC}"
  exit 1
fi

# 检查系统
if [[ ! -f /etc/os-release ]]; then
  echo -e "${RED}❌ 错误: 无法识别操作系统${NC}"
  exit 1
fi

source /etc/os-release
if [[ "$ID" != "ubuntu" && "$ID" != "debian" ]]; then
  echo -e "${YELLOW}⚠️  警告: 当前系统 $ID 可能不受支持，建议 Ubuntu 20.04+${NC}"
  read -p "是否继续? [y/N] " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 1
  fi
fi

# 获取节点信息
echo -e "${YELLOW}📋 请输入节点信息${NC}"
echo ""

read -p "节点名称 (例: sh-01): " NODE_NAME
if [[ -z "$NODE_NAME" ]]; then
  echo -e "${RED}❌ 节点名称不能为空${NC}"
  exit 1
fi

read -p "节点 IP (例: 1.2.3.4): " NODE_IP
if [[ -z "$NODE_IP" ]]; then
  echo -e "${RED}❌ 节点 IP 不能为空${NC}"
  exit 1
fi

read -p "节点地区 (例: 新加坡): " NODE_REGION
if [[ -z "$NODE_REGION" ]]; then
  NODE_REGION="未知"
fi

echo ""
echo -e "${YELLOW}🔗 正在向主控注册节点...${NC}"

# 调用主控 API 自动注册节点
REGISTER_RESPONSE=$(curl -s -X POST \
  -H "Content-Type: application/json" \
  "$MASTER_URL/api/admin/node/auto-register" \
  -d "{\"name\":\"$NODE_NAME\",\"ip\":\"$NODE_IP\",\"region\":\"$NODE_REGION\"}")

# 解析响应
CODE=$(echo "$REGISTER_RESPONSE" | grep -o '"code":[0-9]*' | grep -o '[0-9]*')
if [[ "$CODE" != "0" ]]; then
  echo -e "${RED}❌ 节点注册失败${NC}"
  echo "$REGISTER_RESPONSE"
  exit 1
fi

NODE_ID=$(echo "$REGISTER_RESPONSE" | grep -o '"node_id":"[^"]*"' | cut -d'"' -f4)
TOKEN=$(echo "$REGISTER_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [[ -z "$NODE_ID" || -z "$TOKEN" ]]; then
  echo -e "${RED}❌ 无法解析节点 ID 和 Token${NC}"
  exit 1
fi

echo -e "${GREEN}✅ 节点注册成功${NC}"
echo ""
echo -e "${BLUE}节点信息:${NC}"
echo "  节点 ID: $NODE_ID"
echo "  Token: ${TOKEN:0:16}...${TOKEN: -8}"
echo ""

# 继续安装
echo -e "${YELLOW}📦 安装步骤:${NC}"
echo "  [1/6] 检查系统依赖"
echo "  [2/6] 下载 frps 服务端"
echo "  [3/6] 下载 JumpFrp Agent"
echo "  [4/6] 创建配置文件"
echo "  [5/6] 创建系统服务"
echo "  [6/6] 启动服务并注册"
echo ""

# 检查必要依赖
echo -e "${GREEN}[1/6] 检查系统依赖...${NC}"
DEPS=""
for cmd in tc iptables ip wget curl tar gunzip; do
  if ! command -v $cmd &> /dev/null; then
    DEPS="$DEPS $cmd"
  fi
done
if [[ -n "$DEPS" ]]; then
  echo -e "${YELLOW}缺少依赖: $DEPS${NC}"
  echo -e "${YELLOW}正在安装...${NC}"
  apt-get update && apt-get install -y iproute2 iptables wget curl tar gzip 2>/dev/null || \
  yum install -y iproute iptables wget curl tar gzip 2>/dev/null || \
  echo -e "${RED}自动安装依赖失败，请手动安装: $DEPS${NC}"
fi
echo -e "${GREEN}✓ 依赖检查完成${NC}"

# 安装目录
INSTALL_DIR="/opt/jumpfrp"
mkdir -p $INSTALL_DIR

# 下载 frps
echo -e "${GREEN}[2/6] 下载 frps...${NC}"
FRPS_VERSION="0.61.0"
ARCH=$(uname -m)
if [[ "$ARCH" == "x86_64" ]]; then
  FRPS_ARCH="amd64"
elif [[ "$ARCH" == "aarch64" ]]; then
  FRPS_ARCH="arm64"
else
  echo -e "${RED}❌ 不支持的架构: $ARCH${NC}"
  exit 1
fi

FRPS_URL="https://github.com/fatedier/frp/releases/download/v${FRPS_VERSION}/frp_${FRPS_VERSION}_linux_${FRPS_ARCH}.tar.gz"

if [[ "$USE_PROXY" == "true" ]]; then
  FRPS_URL="https://gitproxy.ake.cx/${FRPS_URL}"
fi

if wget --progress=bar:force --timeout=60 -O /tmp/frp.tar.gz "$FRPS_URL" 2>&1; then
  echo -e "${GREEN}✓ frps 下载完成${NC}"
else
  if [[ "$USE_PROXY" != "true" ]]; then
    echo -e "${YELLOW}直连失败，尝试使用代理...${NC}"
    PROXY_URL="https://gitproxy.ake.cx/https://github.com/fatedier/frp/releases/download/v${FRPS_VERSION}/frp_${FRPS_VERSION}_linux_${FRPS_ARCH}.tar.gz"
    if wget --progress=bar:force --timeout=60 -O /tmp/frp.tar.gz "$PROXY_URL" 2>&1; then
      echo -e "${GREEN}✓ frps 下载完成（通过代理）${NC}"
    else
      echo -e "${RED}❌ 下载 frps 失败${NC}"
      exit 1
    fi
  else
    echo -e "${RED}❌ 下载 frps 失败${NC}"
    exit 1
  fi
fi

tar -xzf /tmp/frp.tar.gz -C /tmp/
cp /tmp/frp_*/frps $INSTALL_DIR/frps
chmod +x $INSTALL_DIR/frps
rm -rf /tmp/frp_*

# 下载 Agent
echo -e "${GREEN}[3/6] 下载 JumpFrp Agent...${NC}"
AGENT_VERSION="v1.1.0"
AGENT_URL="https://github.com/AKEXZ/JumpFrp/releases/download/${AGENT_VERSION}/jumpfrp-agent"

if [[ "$USE_PROXY" == "true" ]]; then
  AGENT_URL="https://gitproxy.ake.cx/${AGENT_URL}"
fi

if wget --progress=bar:force --timeout=60 -O $INSTALL_DIR/jumpfrp-agent "$AGENT_URL" 2>&1; then
  echo -e "${GREEN}✓ Agent 下载完成${NC}"
else
  if [[ "$USE_PROXY" != "true" ]]; then
    echo -e "${YELLOW}直连失败，尝试使用代理...${NC}"
    PROXY_URL="https://gitproxy.ake.cx/https://github.com/AKEXZ/JumpFrp/releases/download/${AGENT_VERSION}/jumpfrp-agent"
    if wget --progress=bar:force --timeout=60 -O $INSTALL_DIR/jumpfrp-agent "$PROXY_URL" 2>&1; then
      echo -e "${GREEN}✓ Agent 下载完成（通过代理）${NC}"
    else
      echo -e "${RED}❌ 下载 Agent 失败${NC}"
      exit 1
    fi
  else
    echo -e "${RED}❌ 下载 Agent 失败${NC}"
    exit 1
  fi
fi

chmod +x $INSTALL_DIR/jumpfrp-agent

if file $INSTALL_DIR/jumpfrp-agent | grep -q "ELF"; then
  echo -e "${GREEN}✓ Agent 已安装 (v${AGENT_VERSION})${NC}"
else
  echo -e "${RED}❌ Agent 文件格式不正确${NC}"
  rm -f $INSTALL_DIR/jumpfrp-agent
  exit 1
fi

# 创建配置文件
echo -e "${GREEN}[4/6] 创建配置文件...${NC}"
cat > $INSTALL_DIR/agent.env << EOF
AGENT_NODE_ID=${NODE_ID}
AGENT_TOKEN=${TOKEN}
AGENT_MASTER_URL=${MASTER_URL}
AGENT_PORT=${AGENT_PORT}
FRPS_PORT=${FRPS_PORT}
EOF

cat > $INSTALL_DIR/frps.toml << EOF
# frps.toml - JumpFrp 服务端配置
# 此文件由 Agent 自动从主控获取，版本可能不同
bindPort = ${FRPS_PORT}
auth.method = "token"

[transport]
max_pool_count = 100
pool_count = 10

[log]
to = "/var/log/frps.log"
level = "info"
max_days = 3
EOF

echo -e "${GREEN}✓ 配置文件已创建${NC}"

# 创建 systemd 服务
echo -e "${GREEN}[5/6] 创建系统服务...${NC}"

cat > /etc/systemd/system/jumpfrp-agent.service << EOF
[Unit]
Description=JumpFrp Agent
After=network.target

[Service]
Type=simple
EnvironmentFile=/opt/jumpfrp/agent.env
ExecStart=/opt/jumpfrp/jumpfrp-agent
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# 启动服务
echo -e "${GREEN}[6/6] 启动服务...${NC}"
systemctl daemon-reload
systemctl enable jumpfrp-agent 2>/dev/null

if systemctl start jumpfrp-agent; then
  sleep 2
  if systemctl is-active --quiet jumpfrp-agent; then
    echo -e "${GREEN}✓ Agent 服务运行正常${NC}"
  else
    echo -e "${RED}✗ Agent 服务启动失败${NC}"
    echo "  查看错误: journalctl -u jumpfrp-agent -n 20 --no-pager"
  fi
else
  echo -e "${RED}✗ Agent 启动命令执行失败${NC}"
fi

echo ""
echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║        ✅ 安装完成                     ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════╝${NC}"
echo ""
echo -e "${BLUE}📋 安装信息:${NC}"
echo "  ├─ 节点名称: $NODE_NAME"
echo "  ├─ 节点 ID: $NODE_ID"
echo "  ├─ 安装目录: $INSTALL_DIR"
echo "  ├─ frps 端口: $FRPS_PORT"
echo "  └─ Agent 端口: $AGENT_PORT"
echo ""
echo -e "${BLUE}🔧 常用命令:${NC}"
echo "  ├─ 查看状态: systemctl status jumpfrp-agent"
echo "  ├─ 查看日志: journalctl -u jumpfrp-agent -f"
echo "  └─ 重启服务: systemctl restart jumpfrp-agent"
echo ""
echo -e "${YELLOW}💡 提示:${NC}"
echo "  1. 请前往管理后台确认节点状态已变为「在线」"
echo "  2. 如果显示离线，请检查防火墙是否开放端口 $FRPS_PORT 和 $AGENT_PORT"
echo "  3. 运行 'journalctl -u jumpfrp-agent -n 50' 查看详细错误"
echo ""
