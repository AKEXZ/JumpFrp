#!/bin/bash
# JumpFrp 节点一键安装脚本
# 用法: bash <(wget -qO- https://api.jumpfrp.top/install.sh) --node-id xxx --token xxx

set -e

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 默认参数
NODE_ID=""
TOKEN=""
MASTER_URL="https://api.jumpfrp.top"
FRPS_PORT=7000
AGENT_PORT=7500
USE_PROXY=false

# 解析参数
while [[ $# -gt 0 ]]; do
  case $1 in
    --node-id) NODE_ID="$2"; shift 2 ;;
    --token) TOKEN="$2"; shift 2 ;;
    --master-url) MASTER_URL="$2"; shift 2 ;;
    --frps-port) FRPS_PORT="$2"; shift 2 ;;
    --agent-port) AGENT_PORT="$2"; shift 2 ;;
    --proxy) USE_PROXY="$2"; shift 2 ;;
    --use-proxy) USE_PROXY=true; shift ;;
    --no-proxy) USE_PROXY=false; shift ;;
    *) echo "未知参数: $1"; exit 1 ;;
  esac
done

# 检查必需参数
if [[ -z "$NODE_ID" || -z "$TOKEN" ]]; then
  echo -e "${RED}错误: 必须指定 --node-id 和 --token${NC}"
  echo "用法: bash install.sh --node-id <节点标识> --token <Token>"
  echo ""
  echo "可选参数:"
  echo "  --master-url     主控地址 (默认: https://api.jumpfrp.top)"
  echo "  --frps-port      frps 端口 (默认: 7000)"
  echo "  --agent-port     Agent 端口 (默认: 7500)"
  echo "  --proxy          是否使用代理下载 (true/false，默认: false)"
  echo "  --use-proxy      等同于 --proxy true"
  echo "  --no-proxy       等同于 --proxy false"
  echo ""
  echo "示例:"
  echo "  bash install.sh --node-id sh-01 --token xxx"
  echo "  bash install.sh --node-id sh-01 --token xxx --proxy true"
  exit 1
fi

echo -e "${GREEN}=== JumpFrp 节点安装程序 ===${NC}"
echo "节点标识: $NODE_ID"
echo "主控地址: $MASTER_URL"
echo "frps 端口: $FRPS_PORT"
echo "Agent 端口: $AGENT_PORT"
echo "使用代理: $USE_PROXY"
echo ""
echo -e "${YELLOW}安装步骤:${NC}"
echo "  [1/6] 检查系统依赖"
echo "  [2/6] 下载 frps 服务端"
echo "  [3/6] 下载 JumpFrp Agent"
echo "  [4/6] 创建配置文件"
echo "  [5/6] 创建系统服务"
echo "  [6/6] 启动服务并注册"
echo ""

# 检查系统
if [[ ! -f /etc/os-release ]]; then
  echo -e "${RED}错误: 无法识别操作系统${NC}"
  exit 1
fi

source /etc/os-release
if [[ "$ID" != "ubuntu" && "$ID" != "debian" ]]; then
  echo -e "${YELLOW}警告: 当前系统 $ID 可能不受支持，建议 Ubuntu 20.04+${NC}"
  read -p "是否继续? [y/N] " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 1
  fi
fi

# 检查 root
if [[ $EUID -ne 0 ]]; then
  echo -e "${RED}错误: 请使用 root 用户运行此脚本${NC}"
  exit 1
fi

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
echo -e "${GREEN}依赖检查完成${NC}"

# 安装目录
INSTALL_DIR="/opt/jumpfrp"
mkdir -p $INSTALL_DIR

# 下载 frps（从 GitHub 或国内镜像）
echo -e "${GREEN}[2/6] 下载 frps...${NC}"
FRPS_VERSION="0.61.0"
ARCH=$(uname -m)
if [[ "$ARCH" == "x86_64" ]]; then
  FRPS_ARCH="amd64"
elif [[ "$ARCH" == "aarch64" ]]; then
  FRPS_ARCH="arm64"
else
  echo -e "${RED}不支持的架构: $ARCH${NC}"
  exit 1
fi

FRPS_URL="https://github.com/fatedier/frp/releases/download/v${FRPS_VERSION}/frp_${FRPS_VERSION}_linux_${FRPS_ARCH}.tar.gz"

# 根据是否使用代理选择下载地址
if [[ "$USE_PROXY" == "true" ]]; then
  FRPS_URL="https://gitproxy.ake.cx/${FRPS_URL}"
  echo "使用代理下载，下载地址: $FRPS_URL"
else
  echo "直连 GitHub 下载，下载地址: $FRPS_URL"
fi

echo -e "${YELLOW}正在连接...${NC}"
if wget --progress=bar:force --timeout=60 -O /tmp/frp.tar.gz "$FRPS_URL" 2>&1; then
  echo -e "${GREEN}frps 下载完成${NC}"
else
  # 如果直连失败，且未使用代理，尝试使用代理
  if [[ "$USE_PROXY" != "true" ]]; then
    echo -e "${YELLOW}直连失败，尝试使用代理下载...${NC}"
    PROXY_URL="https://gitproxy.ake.cx/${FRPS_URL}"
    if wget --progress=bar:force --timeout=60 -O /tmp/frp.tar.gz "$PROXY_URL" 2>&1; then
      echo -e "${GREEN}frps 下载完成（通过代理）${NC}"
    else
      echo -e "${RED}下载 frps 失败${NC}"
      exit 1
    fi
  else
    echo -e "${RED}下载 frps 失败，请检查网络连接${NC}"
    exit 1
  fi
fi

tar -xzf /tmp/frp.tar.gz -C /tmp/
cp /tmp/frp_*/frps $INSTALL_DIR/frps
chmod +x $INSTALL_DIR/frps
rm -rf /tmp/frp_*
echo -e "${GREEN}frps 已安装 (v${FRPS_VERSION})${NC}"

# 下载 Agent（从 GitHub Releases 下载）
echo -e "${GREEN}[3/6] 下载 JumpFrp Agent...${NC}"
AGENT_VERSION="v1.1.0"
AGENT_URL="https://github.com/AKEXZ/JumpFrp/releases/download/${AGENT_VERSION}/jumpfrp-agent"

# 根据是否使用代理选择下载地址
if [[ "$USE_PROXY" == "true" ]]; then
  AGENT_URL="https://gitproxy.ake.cx/${AGENT_URL}"
  echo "使用代理下载，下载地址: $AGENT_URL"
else
  echo "直连 GitHub 下载，下载地址: $AGENT_URL"
fi

echo -e "${YELLOW}正在连接...${NC}"
if wget --progress=bar:force --timeout=60 -O $INSTALL_DIR/jumpfrp-agent "$AGENT_URL" 2>&1; then
  echo -e "${GREEN}Agent 下载完成${NC}"
else
  # 如果直连失败，且未使用代理，尝试使用代理
  if [[ "$USE_PROXY" != "true" ]]; then
    echo -e "${YELLOW}直连失败，尝试使用代理下载...${NC}"
    PROXY_URL="https://gitproxy.ake.cx/https://github.com/AKEXZ/JumpFrp/releases/download/${AGENT_VERSION}/jumpfrp-agent"
    if wget --progress=bar:force --timeout=60 -O $INSTALL_DIR/jumpfrp-agent "$PROXY_URL" 2>&1; then
      echo -e "${GREEN}Agent 下载完成（通过代理）${NC}"
    else
      echo -e "${RED}下载 Agent 失败${NC}"
      exit 1
    fi
  else
    echo -e "${RED}下载 Agent 失败${NC}"
    exit 1
  fi
fi

chmod +x $INSTALL_DIR/jumpfrp-agent

# 验证是否为有效的 ELF 可执行文件
if file $INSTALL_DIR/jumpfrp-agent | grep -q "ELF"; then
  echo -e "${GREEN}Agent 已安装 (v${AGENT_VERSION})${NC}"
else
  echo -e "${RED}Agent 文件格式不正确${NC}"
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

# 创建基础 frps.toml 配置（首次安装用，Agent 注册后会从主控获取完整配置）
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

echo -e "${GREEN}配置文件已创建${NC}"

# 创建 systemd 服务（Agent 会自动管理 frps，无需单独服务）
echo -e "${GREEN}[5/6] 创建系统服务...${NC}"

# Agent 服务（Agent 会 fork frps 进程）
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

echo -e "${YELLOW}正在启动 Agent...${NC}"
if systemctl start jumpfrp-agent; then
  sleep 2
  if systemctl is-active --quiet jumpfrp-agent; then
    echo -e "${GREEN}✓ Agent 服务运行正常${NC}"
    echo -e "${YELLOW}  正在向主控注册节点，获取 frps 配置...${NC}"
    sleep 5
  else
    echo -e "${RED}✗ Agent 服务启动失败${NC}"
    echo "  查看错误: journalctl -u jumpfrp-agent -n 20 --no-pager"
  fi
else
  echo -e "${RED}✗ Agent 启动命令执行失败${NC}"
fi

echo ""
echo -e "${GREEN}══════════════════════════════════════${NC}"
echo -e "${GREEN}        安装完成${NC}"
echo -e "${GREEN}══════════════════════════════════════${NC}"
echo ""
echo "安装信息:"
echo "  ├─ 安装目录: $INSTALL_DIR"
echo "  ├─ frps 端口: $FRPS_PORT"
echo "  ├─ Agent 端口: $AGENT_PORT"
echo "  └─ 节点标识: $NODE_ID"
echo ""
echo "常用命令:"
echo "  ├─ 查看状态: systemctl status jumpfrp-agent"
echo "  ├─ 查看日志: journalctl -u jumpfrp-agent -f"
echo "  └─ 重启服务: systemctl restart jumpfrp-agent"
echo ""
echo -e "${YELLOW}提示: 请前往管理后台确认节点状态已变为「在线」${NC}"
echo -e "${YELLOW}如果显示离线，请检查: ${NC}"
echo "  1. 服务器防火墙是否开放端口 $FRPS_PORT 和 $AGENT_PORT"
echo "  2. 运行 'journalctl -u jumpfrp-agent -n 50' 查看详细错误"
